# Golazo Installer for Windows
# Run with: irm https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"

# ASCII art logo
$asciiLogo = @"
  ________       .__                       
 /  _____/  ____ |  | _____  ____________  
/   \  ___ /  _ \|  | \__  \ \___   /  _ \ 
\    \_\  (  <_> )  |__/ __ \_/    (  <_> )
 \______  /\____/|____(____  /_____ \____/ 
        \/                 \/      \/      
"@

$repo = "0xjuanma/golazo"
$binaryName = "golazo"

# Print header
Write-Host $asciiLogo -ForegroundColor Cyan
Write-Host ""
Write-Host "Installing $binaryName..." -ForegroundColor Green
Write-Host ""

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_IDENTIFIER -match "ARM") {
        "arm64"
    } else {
        "amd64"
    }
} else {
    Write-Host "Unsupported architecture: 32-bit systems are not supported" -ForegroundColor Red
    exit 1
}

Write-Host "Detected: windows/$arch" -ForegroundColor Cyan

# Get the latest release tag
Write-Host "Fetching latest release..." -ForegroundColor Cyan
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $latest = $release.tag_name
} catch {
    Write-Host "Failed to fetch latest release: $_" -ForegroundColor Red
    exit 1
}

Write-Host "Latest version: $latest" -ForegroundColor Cyan

# Construct download URL
$fileName = "$binaryName-windows-$arch.exe"
$url = "https://github.com/$repo/releases/download/$latest/$fileName"

# Determine install directory
$installDir = "$env:LOCALAPPDATA\Programs\golazo"
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

$installPath = Join-Path $installDir "$binaryName.exe"

# Download the binary
Write-Host "Downloading $binaryName $latest for windows/$arch..." -ForegroundColor Cyan
try {
    Invoke-WebRequest -Uri $url -OutFile $installPath -UseBasicParsing
} catch {
    Write-Host "Failed to download binary: $_" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable(
        "Path",
        "$userPath;$installDir",
        "User"
    )
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added to PATH. You may need to restart your terminal for changes to take effect." -ForegroundColor Yellow
}

# Verify installation
if (Test-Path $installPath) {
    Write-Host ""
    Write-Host "✓ $binaryName $latest installed successfully!" -ForegroundColor Green
    Write-Host "  Installed to: $installPath" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Run '$binaryName' to start watching live football matches." -ForegroundColor Green
} else {
    Write-Host "Installation failed" -ForegroundColor Red
    exit 1
}

