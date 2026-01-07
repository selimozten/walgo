# Walgo Installation Script for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$REPO = "selimozten/walgo"
$BINARY_NAME = "walgo"
$TEMP_DIR = [IO.Path]::GetTempPath()

# Color handling
$useColors = $true
try {
    $null = $host.UI.RawUI.ForegroundColor
} catch {
    $useColors = $false
}

if ($useColors) {
    $colorSuccess = "Green"
    $colorError = "Red"
    $colorWarning = "Yellow"
    $colorInfo = "Cyan"
} else {
    $colorSuccess = $colorError = $colorWarning = $colorInfo = $null
}

function Print-Success {
    param([string]$Message)
    if ($useColors) {
        Write-Host "✓ $Message" -ForegroundColor $colorSuccess
    } else {
        Write-Host "[OK] $Message"
    }
}

function Print-Error {
    param([string]$Message)
    if ($useColors) {
        Write-Host "✗ $Message" -ForegroundColor $colorError
    } else {
        Write-Host "[ERROR] $Message"
    }
}

function Print-Warning {
    param([string]$Message)
    if ($useColors) {
        Write-Host "⚠ $Message" -ForegroundColor $colorWarning
    } else {
        Write-Host "[WARN] $Message"
    }
}

function Print-Info {
    param([string]$Message)
    if ($useColors) {
        Write-Host "ℹ $Message" -ForegroundColor $colorInfo
    } else {
        Write-Host "[INFO] $Message"
    }
}

# Helpers
$script:PathUpdates = [System.Collections.Generic.List[string]]::new()
$script:IsAdmin = $false
$script:WalgoInstallRoot = $null
$script:WalgoBinaryPath = $null
$script:WalgoToolsDir = Join-Path $env:USERPROFILE ".walgo\bin"
$script:LocalBinDir = Join-Path $env:USERPROFILE ".local\bin"
$script:DesktopInstallDir = Join-Path $env:LOCALAPPDATA "Programs\Walgo"

function Ensure-Directory {
    param([Parameter(Mandatory = $true)][string]$Path)
    if (-not [string]::IsNullOrWhiteSpace($Path)) {
        New-Item -ItemType Directory -Force -Path $Path | Out-Null
    }
}

function Add-PathEntry {
    param(
        [Parameter(Mandatory = $true)][string]$PathValue,
        [switch]$Machine
    )

    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return
    }

    Ensure-Directory $PathValue

    $resolved = $PathValue
    try {
        $resolved = (Resolve-Path -Path $PathValue).Path
    } catch {
        # ignore
    }

    $target = [EnvironmentVariableTarget]::User
    if ($Machine -and $script:IsAdmin) {
        $target = [EnvironmentVariableTarget]::Machine
    } elseif ($Machine -and -not $script:IsAdmin) {
        Print-Warning "Administrator privileges required for system PATH. Added to user PATH instead."
    }

    $existing = [Environment]::GetEnvironmentVariable("Path", $target)
    if ([string]::IsNullOrWhiteSpace($existing)) {
        $existingList = @()
    } else {
        $existingList = $existing -split ';' | Where-Object { $_ }
    }

    if ($existingList -notcontains $resolved) {
        $newList = $existingList + $resolved
        $newPath = ($newList -join ';').Trim(';')
        [Environment]::SetEnvironmentVariable("Path", $newPath, $target)
        if (-not $script:PathUpdates.Contains($resolved)) {
            $script:PathUpdates.Add($resolved)
        }
        Print-Success "Added $resolved to PATH"
    } else {
        Print-Info "$resolved already in PATH"
    }

    if (($env:Path -split ';') -notcontains $resolved) {
        $env:Path = ($env:Path.TrimEnd(';') + ';' + $resolved).Trim(';')
    }
}

function Download-File {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Destination,
        [string]$Description
    )

    $label = if ($Description) { $Description } else { $Url }
    Print-Info "Downloading $label..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing
        return $true
    } catch {
        Print-Error "Failed to download $label: $_"
        return $false
    }
}

