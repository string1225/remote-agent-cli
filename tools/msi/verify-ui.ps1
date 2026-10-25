[CmdletBinding()]
param([Parameter(Mandatory=$true)][string]$PackagePath)

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
    if ((Read-Scalar 'SELECT `Value` FROM `Property` WHERE `Property`=''TOKEN_SETUP_COMMAND''') -ne 'remote-agent codex setup') { throw 'Default Codex setup must use the enrolled account.' }
    if ((Read-Scalar 'SELECT `Property` FROM `Control` WHERE `Dialog_`=''AgentComplete'' AND `Control`=''TokenCommand''') -ne 'DEVICE_SETUP_COMMAND') { throw 'Completion must show website binding before Token setup.' }
    if ((Read-Scalar 'SELECT `Property` FROM `Control` WHERE `Dialog_`=''AgentComplete'' AND `Control`=''DeviceCommand''') -ne 'TOKEN_SETUP_COMMAND') { throw 'Completion must show account-scoped Codex setup after binding.' }
    $session = $installer.OpenPackage($package,0)
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
