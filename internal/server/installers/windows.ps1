$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$server = '__SERVER__'
$root = Join-Path $HOME '.remote-agent'
$config = Join-Path $root 'config.json'
if (Test-Path -LiteralPath $config) { throw 'An agent is already bound. Revoke it and remove its config before reinstalling.' }
$arch = if ($env:PROCESSOR_ARCHITEW6432 -eq 'ARM64' -or $env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$name = "remote-agent-windows-$arch.exe"
$bin = Join-Path $root 'bin'
New-Item -ItemType Directory -Force -Path $bin | Out-Null
$sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
& icacls.exe $root /inheritance:r /grant:r "*${sid}:(OI)(CI)F" '*S-1-5-18:(OI)(CI)F' | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Could not protect agent credentials directory.' }
$temp = Join-Path $bin ($name + '.download')
try {
    Invoke-WebRequest -UseBasicParsing "$server/downloads/$name" -OutFile $temp
    $expected = ((Invoke-WebRequest -UseBasicParsing "$server/downloads/$name.sha256").Content.Trim() -split '\s+')[0]
    if ($expected -notmatch '^[a-fA-F0-9]{64}$') { throw 'Invalid checksum.' }
    if ((Get-FileHash -LiteralPath $temp -Algorithm SHA256).Hash -ne $expected) { throw 'Download checksum mismatch.' }
    $exe = Join-Path $bin 'remote-agent.exe'
    Move-Item -LiteralPath $temp -Destination $exe -Force
    '__TOKEN__' | & $exe enroll --server $server --config $config --token-stdin
    if ($LASTEXITCODE -ne 0) { throw 'Device enrollment failed.' }
    & $exe autostart install --config $config
    if ($LASTEXITCODE -ne 0) { throw 'Device bound, but autostart failed. Run remote-agent autostart install again.' }
    Write-Host 'Agent installed and started. Return to the web console to register Codex or Qoder.'
} finally {
    if (Test-Path -LiteralPath $temp) { Remove-Item -LiteralPath $temp -Force }
}
