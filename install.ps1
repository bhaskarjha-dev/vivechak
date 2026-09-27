#Requires -Version 5.1
# Vivechak installer for Windows
# Usage: irm https://raw.githubusercontent.com/bhaskarjha-dev/vivechak/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'

$Repo = 'bhaskarjha-dev/vivechak'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\vivechak'

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64' -or $env:PROCESSOR_IDENTIFIER -match 'ARM') {
        'arm64'
    } else {
        'amd64'
    }
} else {
    Write-Error 'Vivechak requires a 64-bit operating system'
    return
}

Write-Host "Detected: windows/$Arch" -ForegroundColor Cyan

# Get latest release
Write-Host 'Fetching latest release...'
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
$Tag = $Release.tag_name
$Version = $Tag.TrimStart('v')
Write-Host "Latest version: $Tag" -ForegroundColor Cyan

# Download
$Archive = "vivechak_${Version}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Archive"
$ChecksumUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) "vivechak-install-$([guid]::NewGuid().ToString('N').Substring(0,8))"
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    Write-Host "Downloading $Archive..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $TmpDir $Archive) -UseBasicParsing
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile (Join-Path $TmpDir 'checksums.txt') -UseBasicParsing

    # Verify checksum
    Write-Host 'Verifying checksum...'
    $ExpectedLine = Get-Content (Join-Path $TmpDir 'checksums.txt') | Where-Object { $_ -match $Archive }
    if ($ExpectedLine) {
        $ExpectedHash = ($ExpectedLine -split '\s+')[0]
        $ActualHash = (Get-FileHash (Join-Path $TmpDir $Archive) -Algorithm SHA256).Hash.ToLower()
        if ($ActualHash -ne $ExpectedHash.ToLower()) {
            Write-Error "Checksum mismatch! Expected: $ExpectedHash, Got: $ActualHash"
            return
        }
        Write-Host 'Checksum verified.' -ForegroundColor Green
    } else {
        Write-Warning 'Could not find checksum entry, skipping verification'
    }

    # Extract
    Write-Host 'Extracting...'
    Expand-Archive -Path (Join-Path $TmpDir $Archive) -DestinationPath $TmpDir -Force

    # Install
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item (Join-Path $TmpDir 'vivechak.exe') -Destination (Join-Path $InstallDir 'vivechak.exe') -Force

    # Add to PATH if not already there
    $UserPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
    if ($UserPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable('PATH', "$InstallDir;$UserPath", 'User')
        Write-Host "Added $InstallDir to user PATH" -ForegroundColor Yellow
        Write-Host 'Note: restart your terminal for PATH changes to take effect' -ForegroundColor Yellow
    }

    # Also update current session
    $env:PATH = "$InstallDir;$env:PATH"

    Write-Host ''
    Write-Host "vivechak $Tag installed to $InstallDir\vivechak.exe" -ForegroundColor Green
    Write-Host ''
    Write-Host 'Next step: configure your AI host:' -ForegroundColor Cyan
    Write-Host '  vivechak mcp-config --client cursor --write'
    Write-Host ''
} finally {
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
