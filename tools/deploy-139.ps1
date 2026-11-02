[CmdletBinding()]
param(
    [string]$SSHHost = 'junte@139.224.12.141',
    [string]$IdentityFile = '',
    [switch]$BuildOnly,
    [string]$UploadFile = '',
    [string]$ConfigPath = ''
)
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$safeRoot = $root.Replace('\','/')
function Invoke-Checked([string]$Command, [string[]]$Arguments) {
    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$Command exited with code $LASTEXITCODE" }
}
if ($SSHHost -notmatch '^[A-Za-z0-9_.-]+@[A-Za-z0-9_.-]+$') { throw 'SSHHost must be user@host.' }
$revision = (& git -c "safe.directory=$safeRoot" -C $root rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $revision -notmatch '^[0-9a-f]{40}$') { throw 'Cannot read committed revision.' }
$label = "$revision-$(Get-Date -Format yyyyMMddHHmmss)"
$work = Join-Path $root ".local/deploy-139/$label"
$source = Join-Path $work 'source'
$bundle = Join-Path $work 'bundle'
New-Item -ItemType Directory -Path $source,$bundle -Force | Out-Null
Invoke-Checked git @('-c',"safe.directory=$safeRoot",'-C',$root,'archive','--format=tar',"--output=$work/source.tar",$revision)
Invoke-Checked tar @('-xf',"$work/source.tar",'-C',$source)
$oldOS, $oldArch, $oldCGO, $oldFlags = $env:GOOS, $env:GOARCH, $env:CGO_ENABLED, $env:GOFLAGS
Push-Location $source
try {
    $env:GOFLAGS = '-buildvcs=false'
    $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; $env:CGO_ENABLED = '0'
    Invoke-Checked go @('build','-trimpath',"-ldflags=-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version=$revision",'-o',"$bundle/remote-server",'./cmd/server')
    $env:GOOS = 'windows'; $env:GOARCH = 'amd64'
    Invoke-Checked go @('build','-trimpath',"-ldflags=-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version=$revision",'-o',"$work/remote-agent-upload.exe",'./cmd/agent')
} finally {
    Pop-Location
    $env:GOOS, $env:GOARCH, $env:CGO_ENABLED, $env:GOFLAGS = $oldOS, $oldArch, $oldCGO, $oldFlags
}
$utf8 = [Text.UTF8Encoding]::new($false)
[IO.File]::WriteAllText("$bundle/SOURCE_REVISION", "$revision`n", $utf8)
$serverHash = (Get-FileHash -LiteralPath "$bundle/remote-server" -Algorithm SHA256).Hash.ToLowerInvariant()
[IO.File]::WriteAllText("$bundle/remote-server.sha256", "$serverHash  remote-server`n", $utf8)
Copy-Item -LiteralPath "$source/deploy/shared-files-location.conf" -Destination $bundle
Copy-Item -LiteralPath "$source/deploy/publish-server.sh" -Destination "$work/publish-server.sh"
Invoke-Checked tar @('-czf',"$work/server.tar.gz",'-C',$bundle,'.')
Write-Output "Prepared committed revision $revision in $work"
if ($BuildOnly) { return }
$sshOptions = @('-o','BatchMode=yes','-o','ConnectTimeout=15')
if ($IdentityFile) { $sshOptions += @('-i',[IO.Path]::GetFullPath($IdentityFile)) }
$stage = "/tmp/remote-agent-$label"
Invoke-Checked ssh ($sshOptions + @($SSHHost,"umask 077; mkdir '$stage'"))
Invoke-Checked scp ($sshOptions + @("$work/server.tar.gz","$work/publish-server.sh","${SSHHost}:$stage/"))
$bundleHash = (Get-FileHash -LiteralPath "$work/server.tar.gz" -Algorithm SHA256).Hash.ToLowerInvariant()
Invoke-Checked ssh ($sshOptions + @($SSHHost,"sudo -n bash '$stage/publish-server.sh' '$label' '$bundleHash'"))
if ($UploadFile) {
    $uploadArgs = @('files','upload','--path',[IO.Path]::GetFullPath($UploadFile))
    if ($ConfigPath) { $uploadArgs += @('--config',[IO.Path]::GetFullPath($ConfigPath)) }
    Invoke-Checked "$work/remote-agent-upload.exe" $uploadArgs
}
