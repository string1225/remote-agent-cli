[CmdletBinding()]
param(
    [string]$Version = '',
    [uint64]$Sequence = 0,
    [string]$PreparedPath = '',
    [switch]$PrepareOnly,
    [string]$SSHHost = 'deploy-139',
    [string]$SSHConfig = 'C:/Users/junte/.remote-agent/git-auth/ssh-139/config'
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$repository = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$safeRepository = $repository.Replace('\','/')
$workRoot = Join-Path $repository '.local/runtime-releases'
$utf8 = [Text.UTF8Encoding]::new($false)
function Invoke-Checked([string]$Command, [string[]]$Arguments) {
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command failed with exit code $LASTEXITCODE" }
}
if ($SSHHost -notmatch '^[A-Za-z0-9][A-Za-z0-9_.-]*(@[A-Za-z0-9][A-Za-z0-9_.-]*)?$') { throw 'Invalid SSH host.' }
if (!$PreparedPath) {
    if ($Version -notmatch '^[A-Za-z0-9._-]{1,80}$' -or $Sequence -eq 0) { throw 'Specify -Version and a new monotonic -Sequence.' }
    $revision = (& git -c "safe.directory=$safeRepository" -C $repository rev-parse HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or $revision -notmatch '^[0-9a-f]{40}$') { throw 'Cannot determine committed revision.' }
    $label = $revision + '-' + (Get-Date -Format yyyyMMddHHmmss) + '-' + [Guid]::NewGuid().ToString('N').Substring(0,8)
    $PreparedPath = Join-Path $workRoot $label
    $source = Join-Path $PreparedPath 'source'
    New-Item -ItemType Directory -Path $source -Force | Out-Null
    Invoke-Checked git @('-c',"safe.directory=$safeRepository",'-C',$repository,'archive','--format=tar',"--output=$PreparedPath/source.tar",$revision)
    Invoke-Checked tar @('-xf',"$PreparedPath/source.tar",'-C',$source)
    $savedOS, $savedArch, $savedCGO, $savedFlags = $env:GOOS, $env:GOARCH, $env:CGO_ENABLED, $env:GOFLAGS
    try {
        $env:GOOS = 'windows'
        $env:GOARCH = 'amd64'
        $env:CGO_ENABLED = '0'
        $env:GOFLAGS = '-buildvcs=false'
        Invoke-Checked go @('-C',$source,'run','./tools/build','-version',$Version)
        Invoke-Checked go @('-C',$source,'build','-trimpath','-o',"$PreparedPath/release-sign.exe",'./tools/release-sign')
    } finally {
        $env:GOOS, $env:GOARCH, $env:CGO_ENABLED, $env:GOFLAGS = $savedOS, $savedArch, $savedCGO, $savedFlags
    }
    $metadata = [ordered]@{schemaVersion=1;revision=$revision;version=$Version;sequence=$Sequence}
    [IO.File]::WriteAllText((Join-Path $PreparedPath 'release.json'), ($metadata | ConvertTo-Json), $utf8)
    Copy-Item -LiteralPath (Join-Path $source 'deploy/publish-runtime.sh') -Destination $PreparedPath
    $signScript = @'
$ErrorActionPreference = 'Stop'
$release = Get-Content -LiteralPath (Join-Path $PSScriptRoot 'release.json') -Raw | ConvertFrom-Json
$arguments = @('-dist',(Join-Path $PSScriptRoot 'source/dist'),'-version',$release.version,'-sequence',[string]$release.sequence)
& (Join-Path $PSScriptRoot 'release-sign.exe') @arguments -key-dir (Join-Path $env:USERPROFILE '.remote-agent-release')
if ($LASTEXITCODE -ne 0) { throw 'Signing failed. No server publication was attempted.' }
& (Join-Path $PSScriptRoot 'release-sign.exe') @arguments -verify-only
if ($LASTEXITCODE -ne 0) { throw 'Signed package verification failed.' }
Write-Output 'All six runtime packages signed with the original key and verified. Private keys stayed in their original directory.'
'@
    [IO.File]::WriteAllText((Join-Path $PreparedPath 'sign.ps1'), $signScript, $utf8)
    Write-Output "Prepared committed revision $revision in $PreparedPath"
    Write-Output "Sign as the original desktop user: powershell -NoProfile -File '$PreparedPath/sign.ps1'"
    if ($PrepareOnly) { return }
}
$PreparedPath = [IO.Path]::GetFullPath($PreparedPath)
if (!$PreparedPath.StartsWith($workRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'Prepared release must be under this repository .local/runtime-releases.' }
$release = Get-Content -LiteralPath (Join-Path $PreparedPath 'release.json') -Raw | ConvertFrom-Json
if ($release.schemaVersion -ne 1 -or $release.revision -notmatch '^[0-9a-f]{40}$' -or $release.version -notmatch '^[A-Za-z0-9._-]{1,80}$' -or [uint64]$release.sequence -eq 0) { throw 'Invalid prepared release metadata.' }
$distribution = Join-Path $PreparedPath 'source/dist'
Invoke-Checked (Join-Path $PreparedPath 'release-sign.exe') @('-verify-only','-dist',$distribution,'-version',$release.version,'-sequence',[string]$release.sequence)
if ($PrepareOnly) { return }
$payload = Join-Path $PreparedPath 'payload'
New-Item -ItemType Directory -Path $payload -Force | Out-Null
$names = @()
foreach ($platform in @('windows','darwin','linux')) {
    foreach ($arch in @('amd64','arm64')) {
        $artifact = "remote-agent-$platform-$arch"
        if ($platform -eq 'windows') { $artifact = "remote-agent-core-windows-$arch.exe" }
        $names += @($artifact,"updates/v1/$platform/$arch.json")
    }
}
$checks = @()
foreach ($name in $names) {
    $destination = Join-Path $payload $name
    New-Item -ItemType Directory -Path ([IO.Path]::GetDirectoryName($destination)) -Force | Out-Null
    Copy-Item -LiteralPath (Join-Path $distribution $name) -Destination $destination
    $digest = (Get-FileHash -LiteralPath $destination -Algorithm SHA256).Hash.ToLowerInvariant()
    $checks += "$digest  $name"
    if ($name -notlike 'updates/*') {
        [IO.File]::WriteAllText(($destination + '.sha256'), "$digest  $name`n", $utf8)
        $checks += ((Get-FileHash -LiteralPath ($destination + '.sha256') -Algorithm SHA256).Hash.ToLowerInvariant() + "  $name.sha256")
    }
}
$provenance = Join-Path $payload 'RUNTIME_RELEASE.json'
[IO.File]::WriteAllText($provenance, ($release | ConvertTo-Json), $utf8)
$checks += ((Get-FileHash -LiteralPath $provenance -Algorithm SHA256).Hash.ToLowerInvariant() + '  RUNTIME_RELEASE.json')
[IO.File]::WriteAllText((Join-Path $payload 'files.sha256'), ($checks -join "`n") + "`n", $utf8)
$bundle = Join-Path $PreparedPath 'runtime.tar.gz'
Invoke-Checked tar @('-czf',$bundle,'-C',$payload,'.')
$label = 'runtime-' + $release.revision + '-' + [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$stage = "/tmp/remote-agent-$label"
$sshOptions = @('-F',[IO.Path]::GetFullPath($SSHConfig),'-o','BatchMode=yes','-o','ConnectTimeout=15')
Invoke-Checked ssh ($sshOptions + @($SSHHost,"umask 077; mkdir '$stage'"))
# Use the current deployment procedure even for artifacts prepared earlier.
Invoke-Checked scp ($sshOptions + @($bundle,(Join-Path $repository 'deploy/publish-runtime.sh'),"${SSHHost}:$stage/"))
$bundleHash = (Get-FileHash -LiteralPath $bundle -Algorithm SHA256).Hash.ToLowerInvariant()
Invoke-Checked ssh ($sshOptions + @($SSHHost,"sudo -n bash '$stage/publish-runtime.sh' '$label' '$bundleHash'"))
Write-Output "Published runtime $($release.version), sequence $($release.sequence), source $($release.revision)."
