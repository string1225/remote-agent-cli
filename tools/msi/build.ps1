[CmdletBinding()]
param(
    [string]$PackageVersion = '1.0.0',
    [string]$ReleaseVersion = '',
    [ValidateSet('amd64','arm64')][string[]]$Architectures = @('amd64','arm64'),
    [string]$WixDirectory = '',
    [string]$UpdatePublicKey = '',
    [switch]$SkipGoBuild,
    [switch]$InjectFailure
)
$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
if ($PackageVersion -notmatch '^([0-9]+)\.([0-9]+)\.([0-9]+)$' -or [int]$Matches[1] -gt 255 -or [int]$Matches[2] -gt 255 -or [int]$Matches[3] -gt 65535) { throw 'MSI version must be major.minor.build (255.255.65535 maximum).' }
if (!$ReleaseVersion) { $ReleaseVersion = (& git -C $root rev-parse --short HEAD).Trim() }
if ($ReleaseVersion -notmatch '^[A-Za-z0-9._-]{1,80}$') { throw 'Invalid release version.' }
if ($UpdatePublicKey -and $UpdatePublicKey -notmatch '^[0-9a-f]{64}$') { throw 'Invalid update public key.' }
if (!$WixDirectory) {
    $WixDirectory = Join-Path $root '.local/wix314'
    $archive = Join-Path $root '.local/wix314.zip'
    $digest = '6AC824E1642D6F7277D0ED7EA09411A508F6116BA6FAE0AA5F2C7DAA2FF43D31'
    if (!(Test-Path -LiteralPath (Join-Path $WixDirectory 'candle.exe'))) {
        New-Item -ItemType Directory -Force -Path $WixDirectory | Out-Null
        Invoke-WebRequest -UseBasicParsing 'https://github.com/wixtoolset/wix3/releases/download/wix3141rtm/wix314-binaries.zip' -OutFile $archive
        if ((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash -ne $digest) { throw 'WiX archive checksum mismatch.' }
        Expand-Archive -LiteralPath $archive -DestinationPath $WixDirectory -Force
    }
}
$wix = [IO.Path]::GetFullPath($WixDirectory)
$dist = Join-Path $root 'dist'
$objects = Join-Path $root '.local/msi-build'
New-Item -ItemType Directory -Force -Path $dist,$objects | Out-Null
Push-Location $root
$previousOS, $previousArch, $previousCGO = $env:GOOS, $env:GOARCH, $env:CGO_ENABLED
try {
    foreach ($arch in $Architectures) {
        $binary = Join-Path $dist "remote-agent-windows-$arch.exe"
		$coreBinary = Join-Path $dist "remote-agent-core-windows-$arch.exe"
        if (!$SkipGoBuild) {
            $env:GOOS = 'windows'; $env:GOARCH = $arch; $env:CGO_ENABLED = '0'
            $linkFlags = "-s -w -X github.com/string1225/remote-agent-cli/internal/buildinfo.Version=$ReleaseVersion"
            if ($UpdatePublicKey) { $linkFlags += " -X github.com/string1225/remote-agent-cli/internal/clientupdate.ReleasePublicKey=$UpdatePublicKey" }
            & go build -trimpath "-ldflags=$linkFlags" -o $binary ./cmd/agent-host
            if ($LASTEXITCODE -ne 0) { throw 'Go agent build failed.' }
			& go build -trimpath "-ldflags=$linkFlags" -o $coreBinary ./cmd/agent
			if ($LASTEXITCODE -ne 0) { throw 'Go worker build failed.' }
        }
        $platform = if ($arch -eq 'amd64') { 'x64' } else { 'arm64' }
        $object = Join-Path $objects "agent-$arch.wixobj"
        $package = Join-Path $dist "remote-agent-windows-$arch.msi"
        $testDefine = @()
        if ($InjectFailure) { $testDefine = @('-dInjectFailure=1') }
        & (Join-Path $wix 'candle.exe') -nologo -arch $platform "-dPlatform=$platform" "-dPackageVersion=$PackageVersion" "-dAgentBinary=$binary" "-dCoreBinary=$coreBinary" "-dSourceRoot=$root" @testDefine -out $object packaging/windows/agent.wxs
        if ($LASTEXITCODE -ne 0) { throw 'WiX compilation failed.' }
        & (Join-Path $wix 'light.exe') -nologo -ext (Join-Path $wix 'WixUIExtension.dll') -cultures:zh-CN -out $package $object
        if ($LASTEXITCODE -ne 0) { throw 'WiX linking/validation failed.' }
        # These checks open a read-only MSI session and only run costing/property
        # actions. They never install files or touch the machine service.
        if ($arch -eq 'amd64' -and [Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'X64') {
            & (Join-Path $PSScriptRoot 'verify-ui.ps1') -PackagePath $package
        }
        foreach ($artifact in @($binary,$coreBinary,$package)) {
            $hash = (Get-FileHash -LiteralPath $artifact -Algorithm SHA256).Hash.ToLowerInvariant()
            [IO.File]::WriteAllText("$artifact.sha256", "$hash  $([IO.Path]::GetFileName($artifact))`n", [Text.UTF8Encoding]::new($false))
        }
        Write-Output "Built $package ($PackageVersion, $ReleaseVersion)"
    }
} finally {
    $env:GOOS = $previousOS; $env:GOARCH = $previousArch; $env:CGO_ENABLED = $previousCGO
    Pop-Location
}
