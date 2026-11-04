# One-time host 1/2 -> 3 maintenance, run from an independent administrator shell.
# Does not install MSI, register services, or rewrite user configuration/history.
[CmdletBinding()]
param([switch]$StageOnly)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
if (!([Security.Principal.WindowsPrincipal]::new($identity)).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw 'Run this script from an independent administrator PowerShell after tasks and recordings finish. No MSI is required.'
}
$root = Join-Path ([Environment]::GetFolderPath('CommonApplicationData')) 'RemoteAgent'
$hostPath = Join-Path (Get-ItemProperty 'HKLM:\SOFTWARE\RemoteAgent').InstallDir 'remote-agent.exe'
$profilePath = Join-Path $root 'service.json'
$profile = Get-Content -LiteralPath $profilePath -Raw | ConvertFrom-Json
$server = ([string]$profile.updateServer).TrimEnd('/')
if (!$server) { $server = 'https://codex.sunny-string.cn/agents' }
$uri = [Uri]$server
if ($uri.Scheme -ne 'https' -or $uri.UserInfo -or $uri.Query -or $uri.Fragment) { throw 'The enrolled update source must be HTTPS.' }
$services = @{}
foreach ($name in @('RemoteAgent','RemoteAgentUpdater')) {
    $entry = Get-CimInstance Win32_Service -Filter "Name='$name'"
    $suffix = if ($name -eq 'RemoteAgent') { 'service run' } else { 'update-service run' }
    if (!$entry -or $entry.PathName -ne ('"' + $hostPath + '" ' + $suffix) -or $entry.State -ne 'Running') { throw "Unexpected installation or stopped service: $name" }
    $services[$name] = $entry
}
if ($services.RemoteAgent.StartName -ne 'NT AUTHORITY\LocalService' -or $services.RemoteAgentUpdater.StartName -notin @('LocalSystem','NT AUTHORITY\SYSTEM')) { throw 'Unexpected service identities.' }
$ancestor = $PID
while ($ancestor -gt 0) {
    if ($ancestor -in @($services.RemoteAgent.ProcessId,$services.RemoteAgentUpdater.ProcessId)) { throw 'Run outside the Remote Agent process tree; replacing its host would terminate this shell.' }
    $process = Get-CimInstance Win32_Process -Filter "ProcessId=$ancestor"
    if (!$process) { break }
    $ancestor = $process.ParentProcessId
}
$arch = switch ([Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) { 'X64' { 'amd64' } 'Arm64' { 'arm64' } default { throw 'Unsupported architecture.' } }
$healthPath = Join-Path $root 'health/status.json'
$statePath = Join-Path $root 'runtime/state.json'
$health = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
if (!$health.ready -or $health.paused -or $health.hostPid -ne $services.RemoteAgent.ProcessId) { throw 'The existing runtime must be healthy and unpaused.' }
$expectedVersion = $health.version
$logPath = Join-Path (Split-Path -Parent $profile.config) 'agent.log'
function LogOffset { if (Test-Path -LiteralPath $logPath) { return (Get-Item -LiteralPath $logPath).Length }; return 0 }
function Wait-Connected([long]$Offset) {
    $until = [DateTime]::UtcNow.AddSeconds(90)
    while ([DateTime]::UtcNow -lt $until) {
        $service = Get-CimInstance Win32_Service -Filter "Name='RemoteAgent'"
        $status = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
        if ($service.State -eq 'Running' -and $status.ready -and !$status.paused -and $status.hostPid -eq $service.ProcessId -and (Test-Path -LiteralPath $logPath)) {
            $stream = [IO.File]::Open($logPath, 'Open', 'Read', 'ReadWrite')
            try {
                if ($stream.Length -lt $Offset) { $Offset = 0 }
                $null = $stream.Seek($Offset, 'Begin')
                $reader = [IO.StreamReader]::new($stream)
                try { $connected = $reader.ReadToEnd().Contains('agent connected') } finally { $reader.Dispose() }
            } finally { $stream.Dispose() }
            if ($connected) { return }
        }
        Start-Sleep -Milliseconds 500
    }
    throw 'Runtime did not confirm a fresh connection to its enrolled server within 90 seconds.'
}
function Stop-Existing([string]$Name) {
    $service = Get-Service -Name $Name
    if ($service.Status -ne 'Stopped') { Stop-Service -InputObject $service; $service.WaitForStatus('Stopped', [TimeSpan]::FromSeconds(90)) }
}
function Start-Existing([string]$Name) {
    $service = Get-Service -Name $Name
    if ($service.Status -ne 'Running') { Start-Service -InputObject $service; $service.WaitForStatus('Running', [TimeSpan]::FromSeconds(90)) }
}
# A fresh private directory prevents a desktop user from replacing an executable
# between its download and elevated activation. Refuse redirected parent paths.
for ($parent = Get-Item -LiteralPath $root; $parent; $parent = $parent.Parent) {
    if ($parent.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Refusing a redirected maintenance directory.' }
}
$backup = Join-Path $root ('host-upgrade-' + [Guid]::NewGuid().ToString('N'))
$acl = [Security.AccessControl.DirectorySecurity]::new()
$acl.SetAccessRuleProtection($true, $false)
foreach ($sid in @('S-1-5-18','S-1-5-32-544')) {
    $acl.AddAccessRule([Security.AccessControl.FileSystemAccessRule]::new([Security.Principal.SecurityIdentifier]::new($sid), 'FullControl', 'ContainerInherit,ObjectInherit', 'None', 'Allow'))
}
$null = New-Item -ItemType Directory -Path $backup
Set-Acl -LiteralPath $backup -AclObject $acl
$candidate = Join-Path $backup 'host-new.exe'
$artifact = "remote-agent-windows-$arch.exe"
Add-Type -AssemblyName System.Net.Http
$handler = [Net.Http.HttpClientHandler]::new()
$handler.AllowAutoRedirect = $false
$http = [Net.Http.HttpClient]::new($handler)
$http.Timeout = [TimeSpan]::FromMinutes(3)
function Fetch([string]$Path) {
    $response = $http.GetAsync($server + '/downloads/' + $Path).GetAwaiter().GetResult()
    try {
        if ([int]$response.StatusCode -ne 200) { throw "Download returned HTTP $([int]$response.StatusCode)." }
        return ,($response.Content.ReadAsByteArrayAsync().GetAwaiter().GetResult())
    } finally { $response.Dispose() }
}
try {
    $checksum = [Text.Encoding]::UTF8.GetString((Fetch ($artifact + '.sha256'))).Trim()
    if ($checksum -notmatch ('^([0-9a-f]{64})  ' + [regex]::Escape($artifact) + '$')) { throw 'Invalid host checksum metadata.' }
    $digest = $Matches[1]
    [IO.File]::WriteAllBytes($candidate, (Fetch $artifact))
    if ((Get-FileHash -LiteralPath $candidate -Algorithm SHA256).Hash.ToLowerInvariant() -ne $digest) { throw 'Host checksum mismatch.' }
} finally { $http.Dispose(); $handler.Dispose() }
$version = & $candidate version 2>&1
if ($LASTEXITCODE -ne 0 -or ($version -join "`n") -notmatch '(host|宿主) 3') { throw 'Expected fixed host version 3.' }
if ($StageOnly) { Write-Output "Host 3 verified; services unchanged. Staged at $backup"; return }
$lock = $null
$paused = $false
$replaced = $false
$stopped = $false
try {
    $lock = [IO.File]::Open((Join-Path $root 'runtime/update.lock'), 'OpenOrCreate', 'ReadWrite', 'None')
    if (Test-Path -LiteralPath $statePath) {
        $state = Get-Content -LiteralPath $statePath -Raw | ConvertFrom-Json
        if ($state.pending) { throw 'Finish the existing runtime transaction before replacing the host.' }
    }
    $before = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
    & sc.exe control RemoteAgent 128 | Out-Null
    if ($LASTEXITCODE -ne 0) { throw 'Cannot request an idle pause.' }
    $paused = $true
    $until = [DateTime]::UtcNow.AddSeconds(35)
    do {
        Start-Sleep -Milliseconds 100
        $status = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
    } while ($status.generation -eq $before.generation -and [DateTime]::UtcNow -lt $until)
    if ($status.generation -eq $before.generation -or !$status.paused -or !$status.ready) { throw 'Tasks or recording are active, or the idle pause was not acknowledged. Retry when idle.' }
    $offset = LogOffset
    $stopped = $true
    Stop-Existing 'RemoteAgentUpdater'
    Stop-Existing 'RemoteAgent'
    Copy-Item -LiteralPath $profilePath -Destination (Join-Path $backup 'service.json')
    if (Test-Path -LiteralPath $statePath) { Copy-Item -LiteralPath $statePath -Destination (Join-Path $backup 'state.json') }
    # File.Replace preserves the original destination ACL and keeps its backup.
    [IO.File]::Replace($candidate, $hostPath, (Join-Path $backup 'host-old.exe'))
    $replaced = $true
    Start-Existing 'RemoteAgent'
    $paused = $false
    Wait-Connected $offset
    $status = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
    if ($status.version -ne $expectedVersion) { throw 'The retained runtime changed during host replacement.' }
    $retainedHostPID = $status.hostPid
    $offset = LogOffset
    $lock.Dispose(); $lock = $null
    Start-Existing 'RemoteAgentUpdater'
} catch {
    $failure = $_
    if ($replaced) {
        if ($lock) { $lock.Dispose(); $lock = $null }
        Stop-Existing 'RemoteAgentUpdater'
        Stop-Existing 'RemoteAgent'
        $offset = LogOffset
        [IO.File]::Replace((Join-Path $backup 'host-old.exe'), $hostPath, (Join-Path $backup 'host-failed.exe'))
        if (Test-Path -LiteralPath (Join-Path $backup 'state.json')) {
            Copy-Item -LiteralPath (Join-Path $backup 'state.json') -Destination $statePath -Force
        } elseif (Test-Path -LiteralPath $statePath) { Remove-Item -LiteralPath $statePath }
        Start-Existing 'RemoteAgent'
        $paused = $false
        Start-Existing 'RemoteAgentUpdater'
        Wait-Connected $offset
        Write-Warning 'Original host/runtime restored and server reconnection confirmed.'
    } elseif ($stopped) {
        Start-Existing 'RemoteAgent'
        $paused = $false
        Start-Existing 'RemoteAgentUpdater'
        Wait-Connected $offset
    }
    throw $failure
} finally {
    if ($lock) { $lock.Dispose() }
    if ($paused) { & sc.exe control RemoteAgent 130 | Out-Null }
}
# The verified host migration is complete. The normal update engine owns core
# rollback; a busy/refused core update must never force-stop newly accepted work.
& $hostPath update
$updateExit = $LASTEXITCODE
$status = Get-Content -LiteralPath $healthPath -Raw | ConvertFrom-Json
if ($status.hostPid -ne $retainedHostPID -or $status.version -ne $expectedVersion) { Wait-Connected $offset }
if ($updateExit -ne 0) { throw "Host 3 is installed and the existing runtime was retained by the update engine. Retry remote-agent update when idle. Backup: $backup" }
Write-Output "Host 3 and runtime update completed; server connection confirmed. Backup: $backup"
