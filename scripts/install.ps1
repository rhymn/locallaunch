#Requires -Version 5.1
$ErrorActionPreference = "Stop"

$Version = if ($env:VERSION) { $env:VERSION } else { "0.1.0" }
$InstallDir = "$env:APPDATA\LocalLaunch"
$BinaryName = "locallaunch.exe"
$TaskName = "LocalLaunch"

function Detect-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

$Arch = Detect-Arch

Write-Host "Installing LocalLaunch v$Version (windows/$Arch)..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$BinaryUrl = "https://github.com/rhymn/locallaunch/releases/download/v$Version/locallaunch-windows-$Arch.exe"
Write-Host "Downloading from: $BinaryUrl"

Invoke-WebRequest -Uri $BinaryUrl -OutFile "$InstallDir\$BinaryName"

Write-Host "Binary installed to: $InstallDir\$BinaryName"

$Action = New-ScheduledTaskAction -Execute "$InstallDir\$BinaryName"
$Trigger = New-ScheduledTaskTrigger -AtLogon
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable

Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Settings $Settings -Force

Start-ScheduledTask -TaskName $TaskName

Write-Host ""
Write-Host "Installation complete!"
Write-Host "Config: $InstallDir\config.json"
Write-Host ""
Write-Host "Usage:"
Write-Host "  $InstallDir\$BinaryName           # Start server"
Write-Host "  $InstallDir\$BinaryName token     # Show auth token"
Write-Host "  $InstallDir\$BinaryName version   # Show version"
