$ErrorActionPreference = "Stop"

$InstallDir = "$env:APPDATA\kaddio-bridge"
$BinaryName = "kaddio-bridge.exe"
$TaskName = "Kaddio Bridge"
$RepoUrl = "https://github.com/kaddio/kaddio-bridge"

function Resolve-Version {
    $resp = Invoke-WebRequest -Uri "$RepoUrl/releases/latest" -MaximumRedirection 0 -ErrorAction SilentlyContinue
    if ($resp.StatusCode -eq 302) {
        $location = $resp.Headers.Location
        return ($location -split "/")[-1] -replace "^v", ""
    }
    Write-Error "Unable to resolve latest release version."
    exit 1
}

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
$Version = Resolve-Version

Write-Host "Installing Kaddio Bridge v$Version (windows/$Arch)..."

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$BinaryUrl = "$RepoUrl/releases/download/v$Version/kaddio-bridge-windows-$Arch.exe"
Write-Host "Downloading from: $BinaryUrl"

Invoke-WebRequest -Uri $BinaryUrl -OutFile "$InstallDir\$BinaryName"

Write-Host "Binary installed to: $InstallDir\$BinaryName"

# Stop any existing instance
Stop-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue
Get-Process -Name "kaddio-bridge" -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Seconds 1

$Action = New-ScheduledTaskAction -Execute "$InstallDir\$BinaryName"
$Trigger = New-ScheduledTaskTrigger -AtLogon
$Settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -StartWhenAvailable

Register-ScheduledTask -TaskName $TaskName -Action $Action -Trigger $Trigger -Settings $Settings -Force

Start-ScheduledTask -TaskName $TaskName

Write-Host ""
Write-Host "Installation complete!"
Write-Host "Config: $InstallDir\config.json"
