param([switch]$NoStart, [switch]$SkipCodex)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$server = '__SERVER__'
$codexOnly = '__MODE__' -eq 'codex'
$cliOnly = '__MODE__' -eq 'cli'
$setupCodex = '__SETUP_CODEX__' -eq 'true' -and !$SkipCodex
$root = Join-Path $HOME '.remote-agent'
$config = Join-Path $root 'config.json'
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
    if ($setupCodex -or $cliOnly -or $codexOnly) {
        $env:PATH = "$bin;$env:PATH"
        $userPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
        if (($userPath -split ';') -notcontains $bin) { [Environment]::SetEnvironmentVariable('PATH', "$bin;$userPath", 'User') }
    }
    if ($cliOnly -or $codexOnly) {
        $localCLI = Join-Path $bin 'remote-agent.exe'
        if (Test-Path -LiteralPath $config) {
            & $temp update --installer --preserve-start --server $server --config $config
            if ($LASTEXITCODE -ne 0) { throw 'Local CLI update failed.' }
        } else { Move-Item -LiteralPath $temp -Destination $localCLI -Force }
        if ($cliOnly) { Write-Host 'Local CLI installed. No remote service enabled. Run remote-agent --help for commands in a new terminal.'; return }
    }
    if ($codexOnly) {
        $secureToken = Read-Host 'Paste a personal proxy Token from /agents (input hidden)' -AsSecureString
        $plainToken = (New-Object Net.NetworkCredential('', $secureToken)).Password
        try {
            $plainToken | & $localCLI codex setup --server $server --token-stdin
            if ($LASTEXITCODE -ne 0) { throw 'Codex setup failed; existing configuration was retained.' }
        } finally { Remove-Variable plainToken, secureToken -ErrorAction SilentlyContinue }
        Write-Host 'Codex ready. Open a new terminal to use codex. Fully quit and reopen Codex App, then start a new local chat.'
        return
    }
    if (Test-Path -LiteralPath $config) {
        if ($setupCodex) {
            & $temp codex setup --config $config --server $server
            if ($LASTEXITCODE -ne 0) { throw 'Codex setup failed. Device binding and existing Agent are retained; retry this installer.' }
        }
        & $temp update --installer "--preserve-start=$($NoStart.IsPresent.ToString().ToLowerInvariant())" --server $server --config $config
        if ($LASTEXITCODE -ne 0) { throw 'Agent upgrade failed; existing device binding is retained.' }
        if ($NoStart) { & (Join-Path $bin 'remote-agent.exe') stop --config $config; if ($LASTEXITCODE -ne 0) { throw 'Could not stop remote service.' } }
        else {
            & (Join-Path $bin 'remote-agent.exe') autostart install --config $config
            if ($LASTEXITCODE -ne 0) { throw 'Agent updated, but autostart failed. Run remote-agent start to retry.' }
        }
        Write-Host 'Existing agent upgraded. Its account binding, projects, permissions, and conversations are retained. Reconnect in the web console.'
        return
    }
    $exe = Join-Path $bin 'remote-agent.exe'
    Move-Item -LiteralPath $temp -Destination $exe -Force
    '__TOKEN__' | & $exe enroll --server $server --config $config --token-stdin
    if ($LASTEXITCODE -ne 0) { throw 'Device enrollment failed.' }
    if ($setupCodex) {
        & $exe codex setup --config $config
        if ($LASTEXITCODE -ne 0) { throw 'Device bound, but Codex setup failed. Retry this installation command to finish setup.' }
    }
    if ($NoStart) { Write-Host 'Local CLI installed and bound. Remote service is stopped. Use remote-agent start when needed.'; return }
    & $exe autostart install --config $config
    if ($LASTEXITCODE -ne 0) { throw 'Device bound, but autostart failed. Run remote-agent autostart install again.' }
    Write-Host 'Codex and Agent ready. Return to the web console, connect this device and add a project. Qoder remains optional.'
} finally {
    if (Test-Path -LiteralPath $temp) { Remove-Item -LiteralPath $temp -Force }
}