function Expand-ArchiveSafe {
    param(
        [Parameter(Mandatory = $true)][string]$Archive,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    try {
        Expand-Archive -Path $Archive -DestinationPath $Destination -Force
        return $true
    } catch {
        Print-Error "Failed to extract archive: $_"
        return $false
    }
}

function New-TempDirectory {
    param([string]$Suffix)
    $folder = Join-Path $TEMP_DIR ("walgo-" + $Suffix + "-" + [Guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Force -Path $folder | Out-Null
    return $folder
}

# Platform helpers
function Get-Architecture {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Print-Error "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
            exit 1
        }
    }
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = [Security.Principal.WindowsPrincipal]::new($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Initialize-Environment {
    $script:IsAdmin = Test-Administrator
    if ($script:IsAdmin) {
        $script:WalgoInstallRoot = "C:\\Program Files\\Walgo"
    } else {
        $script:WalgoInstallRoot = Join-Path $env:LOCALAPPDATA "Walgo"
    }

    $script:WalgoBinaryPath = Join-Path $script:WalgoInstallRoot "walgo.exe"

    Ensure-Directory $script:WalgoInstallRoot
    Ensure-Directory $script:WalgoToolsDir
    Ensure-Directory $script:LocalBinDir
    Ensure-Directory $script:DesktopInstallDir
}

# GitHub release helper
function Get-LatestRelease {
    param([Parameter(Mandatory = $true)][string]$Repository)
    try {
        return Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest" -UseBasicParsing
    } catch {
        Print-Error "Failed to fetch latest release for $Repository: $_"
        return $null
    }
}

function Get-LatestVersion {
    Print-Info "Fetching latest Walgo version..."
    $release = Get-LatestRelease -Repository $REPO
    if (-not $release) {
        exit 1
    }
    $version = $release.tag_name -replace '^v', ''
    Print-Success "Latest version: $version"
    return $version
}

function Install-WalgoBinary {
    param([Parameter(Mandatory = $true)][string]$Version)

    $arch = Get-Architecture
    $filename = "${BINARY_NAME}_${Version}_windows_${arch}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/v$Version/$filename"
    $tempFile = Join-Path $TEMP_DIR $filename

    if (-not (Download-File -Url $downloadUrl -Destination $tempFile -Description "$BINARY_NAME CLI")) {
        exit 1
    }

    $extractDir = New-TempDirectory "cli"
    if (-not (Expand-ArchiveSafe -Archive $tempFile -Destination $extractDir)) {
        exit 1
    }

    $binary = Get-ChildItem -Path $extractDir -Filter "${BINARY_NAME}.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Print-Error "walgo.exe not found in archive"
        exit 1
    }

    Copy-Item -Path $binary.FullName -Destination $script:WalgoBinaryPath -Force
    Print-Success "Installed $BINARY_NAME to $script:WalgoBinaryPath"

    if ($script:IsAdmin) {
        Add-PathEntry -PathValue $script:WalgoInstallRoot -Machine
    } else {
        Add-PathEntry -PathValue $script:WalgoInstallRoot
    }

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
}

function Install-WalgoDesktop {
    param([Parameter(Mandatory = $true)][string]$Version)

    $arch = Get-Architecture
    $filename = "walgo-desktop_${Version}_windows_${arch}.zip"
    $downloadUrl = "https://github.com/$REPO/releases/download/v$Version/$filename"
    $tempFile = Join-Path $TEMP_DIR $filename

    if (-not (Download-File -Url $downloadUrl -Destination $tempFile -Description "Walgo Desktop")) {
        Print-Warning "Skipping desktop installation"
        return
    }

    $extractDir = New-TempDirectory "desktop"
    if (-not (Expand-ArchiveSafe -Archive $tempFile -Destination $extractDir)) {
        Print-Warning "Failed to extract desktop app"
        return
    }

    $exe = Get-ChildItem -Path $extractDir -Filter "Walgo*.exe" -Recurse | Select-Object -First 1
    if (-not $exe) {
        Print-Warning "Desktop executable not found in archive"
        return
    }

    Ensure-Directory $script:DesktopInstallDir
    $destination = Join-Path $script:DesktopInstallDir "Walgo.exe"
    Copy-Item -Path $exe.FullName -Destination $destination -Force
    Print-Success "Walgo Desktop installed to $destination"
    Print-Info "Launch via: walgo desktop"

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
}

function Test-Installation {
    Print-Info "Verifying Walgo CLI..."
    try {
        $versionOutput = & $script:WalgoBinaryPath --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Print-Success "walgo --version"
            Write-Host $versionOutput
        } else {
            Print-Warning "Walgo installed but version check failed. Try restarting PowerShell."
        }
    } catch {
        Print-Warning "Walgo installed but not yet available in this session. Restart PowerShell."
    }
}

function Install-HugoDirect {
    Print-Info "Fetching latest Hugo version..."
    $release = Get-LatestRelease -Repository "gohugoio/hugo"
    if (-not $release) {
        return $false
    }

    $version = $release.tag_name -replace '^v', ''
    Print-Info "Latest Hugo version: $version"

    $arch = Get-Architecture
    $filename = "hugo_extended_${version}_windows-${arch}.zip"
    $downloadUrl = "https://github.com/gohugoio/hugo/releases/download/v$version/$filename"
    $tempFile = Join-Path $TEMP_DIR $filename

    if (-not (Download-File -Url $downloadUrl -Destination $tempFile -Description "Hugo")) {
        return $false
    }

    $extractDir = New-TempDirectory "hugo"
    if (-not (Expand-ArchiveSafe -Archive $tempFile -Destination $extractDir)) {
        return $false
    }

    $binary = Get-ChildItem -Path $extractDir -Filter "hugo.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Print-Error "hugo.exe not found in archive"
        return $false
    }

    $destination = Join-Path $script:WalgoToolsDir "hugo.exe"
    Copy-Item -Path $binary.FullName -Destination $destination -Force
    Add-PathEntry -PathValue $script:WalgoToolsDir
    Print-Success "Hugo installed to $destination"

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    return $true
}

