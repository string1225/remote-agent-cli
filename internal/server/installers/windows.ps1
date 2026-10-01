$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$server = '__SERVER__'
$root = Join-Path $HOME '.remote-agent'
$config = Join-Path $root 'config.json'
if (Test-Path -LiteralPath $config) {
    $bound = Get-Content -LiteralPath $config -Raw | ConvertFrom-Json
    if ($bound.server.TrimEnd('/') -ne $server.TrimEnd('/')) { throw 'This device is bound to a different server. Use its original installer or remote-agent update.' }
}
$arch = if ($env:PROCESSOR_ARCHITEW6432 -eq 'ARM64' -or $env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$name = "remote-agent-windows-$arch.exe"
$bin = Join-Path $root 'bin'
New-Item -ItemType Directory -Force -Path $bin | Out-Null
$sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
& icacls.exe $root /inheritance:r /grant:r "*${sid}:(OI)(CI)F" '*S-1-5-18:(OI)(CI)F' | Out-Null
if ($LASTEXITCODE -ne 0) { throw 'Could not protect agent credentials directory.' }
$temp = Join-Path $bin ('.remote-agent-install-' + [Guid]::NewGuid().ToString('N') + '.exe')
try {
    Invoke-WebRequest -UseBasicParsing "$server/downloads/$name" -OutFile $temp
    $expected = ((Invoke-WebRequest -UseBasicParsing "$server/downloads/$name.sha256").Content.Trim() -split '\s+')[0]
    if ($expected -notmatch '^[a-fA-F0-9]{64}$') { throw 'Invalid checksum.' }
    $checksumStream = [IO.File]::OpenRead($temp)
    $sha256 = [Security.Cryptography.SHA256]::Create()
    try { $actual = ([BitConverter]::ToString($sha256.ComputeHash($checksumStream))).Replace('-', '') }
    finally { $checksumStream.Dispose(); $sha256.Dispose() }
    if ($actual -ne $expected) { throw 'Download checksum mismatch.' }
    if (Test-Path -LiteralPath $config) {
        & $temp update --installer --server $server --config $config
        if ($LASTEXITCODE -ne 0) { throw 'Agent upgrade failed; existing device binding is retained.' }
        Write-Host 'Existing agent upgraded. Its account binding, projects, permissions, and conversations are retained. Reconnect in the web console.'
        return
    }
    $exe = Join-Path $bin 'remote-agent.exe'
    Move-Item -LiteralPath $temp -Destination $exe -Force
    '__TOKEN__' | & $exe enroll --server $server --config $config --token-stdin
    if ($LASTEXITCODE -ne 0) { throw 'Device enrollment failed.' }
    & $exe autostart install --config $config
    if ($LASTEXITCODE -ne 0) { throw 'Device bound, but autostart failed. Run remote-agent autostart install again.' }
    Write-Host 'Agent installed and started. Return to the web console to connect this device. Codex and Qoder services are optional.'
} finally {
    if (Test-Path -LiteralPath $temp) { Remove-Item -LiteralPath $temp -Force }
}
