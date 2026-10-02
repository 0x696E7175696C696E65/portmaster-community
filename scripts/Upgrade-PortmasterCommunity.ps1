param(
    [Parameter(Mandatory=$true)][string]$Payload,
    [Parameter(Mandatory=$true)][string]$ResultFile
)
$ErrorActionPreference = 'Stop'
$installRoot = [IO.Path]::GetFullPath((Join-Path $env:ProgramFiles 'Portmaster Community'))
$principal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Run the upgrade in an administrator PowerShell window.' }
$backupRoot = Join-Path $installRoot ('backups\' + (Get-Date).ToUniversalTime().ToString('yyyyMMdd-HHmmss'))
$targets = @('portmaster-core.exe', 'portmaster.zip')
$stopped = $false
$replaced = $false
try {
    $service = Get-CimInstance Win32_Service -Filter "Name='PortmasterCore'"
    $expectedExecutable = Join-Path $installRoot 'portmaster-core.exe'
    if (-not $service -or -not $service.PathName.StartsWith(('"' + $expectedExecutable + '" '), [StringComparison]::OrdinalIgnoreCase)) { throw 'The installed service does not match this community installation.' }
    $manifest = Get-Content -LiteralPath (Join-Path $Payload 'manifest.json') -Raw | ConvertFrom-Json
    if ($manifest.Version -ne '2.2.3-community.2') { throw 'Unexpected upgrade package version.' }
    New-Item -ItemType Directory -Path $backupRoot | Out-Null
    New-Item -ItemType Directory -Path (Join-Path $backupRoot 'staged') | Out-Null
    foreach ($name in $targets) {
        $entry = @($manifest.Files | Where-Object { $_.Path -eq $name })
        if ($entry.Count -ne 1) { throw "Missing or ambiguous manifest entry: $name" }
        $staged = Join-Path $backupRoot ('staged\' + $name)
        Copy-Item -LiteralPath (Join-Path $Payload $name) -Destination $staged
        if ((Get-FileHash -LiteralPath $staged -Algorithm SHA256).Hash -ne $entry[0].SHA256) { throw "Payload integrity check failed: $name" }
        Copy-Item -LiteralPath (Join-Path $installRoot $name) -Destination (Join-Path $backupRoot $name)
    }
    Get-Process -Name portmaster -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq (Join-Path $installRoot 'portmaster.exe') } | Stop-Process
    Stop-Service -Name PortmasterCore
    (Get-Service PortmasterCore).WaitForStatus('Stopped', [TimeSpan]::FromSeconds(45))
    $stopped = $true
    $replaced = $true
    foreach ($name in $targets) { Copy-Item -LiteralPath (Join-Path $backupRoot ('staged\' + $name)) -Destination (Join-Path $installRoot $name) -Force }
    Start-Service -Name PortmasterCore
    (Get-Service PortmasterCore).WaitForStatus('Running', [TimeSpan]::FromSeconds(45))
    $healthy = $false
    for ($attempt=0; $attempt -lt 20; $attempt++) {
        try {
            $response = Invoke-WebRequest -Uri 'http://127.0.0.1:817/api/v1/ready' -UseBasicParsing -TimeoutSec 2
            if ($response.StatusCode -eq 200) { $healthy = $true; break }
        } catch { Start-Sleep -Seconds 1 }
    }
    if (-not $healthy) { throw 'Updated service did not pass the local API health check.' }
    $removedEndpointStatus = 0
    try { $response = Invoke-WebRequest -Uri 'http://127.0.0.1:817/api/v1/updates/from-url' -UseBasicParsing -TimeoutSec 3; $removedEndpointStatus = [int]$response.StatusCode }
    catch { if ($_.Exception.Response) { $removedEndpointStatus = [int]$_.Exception.Response.StatusCode } else { throw } }
    if ($removedEndpointStatus -ne 404) { throw "Disabled upgrade endpoint returned $removedEndpointStatus instead of 404." }
    New-ItemProperty -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\PortmasterCommunity' -Name DisplayVersion -Value $manifest.Version -PropertyType String -Force | Out-Null
    @{Upgraded=$true; Version=$manifest.Version; Backup=$backupRoot; RemovedEndpointStatus=$removedEndpointStatus; Time=(Get-Date).ToString('o')} | ConvertTo-Json | Set-Content -LiteralPath $ResultFile -Encoding UTF8
} catch {
    $failure = $_.Exception.Message
    $rolledBack = $false
    if ($stopped) {
        try {
            Stop-Service -Name PortmasterCore -ErrorAction SilentlyContinue
            (Get-Service PortmasterCore).WaitForStatus('Stopped', [TimeSpan]::FromSeconds(45))
            if ($replaced) { foreach ($name in $targets) { Copy-Item -LiteralPath (Join-Path $backupRoot $name) -Destination (Join-Path $installRoot $name) -Force } }
            Start-Service -Name PortmasterCore
            (Get-Service PortmasterCore).WaitForStatus('Running', [TimeSpan]::FromSeconds(45))
            $rolledBack = $true
        } catch { $failure += '; rollback failed: ' + $_.Exception.Message }
    }
    @{Upgraded=$false; Error=$failure; RolledBack=$rolledBack; Backup=$backupRoot; Time=(Get-Date).ToString('o')} | ConvertTo-Json | Set-Content -LiteralPath $ResultFile -Encoding UTF8
    throw $failure
}