function Check-Dependencies {
    Print-Info "Checking optional dependencies..."
    $hugo = Get-Command hugo.exe -ErrorAction SilentlyContinue
    if ($null -eq $hugo) {
        Print-Warning "Hugo not found. Installing from GitHub releases..."
        if (-not (Install-HugoDirect)) {
            Print-Warning "Hugo installation failed. Install manually from https://gohugo.io/installation/"
        }
    } else {
        Print-Success "Hugo found: $($hugo.Source)"
    }
}

function Install-Suiup {
    $target = Join-Path $script:WalgoToolsDir "suiup.exe"
    if (Test-Path $target) {
        return $target
    }

    Print-Info "Installing suiup..."
    $release = Get-LatestRelease -Repository "MystenLabs/suiup"
    if (-not $release) {
        return $null
    }

    $tag = $release.tag_name
    $arch = Get-Architecture
    $filename = "suiup-windows-${arch}.zip"
    $downloadUrl = "https://github.com/MystenLabs/suiup/releases/download/$tag/$filename"
    $tempFile = Join-Path $TEMP_DIR $filename

    if (-not (Download-File -Url $downloadUrl -Destination $tempFile -Description "suiup")) {
        return $null
    }

    $extractDir = New-TempDirectory "suiup"
    if (-not (Expand-ArchiveSafe -Archive $tempFile -Destination $extractDir)) {
        return $null
    }

    $binary = Get-ChildItem -Path $extractDir -Filter "suiup.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Print-Warning "suiup.exe not found in archive"
        return $null
    }

    Copy-Item -Path $binary.FullName -Destination $target -Force
    Add-PathEntry -PathValue $script:WalgoToolsDir
    Add-PathEntry -PathValue $script:LocalBinDir
    Print-Success "suiup installed to $target"

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    return $target
}

function Download-WalrusConfigs {
    $configDir = Join-Path $env:USERPROFILE ".config\walrus"
    Ensure-Directory $configDir

    try {
        Invoke-WebRequest -Uri "https://docs.wal.app/setup/client_config.yaml" -OutFile (Join-Path $configDir "client_config.yaml") -UseBasicParsing
        Print-Success "Walrus client config downloaded"
    } catch {
        Print-Warning "Failed to download Walrus client config"
    }

    try {
        Invoke-WebRequest -Uri "https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml" -OutFile (Join-Path $configDir "sites-config.yaml") -UseBasicParsing
        Print-Success "site-builder config downloaded"
    } catch {
        Print-Warning "Failed to download site-builder config"
    }
}

