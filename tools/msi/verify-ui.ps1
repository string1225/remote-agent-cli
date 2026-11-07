[CmdletBinding()]
param([Parameter(Mandatory=$true)][string]$PackagePath,[switch]$TablesOnly)

# Read-only package/session checks. Never executes InstallInitialize,
# ExecuteAction, file/service actions, or any custom EXE action.
$ErrorActionPreference = 'Stop'
$package = [IO.Path]::GetFullPath($PackagePath)
$installer = New-Object -ComObject WindowsInstaller.Installer
$database = $null
$session = $null
function Set-SessionProperty([string]$Name, [string]$Value) {
    $session.GetType().InvokeMember('Property', [Reflection.BindingFlags]::SetProperty, $null, $session, @($Name,$Value)) | Out-Null
}
function Read-Scalar([string]$Query) {
    $view = $database.OpenView($Query)
    try {
        $null = $view.Execute()
        $record = $view.Fetch()
        if (!$record) { throw "MSI query returned no row: $Query" }
        try { return $record.StringData(1) } finally { [Runtime.InteropServices.Marshal]::FinalReleaseComObject($record) | Out-Null }
    } finally { $null = $view.Close(); [Runtime.InteropServices.Marshal]::FinalReleaseComObject($view) | Out-Null }
}
function Format-ControlText([string]$Control) {
    $text = Read-Scalar "SELECT ``Text`` FROM ``Control`` WHERE ``Dialog_``='AgentWelcome' AND ``Control``='$Control'"
    $record = $installer.CreateRecord(0)
    try {
        $record.GetType().InvokeMember('StringData', [Reflection.BindingFlags]::SetProperty, $null, $record, @(0,$text)) | Out-Null
        return $session.FormatRecord($record)
    } finally { [Runtime.InteropServices.Marshal]::FinalReleaseComObject($record) | Out-Null }
}
try {
    $database = $installer.OpenDatabase($package,0)
    if ((Read-Scalar 'SELECT `Target` FROM `CustomAction` WHERE `Action`=''InstallUserSkill''') -ne 'skill install') { throw 'MSI must install the shared user skill.' }
    if ((Read-Scalar 'SELECT `Source` FROM `CustomAction` WHERE `Action`=''InstallUserSkill''') -ne 'AgentExe') { throw 'Skill installation must use the active runtime so MSI repair preserves newer skills.' }
    if ([int](Read-Scalar 'SELECT `Sequence` FROM `InstallExecuteSequence` WHERE `Action`=''InstallUserSkill''') -le [int](Read-Scalar 'SELECT `Sequence` FROM `InstallExecuteSequence` WHERE `Action`=''InstallFinalize''')) { throw 'User skill installation must follow a successful machine install.' }
    if ((Read-Scalar 'SELECT `Property` FROM `Control` WHERE `Dialog_`=''AgentComplete'' AND `Control`=''BindingCode''') -ne 'BINDING_CODE') { throw 'Completion must accept a device binding code in the MSI.' }
    foreach ($choice in @('AUTO_START','CONFIGURE_CODEX','INSTALL_CODEX')) {
        if ((Read-Scalar "SELECT ``Value`` FROM ``Property`` WHERE ``Property``='$choice'") -ne '1') { throw "Default option must be enabled: $choice" }
    }
    $proxy = Read-Scalar 'SELECT `Property` FROM `Control` WHERE `Dialog_`=''AgentComplete'' AND `Control`=''WebProxy'''
    if ($proxy -ne 'WEB_PROXY') { throw 'Missing independent web proxy choice.' }
    $hidden = Read-Scalar 'SELECT `Value` FROM `Property` WHERE `Property`=''MsiHiddenProperties'''
    if ($hidden -notmatch '(^|;)BINDING_CODE(;|$)') { throw 'Binding code must be hidden from MSI logs.' }
    $activate = Read-Scalar 'SELECT `Argument` FROM `ControlEvent` WHERE `Dialog_`=''AgentComplete'' AND `Control_`=''Activate'' AND `Event`=''DoAction'''
    if ($activate -ne 'ActivateDevice') { throw 'Activation must execute inside the existing MSI UI.' }
    if ($TablesOnly) {
        $view = $database.OpenView('SELECT `Value` FROM `Property` WHERE `Property`=''WEB_PROXY''')
        try { $null=$view.Execute(); $row=$view.Fetch(); if ($row -and $row.StringData(1)) { throw 'Web proxy must default to off.' } }
        finally { $null=$view.Close(); [Runtime.InteropServices.Marshal]::FinalReleaseComObject($view) | Out-Null }
        Write-Output 'MSI tables verified: inline activation, default choices, hidden binding code. Installer session / ICE checks not run.'
        return
    }
    $session = $installer.OpenPackage($package,0)
    if ($session.GetType().InvokeMember('Property', [Reflection.BindingFlags]::GetProperty, $null, $session, @('WEB_PROXY'))) { throw 'Web proxy must default to off.' }
    foreach ($action in @('FindRelatedProducts','CostInitialize','FileCost','CostFinalize')) {
        if ($session.DoAction($action) -ne 1) { throw "Read-only MSI action failed: $action" }
    }
    $disable = Read-Scalar "SELECT ``Condition`` FROM ``ControlCondition`` WHERE ``Dialog_``='AgentWelcome' AND ``Control_``='Install' AND ``Action``='Disable'"
    Set-SessionProperty 'OutOfDiskSpace' '0'
    if ($session.EvaluateCondition($disable) -ne 0) { throw 'Install/upgrade button is disabled even though OutOfDiskSpace=0.' }
    Set-SessionProperty 'OutOfDiskSpace' '1'
    if ($session.EvaluateCondition($disable) -ne 1) { throw 'Install/upgrade button permits a real insufficient-disk-space state.' }
    $enable = Read-Scalar "SELECT ``Condition`` FROM ``ControlCondition`` WHERE ``Dialog_``='AgentWelcome' AND ``Control_``='Install' AND ``Action``='Enable'"
    if ($session.EvaluateCondition($enable) -ne 0) { throw 'Install/upgrade button is enabled despite insufficient disk space.' }
    Set-SessionProperty 'OutOfDiskSpace' '0'
    if ($session.EvaluateCondition($enable) -ne 1) { throw 'Install/upgrade button does not re-enable after disk space is available.' }

    $actions = @()
    $view = $database.OpenView('SELECT `Action`,`Condition` FROM `InstallUISequence` ORDER BY `Sequence`')
    try {
        $null = $view.Execute()
        while ($record = $view.Fetch()) {
            if ($record.StringData(1).StartsWith('SetAgent')) { $actions += [pscustomobject]@{Name=$record.StringData(1);Condition=$record.StringData(2)} }
            [Runtime.InteropServices.Marshal]::FinalReleaseComObject($record) | Out-Null
        }
    } finally { $null = $view.Close(); [Runtime.InteropServices.Marshal]::FinalReleaseComObject($view) | Out-Null }
    foreach ($action in $actions) {
        $type = [int](Read-Scalar "SELECT ``Type`` FROM ``CustomAction`` WHERE ``Action``='$($action.Name)'")
        if (($type -band 63) -ne 51) { throw "Refusing to run a non-property custom action: $($action.Name)" }
    }
    # ASCII-only source also works in Windows PowerShell 5.1 without a BOM.
    $installText = [string][char]0x5b89 + [char]0x88c5
    $upgradeText = [string][char]0x5347 + [char]0x7ea7
    $repairText = [string][char]0x4fee + [char]0x590d
    $retainText = [string][char]0x4fdd + [char]0x7559
    foreach ($mode in @(
        @{Name='install';Installed='';Upgrade='';Expected=$installText},
        @{Name='upgrade';Installed='';Upgrade='{fixture-previous-product}';Expected=$upgradeText},
        @{Name='repair';Installed='1';Upgrade='';Expected=$repairText}
    )) {
        foreach ($property in @('AGENT_SETUP_ACTION','AGENT_SETUP_TITLE','AGENT_SETUP_DESCRIPTION','AGENT_COMPLETE_TITLE')) {
            Set-SessionProperty $property (Read-Scalar "SELECT ``Value`` FROM ``Property`` WHERE ``Property``='$property'")
        }
        Set-SessionProperty 'Installed' $mode.Installed
        Set-SessionProperty 'WIX_UPGRADE_DETECTED' $mode.Upgrade
        foreach ($action in $actions) {
            if ($session.EvaluateCondition($action.Condition) -eq 1 -and $session.DoAction($action.Name) -ne 1) { throw "UI property action failed: $($action.Name)" }
        }
        if ((Format-ControlText 'Install') -ne $mode.Expected) { throw "Incorrect primary button text for $($mode.Name)." }
        if (!(Format-ControlText 'Title').Contains($mode.Expected)) { throw "Incorrect title for $($mode.Name)." }
        if ($mode.Name -ne 'install' -and !(Format-ControlText 'Description').Contains($retainText)) { throw "Missing data retention explanation for $($mode.Name)." }
    }
    $repairCondition = Read-Scalar "SELECT ``Condition`` FROM ``ControlEvent`` WHERE ``Dialog_``='AgentWelcome' AND ``Control_``='Install' AND ``Event``='[REINSTALL]'"
    if ($session.EvaluateCondition($repairCondition) -ne 1) { throw 'Repair button does not request reinstallation for an installed product.' }
    if ((Read-Scalar "SELECT ``Argument`` FROM ``ControlEvent`` WHERE ``Dialog_``='AgentWelcome' AND ``Control_``='Install' AND ``Event``='[REINSTALL]'") -ne 'ALL') { throw 'Repair button does not restore all installed features.' }
    $welcomeCondition = Read-Scalar "SELECT ``Condition`` FROM ``InstallUISequence`` WHERE ``Action``='AgentWelcome'"
    if ($session.EvaluateCondition($welcomeCondition) -ne 1) { throw 'Repair confirmation dialog is skipped for an installed product.' }
    Set-SessionProperty 'REMOVE' 'ALL'
    if ($session.EvaluateCondition($welcomeCondition) -ne 0) { throw 'Uninstallation unexpectedly shows the install/repair dialog.' }
    Write-Output 'MSI UI verified: disk-space 0/1, install, upgrade and repair (no installation performed).'
} finally {
    foreach ($com in @($session,$database,$installer)) { if ($null -ne $com) { [Runtime.InteropServices.Marshal]::FinalReleaseComObject($com) | Out-Null } }
}
