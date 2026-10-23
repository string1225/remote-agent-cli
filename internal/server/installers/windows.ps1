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
if (!$cliOnly -and !$codexOnly) {
    # No enrollment or model secrets are passed to Windows Installer or UAC.
    if (Test-Path -LiteralPath $config) {
        $existing = Get-Content -LiteralPath $config -Raw -Encoding UTF8 | ConvertFrom-Json
        if ($existing.server.TrimEnd('/') -ne $server.TrimEnd('/')) { throw 'This device is bound to a different server. Use its original installer.' }
    }
    function Quote-PS([string]$Value) { return "'" + $Value.Replace("'", "''") + "'" }
    function Get-PackageSHA256([string]$Path) {
        $stream = [IO.File]::OpenRead($Path)
        $algorithm = [Security.Cryptography.SHA256]::Create()
        try { return ([BitConverter]::ToString($algorithm.ComputeHash($stream))).Replace('-', '') }
        finally { $stream.Dispose(); $algorithm.Dispose() }
    }
    function Invoke-AdminScript([string]$Script) {
        $wrapped = '$ErrorActionPreference=''Stop''; try { ' + $Script + '; exit 0 } catch { Write-Error $_; exit 1 }'
        $encoded = [Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($wrapped))
        $process = Start-Process -FilePath (Join-Path $env:SystemRoot 'System32/WindowsPowerShell/v1.0/powershell.exe') -ArgumentList @('-NoProfile','-NonInteractive','-WindowStyle','Hidden','-EncodedCommand',$encoded) -Verb RunAs -WindowStyle Hidden -Wait -PassThru
        if ($process.ExitCode -ne 0) { throw 'Administrator installation/configuration failed. Existing device data is retained; inspect the MSI log in the agent data directory and retry.' }
    }
    $msiName = "remote-agent-windows-$arch.msi"
    $msi = Join-Path ([IO.Path]::GetTempPath()) ('remote-agent-' + [Guid]::NewGuid().ToString('N') + '.msi')
    $profile = Join-Path $root 'service-profile.json'
    $installLog = Join-Path $root 'msi-install.log'
    try {
        Invoke-WebRequest -UseBasicParsing "$server/downloads/$msiName" -OutFile $msi
        $manifest = (Invoke-WebRequest -UseBasicParsing "$server/downloads/$msiName.sha256").Content.Trim() -split '\s+'
        if ($manifest.Count -ne 2 -or $manifest[0] -notmatch '^[a-fA-F0-9]{64}$' -or $manifest[1] -ne $msiName) { throw 'Invalid MSI checksum manifest.' }
        $expectedMSI = $manifest[0]
        if ((Get-PackageSHA256 $msi) -ne $expectedMSI) { throw 'MSI checksum mismatch.' }
        # Copy to an administrator-only staging directory and verify again before
        # handing it to msiexec, preventing a change to the unelevated download.
        $install = 'function Get-PackageSHA256 { ' + (Get-Command Get-PackageSHA256).ScriptBlock.ToString() + ' }; '
        $install += '$stage=Join-Path $env:ProgramData (''RemoteAgent-Installer-''+[Guid]::NewGuid().ToString(''N'')); New-Item -ItemType Directory -Path $stage | Out-Null; & icacls.exe $stage /inheritance:r /grant:r ''*S-1-5-18:(OI)(CI)F'' ''*S-1-5-32-544:(OI)(CI)F'' | Out-Null; if ($LASTEXITCODE -ne 0) { throw ''Cannot protect MSI staging directory'' }; '
        $install += '$package=Join-Path $stage ''agent.msi''; try { Copy-Item -LiteralPath ' + (Quote-PS $msi) + ' -Destination $package; if ((Get-PackageSHA256 $package) -ne ' + (Quote-PS $expectedMSI) + ') { throw ''MSI checksum mismatch'' }; '
        $install += '$result=Start-Process -FilePath msiexec.exe -ArgumentList @(''/i'',(''"''+$package+''"''),''/qn'',''/norestart'',''/L*v'',(''"''+' + (Quote-PS $installLog) + '+''"'')) -WindowStyle Hidden -Wait -PassThru; if ($result.ExitCode -notin @(0,3010)) { throw (''MSI failed: ''+$result.ExitCode) } } finally { if (Test-Path -LiteralPath $package) { Remove-Item -LiteralPath $package -Force }; Remove-Item -LiteralPath $stage }'
        Invoke-AdminScript $install
        $programFiles = [Environment]::GetEnvironmentVariable('ProgramW6432')
        if (!$programFiles) { $programFiles = $env:ProgramFiles }
        $exe = Join-Path $programFiles 'Remote Agent/remote-agent.exe'
        if (!(Test-Path -LiteralPath $config)) {
            '__TOKEN__' | & $exe enroll --server $server --config $config --token-stdin
            if ($LASTEXITCODE -ne 0) { throw 'MSI installed, but device enrollment failed. Retry the installer to finish setup.' }
        }
        if ($setupCodex) {
            & $exe codex setup --config $config --server $server
            if ($LASTEXITCODE -ne 0) { throw 'Device retained, but Codex setup failed. Retry with -SkipCodex or fix local Codex setup.' }
        }
        & $exe service profile --config $config --output $profile
        if ($LASTEXITCODE -ne 0) { throw 'Could not capture the original user tool environment.' }
        $configure = '& ' + (Quote-PS $exe) + ' service configure --profile ' + (Quote-PS $profile)
        if ($NoStart) { $configure += ' --no-start' }
        $configure += '; if ($LASTEXITCODE -ne 0) { throw ''Windows Service configuration failed; original data was retained.'' }'
        Invoke-AdminScript $configure
        Write-Host 'MSI installed. Configuration, projects and conversations stay in their original directory.'
        if ($NoStart) { Write-Host 'Service deliberately disabled by -NoStart; run remote-agent start as administrator to enable it.' }
        else { Write-Host 'Windows Service starts at boot and stays online after sign-out.' }
        return
    } finally {
        if (Test-Path -LiteralPath $msi) { Remove-Item -LiteralPath $msi -Force }
        if (Test-Path -LiteralPath $profile) { Remove-Item -LiteralPath $profile -Force }
    }
}
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
} finally {
    if (Test-Path -LiteralPath $temp) { Remove-Item -LiteralPath $temp -Force }
}