function Install-WalrusDependencies {
    Print-Info "Would you like to install Walrus dependencies (Sui CLI, Walrus CLI, site-builder)? [y/N]"
    $response = Read-Host
    if ($response -notmatch '^[Yy]$') {
        Print-Info "Skipping Walrus dependencies"
        return
    }

    $suiupPath = Install-Suiup
    if (-not $suiupPath) {
        Print-Warning "suiup installation failed. Install dependencies manually later using 'walgo setup-deps'"
        return
    }

    $targets = @(
        @{ Label = "Sui CLI (testnet)"; Spec = "sui@testnet" },
        @{ Label = "Sui CLI (mainnet)"; Spec = "sui@mainnet" },
        @{ Label = "Walrus CLI (testnet)"; Spec = "walrus@testnet" },
        @{ Label = "Walrus CLI (mainnet)"; Spec = "walrus@mainnet" },
        @{ Label = "site-builder"; Spec = "site-builder@mainnet" }
    )

    foreach ($target in $targets) {
        Print-Info "Installing $($target.Label)..."
        try {
            & $suiupPath install $($target.Spec) | Out-Null
            Print-Success "$($target.Label) installed"
        } catch {
            Print-Warning "Failed to install $($target.Label): $_"
        }
    }

    try { & $suiupPath default set "sui@testnet" | Out-Null } catch { }
    try { & $suiupPath default set "walrus@testnet" | Out-Null } catch { }
    try { & $suiupPath default set "site-builder@mainnet" | Out-Null } catch { }

    $binaries = @(
        @{ Name = "Sui CLI"; Path = Join-Path $script:LocalBinDir "sui.exe" },
        @{ Name = "Walrus CLI"; Path = Join-Path $script:LocalBinDir "walrus.exe" },
        @{ Name = "site-builder"; Path = Join-Path $script:LocalBinDir "site-builder.exe" }
    )

    foreach ($binary in $binaries) {
        if (Test-Path $binary.Path) {
            try {
                $output = & $binary.Path --version 2>&1 | Select-Object -First 1
                Print-Success "$($binary.Name) ready: $output"
            } catch {
                Print-Info "$($binary.Name) installed"
            }
        } else {
            Print-Warning "$($binary.Name) not found in $script:LocalBinDir. Check suiup output for details."
        }
    }

    Download-WalrusConfigs
    Print-Success "Walrus dependencies installation attempted. Verify with: sui --version"
}

function Show-NextSteps {
    Write-Host ""
    Write-Host "═══════════════════════════════════════════════════════════"
    Print-Success "Walgo installation complete!"
    Write-Host "═══════════════════════════════════════════════════════════"
    Write-Host ""

    Print-Warning "Restart PowerShell (or your terminal) to load updated PATH settings."
    if ($script:PathUpdates.Count -gt 0) {
        Print-Info "PATH updated with:"
        foreach ($entry in $script:PathUpdates) {
            Write-Host "  - $entry"
        }
    }

    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  1. Verify installation:"
    Write-Host "     walgo --help"
    Write-Host ""
    Write-Host "  2. Create your first site:"
    Write-Host "     walgo init my-site"
    Write-Host "     cd my-site"
    Write-Host ""
    Write-Host "  3. Build your site:"
    Write-Host "     walgo build"
    Write-Host ""
    Write-Host "  4. Deploy with the interactive wizard:"
    Write-Host "     walgo launch"
    Write-Host ""
    Write-Host "  5. To launch the desktop app:"
    Write-Host "     walgo desktop"
    Write-Host "     (Desktop installed to $script:DesktopInstallDir\Walgo.exe)"
    Write-Host ""
    Write-Host "  6. If you installed Walrus dependencies:"
    Write-Host "     sui --version"
    Write-Host "     walrus --version"
    Write-Host "     site-builder --version"
    Write-Host ""
    Write-Host "Documentation: https://github.com/$REPO"
    Write-Host ""
}

function Main {
    Write-Host ""
    Write-Host "╔═══════════════════════════════════════════════════════════╗"
    Write-Host "║                   Walgo Installer                         ║"
    Write-Host "║    Ship static sites to Walrus decentralized storage     ║"
    Write-Host "╚═══════════════════════════════════════════════════════════╝"
    Write-Host ""

    Initialize-Environment
    $version = Get-LatestVersion
    Install-WalgoBinary -Version $version
    Install-WalgoDesktop -Version $version
    Test-Installation
    Check-Dependencies
    Install-WalrusDependencies
    Show-NextSteps
}

Main

