[CmdletBinding()]
param([string]$WixDirectory = '.local/wix314')
$ErrorActionPreference='Stop'
$root=[IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$compiler=Join-Path $env:SystemRoot 'Microsoft.NET/Framework64/v4.0.30319/csc.exe'
$work=Join-Path $root '.local/msi-activation/protocol tests'
New-Item -ItemType Directory -Force -Path $work | Out-Null
$dtf=Join-Path ([IO.Path]::GetFullPath($WixDirectory)) 'sdk/Microsoft.Deployment.WindowsInstaller.dll'
$fixture=Join-Path $work 'fixture.exe'
$runner=Join-Path $work 'runner.exe'
& $compiler /nologo /codepage:65001 /target:exe "/out:$fixture" /reference:System.Web.Extensions.dll (Join-Path $root 'packaging/windows/ActivationProtocolFixture.cs')
if($LASTEXITCODE -ne 0){throw 'Fixture compilation failed.'}
& $compiler /nologo /codepage:65001 /target:exe /main:ActivationProtocolTests "/out:$runner" "/reference:$dtf" /reference:System.Web.Extensions.dll (Join-Path $root 'packaging/windows/ActivationActions.cs') (Join-Path $root 'packaging/windows/ActivationProtocolTests.cs')
if($LASTEXITCODE -ne 0){throw 'Protocol test compilation failed.'}
Copy-Item -LiteralPath $dtf -Destination $work -Force
& $runner $fixture
if($LASTEXITCODE -ne 0){throw 'Activation protocol test failed.'}
