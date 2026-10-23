$ErrorActionPreference = 'Stop'
$root = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$fixtures = Join-Path $root '.local/msi-acceptance'
New-Item -ItemType Directory -Force -Path $fixtures | Out-Null
foreach ($build in @(@('1.0.1','base'),@('1.0.2','upgrade'),@('1.0.3','failure'))) {
    $arguments = @{ PackageVersion=$build[0]; ReleaseVersion="acceptance-$($build[1])"; Architectures=@('amd64') }
    if ($build[1] -eq 'failure') { $arguments.InjectFailure = $true }
    & (Join-Path $PSScriptRoot 'build.ps1') @arguments
    Copy-Item -LiteralPath (Join-Path $root 'dist/remote-agent-windows-amd64.msi') -Destination (Join-Path $fixtures "$($build[1]).msi") -Force
}
$env:REMOTE_AGENT_MSI_ACCEPTANCE = '1'
$env:REMOTE_AGENT_MSI_BASE = Join-Path $fixtures 'base.msi'
$env:REMOTE_AGENT_MSI_UPGRADE = Join-Path $fixtures 'upgrade.msi'
$env:REMOTE_AGENT_MSI_FAILURE = Join-Path $fixtures 'failure.msi'
$env:REMOTE_AGENT_MSI_LOG_DIR = $fixtures
Push-Location $root
try {
    & go test ./internal/agent -run '^TestMSIWindowsServiceLifecycle$' -count=1 -v -timeout 10m
    if ($LASTEXITCODE -ne 0) { throw 'Windows MSI/Service acceptance failed.' }
} finally {
    Remove-Item Env:REMOTE_AGENT_MSI_ACCEPTANCE,Env:REMOTE_AGENT_MSI_BASE,Env:REMOTE_AGENT_MSI_UPGRADE,Env:REMOTE_AGENT_MSI_FAILURE,Env:REMOTE_AGENT_MSI_LOG_DIR -ErrorAction SilentlyContinue
    Pop-Location
}
