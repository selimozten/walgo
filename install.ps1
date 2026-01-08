# Walgo Installation Script for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$REPO = "selimozten/walgo"
$BINARY_NAME = "walgo"
$TEMP_DIR = [IO.Path]::GetTempPath()

# Global variables to store wallet info for display at end
$script:SuiMnemonicPhrase = ""
$script:SuiAddress = ""

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
        # Ensure destination directory exists
        $destDir = Split-Path -Parent $Destination
        if ($destDir -and -not (Test-Path $destDir)) {
            New-Item -ItemType Directory -Force -Path $destDir | Out-Null
        }

        Invoke-WebRequest -Uri $Url -OutFile $Destination -UseBasicParsing -ErrorAction Stop

        # Verify file was created and has content
        if (-not (Test-Path $Destination)) {
            Print-Error "Download completed but file not found at $Destination"
            return $false
        }

        $fileInfo = Get-Item $Destination
        if ($fileInfo.Length -eq 0) {
            Print-Error "Downloaded file is empty (0 bytes)"
            Remove-Item $Destination -Force -ErrorAction SilentlyContinue
            return $false
        }

        return $true
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to download $label - $errorMsg"
        Print-Info "Please check your internet connection and try again"
        return $false
    }
}

function Expand-ArchiveSafe {
    param(
        [Parameter(Mandatory = $true)][string]$Archive,
        [Parameter(Mandatory = $true)][string]$Destination
    )

    try {
        # Verify archive exists and is readable
        if (-not (Test-Path $Archive)) {
            Print-Error "Archive file not found: $Archive"
            return $false
        }

        $archiveInfo = Get-Item $Archive
        if ($archiveInfo.Length -eq 0) {
            Print-Error "Archive file is empty (0 bytes)"
            return $false
        }

        # Ensure destination directory exists
        if (-not (Test-Path $Destination)) {
            New-Item -ItemType Directory -Force -Path $Destination | Out-Null
        }

        Expand-Archive -Path $Archive -DestinationPath $Destination -Force -ErrorAction Stop

        # Verify extraction succeeded
        $extractedItems = Get-ChildItem -Path $Destination -ErrorAction SilentlyContinue
        if ($null -eq $extractedItems -or $extractedItems.Count -eq 0) {
            Print-Error "Archive extraction completed but no files found"
            return $false
        }

        return $true
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to extract archive - $errorMsg"
        Print-Info "The archive file may be corrupted. Please try again."
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
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to fetch latest release for $Repository - $errorMsg"
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
        Print-Info "The downloaded archive may be corrupted. Please try again."
        exit 1
    }

    try {
        # Ensure parent directory exists
        $parentDir = Split-Path -Parent $script:WalgoBinaryPath
        if (-not (Test-Path $parentDir)) {
            New-Item -ItemType Directory -Force -Path $parentDir | Out-Null
        }

        Copy-Item -Path $binary.FullName -Destination $script:WalgoBinaryPath -Force -ErrorAction Stop
        Print-Success "Installed $BINARY_NAME to $script:WalgoBinaryPath"
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to copy binary to $script:WalgoBinaryPath - $errorMsg"
        Print-Info "Check if you have write permissions to the directory"
        exit 1
    }

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

    try {
        Ensure-Directory $script:DesktopInstallDir
        $destination = Join-Path $script:DesktopInstallDir "Walgo.exe"
        Copy-Item -Path $exe.FullName -Destination $destination -Force -ErrorAction Stop

        # Verify copy succeeded
        if (-not (Test-Path $destination)) {
            Print-Warning "Desktop installation completed but file not found at $destination"
            return
        }

        Print-Success "Walgo Desktop installed to $destination"
        Print-Info "Launch via: walgo desktop"
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to install desktop app - $errorMsg"
        Print-Info "You can try installing it manually later"
    }

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

    try {
        Ensure-Directory $script:WalgoToolsDir
        $destination = Join-Path $script:WalgoToolsDir "hugo.exe"
        Copy-Item -Path $binary.FullName -Destination $destination -Force -ErrorAction Stop

        # Verify copy succeeded
        if (-not (Test-Path $destination)) {
            Print-Error "Hugo installation completed but file not found at $destination"
            return $false
        }

        Add-PathEntry -PathValue $script:WalgoToolsDir
        Print-Success "Hugo installed to $destination"
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to install Hugo - $errorMsg"
        Print-Info "Install manually from https://gohugo.io/installation/"
        return $false
    }

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    return $true
}

function Check-Dependencies {
    Print-Info "Checking optional dependencies..."

    # Check for Hugo
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

    try {
        Ensure-Directory $script:WalgoToolsDir
        Copy-Item -Path $binary.FullName -Destination $target -Force -ErrorAction Stop

        # Verify copy succeeded
        if (-not (Test-Path $target)) {
            Print-Warning "suiup installation completed but file not found at $target"
            return $null
        }

        Add-PathEntry -PathValue $script:WalgoToolsDir
        Add-PathEntry -PathValue $script:LocalBinDir
        Print-Success "suiup installed to $target"
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to install suiup - $errorMsg"
        return $null
    }

    Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
    return $target
}

function Download-WalrusConfigs {
    $configDir = Join-Path $env:USERPROFILE ".config\walrus"
    Ensure-Directory $configDir

    $clientConfigPath = Join-Path $configDir "client_config.yaml"
    try {
        Invoke-WebRequest -Uri "https://docs.wal.app/setup/client_config.yaml" -OutFile $clientConfigPath -UseBasicParsing -ErrorAction Stop

        # Verify download
        if ((Test-Path $clientConfigPath) -and (Get-Item $clientConfigPath).Length -gt 0) {
            Print-Success "Walrus client config downloaded"
        } else {
            Print-Warning "Failed to download Walrus client config (file empty or missing)"
        }
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to download Walrus client config - $errorMsg"
        Print-Info "You can download it manually from: https://docs.wal.app/setup/client_config.yaml"
    }

    $sitesConfigPath = Join-Path $configDir "sites-config.yaml"
    try {
        Invoke-WebRequest -Uri "https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml" -OutFile $sitesConfigPath -UseBasicParsing -ErrorAction Stop

        # Verify download
        if ((Test-Path $sitesConfigPath) -and (Get-Item $sitesConfigPath).Length -gt 0) {
            Print-Success "site-builder config downloaded"
        } else {
            Print-Warning "Failed to download site-builder config (file empty or missing)"
        }
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to download site-builder config - $errorMsg"
        Print-Info "You can download it manually from: https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml"
    }
}

function Initialize-SuiClient {
    param([string]$Network = "testnet")

    Print-Info "Configuring Sui client..."

    $suiConfigPath = Join-Path $env:USERPROFILE ".sui\sui_config\client.yaml"

    if (Test-Path $suiConfigPath) {
        Print-Success "Sui client already configured"

        # Even if already configured, ensure both environments exist
        Print-Info "Verifying testnet and mainnet environments..."
        $suiPath = Join-Path $script:LocalBinDir "sui.exe"
        if (Test-Path $suiPath) {
            # Add mainnet if missing
            try {
                $envs = & $suiPath client envs 2>$null
                if ($envs -notmatch "mainnet") {
                    $null = & $suiPath client new-env --alias mainnet --rpc https://fullnode.mainnet.sui.io:443 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        Print-Success "Added mainnet environment"
                    }
                }
            } catch { }

            # Add testnet if missing
            try {
                $envs = & $suiPath client envs 2>$null
                if ($envs -notmatch "testnet") {
                    $null = & $suiPath client new-env --alias testnet --rpc https://fullnode.testnet.sui.io:443 2>$null
                    if ($LASTEXITCODE -eq 0) {
                        Print-Success "Added testnet environment"
                    }
                }
            } catch { }
        }

        return
    }

    Print-Info "Initializing Sui client for $Network..."
    Print-Info "Creating new Sui address..."

    $suiPath = Join-Path $script:LocalBinDir "sui.exe"
    if (-not (Test-Path $suiPath)) {
        Print-Warning "Sui CLI not found. Skipping client initialization."
        return
    }

    try {
        # Prepare input for sui client initialization
        $fullNodeUrl = "https://fullnode.$Network.sui.io:443"
        $inputs = @(
            "y",           # Connect to Sui Full Node
            $fullNodeUrl,  # Full node URL
            $Network,      # Environment alias
            "0"            # Key scheme (ed25519)
        )

        # Run sui client with timeout and capture output
        $job = Start-Job -ScriptBlock {
            param($suiPath, $inputs)
            $inputString = $inputs -join "`n"
            $output = $inputString | & $suiPath client 2>&1
            return $output -join "`n"
        } -ArgumentList $suiPath, $inputs

        $completed = Wait-Job $job -Timeout 30
        if ($completed) {
            $suiOutput = Receive-Job $job
            Remove-Job $job -Force

            # Extract recovery phrase from JSON output
            if ($suiOutput -match '"recoveryPhrase"\s*:\s*"([^"]*)"') {
                $script:SuiMnemonicPhrase = $matches[1]
            }

            # Extract address from JSON output
            if ($suiOutput -match '"address"\s*:\s*"([^"]*)"') {
                $script:SuiAddress = $matches[1]
            }

            # Alternative: try to extract from non-JSON format if JSON extraction failed
            if (-not $script:SuiMnemonicPhrase) {
                # Try to find lines with 12 or 24 words (typical mnemonic formats)
                if ($suiOutput -match '([a-z]+\s+){11,23}[a-z]+') {
                    $script:SuiMnemonicPhrase = $matches[0].Trim()
                }
            }

            if (Test-Path $suiConfigPath) {
                Print-Success "Sui client configured successfully"
                if ($script:SuiMnemonicPhrase) {
                    Print-Success "New wallet address created"
                }

                # Add both testnet and mainnet environments
                Print-Info "Configuring both testnet and mainnet environments..."

                $suiPath = Join-Path $script:LocalBinDir "sui.exe"
                if (Test-Path $suiPath) {
                    # Add mainnet if default was testnet
                    if ($Network -eq "testnet") {
                        try {
                            $envs = & $suiPath client envs 2>$null
                            if ($envs -notmatch "mainnet") {
                                $null = & $suiPath client new-env --alias mainnet --rpc https://fullnode.mainnet.sui.io:443 2>$null
                                if ($LASTEXITCODE -eq 0) {
                                    Print-Success "Added mainnet environment"
                                } else {
                                    Print-Warning "Failed to add mainnet environment (you can add it later)"
                                }
                            } else {
                                Print-Info "Mainnet environment already exists"
                            }
                        } catch {
                            $errorMsg = $_.Exception.Message
                            Print-Warning "Failed to add mainnet environment - $errorMsg"
                        }
                    }

                    # Add testnet if default was mainnet
                    if ($Network -eq "mainnet") {
                        try {
                            $envs = & $suiPath client envs 2>$null
                            if ($envs -notmatch "testnet") {
                                $null = & $suiPath client new-env --alias testnet --rpc https://fullnode.testnet.sui.io:443 2>$null
                                if ($LASTEXITCODE -eq 0) {
                                    Print-Success "Added testnet environment"
                                } else {
                                    Print-Warning "Failed to add testnet environment (you can add it later)"
                                }
                            } else {
                                Print-Info "Testnet environment already exists"
                            }
                        } catch {
                            $errorMsg = $_.Exception.Message
                            Print-Warning "Failed to add testnet environment - $errorMsg"
                        }
                    }
                }
            } else {
                Print-Warning "Sui client initialization may have failed"
                Print-Info "You can configure it manually later with: sui client"
            }
        } else {
            Stop-Job $job
            Remove-Job $job -Force
            Print-Warning "Sui client initialization timed out after 30 seconds"
            Print-Info "You can configure it manually later with: sui client"
        }
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to initialize Sui client - $errorMsg"
        Print-Info "You can configure it manually later with: sui client"
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
            # Run with 120 second timeout
            $job = Start-Job -ScriptBlock {
                param($suiupPath, $spec)
                & $suiupPath install $spec 2>&1
            } -ArgumentList $suiupPath, $($target.Spec)

            $completed = Wait-Job $job -Timeout 120
            if ($completed) {
                $output = Receive-Job $job
                Remove-Job $job -Force
                Print-Success "$($target.Label) installed"
            } else {
                Stop-Job $job
                Remove-Job $job -Force
                Print-Warning "$($target.Label) installation timed out after 120 seconds"
            }
        } catch {
            $errorMsg = $_.Exception.Message
            Print-Warning "Failed to install $($target.Label) - $errorMsg"
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
    Initialize-SuiClient -Network "testnet"
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

    # Display mnemonic phrase if a new address was created
    if ($script:SuiMnemonicPhrase) {
        Write-Host ""
        Write-Host "═══════════════════════════════════════════════════════════"
        if ($useColors) {
            Write-Host "║  IMPORTANT: SAVE YOUR WALLET RECOVERY PHRASE" -ForegroundColor $colorError
        } else {
            Write-Host "║  IMPORTANT: SAVE YOUR WALLET RECOVERY PHRASE"
        }
        Write-Host "═══════════════════════════════════════════════════════════"
        Write-Host ""

        if ($script:SuiAddress) {
            if ($useColors) {
                Write-Host "  Your Sui Address:" -ForegroundColor $colorWarning
                Write-Host "  $script:SuiAddress" -ForegroundColor $colorSuccess
            } else {
                Write-Host "  Your Sui Address:"
                Write-Host "  $script:SuiAddress"
            }
            Write-Host ""
        }

        if ($useColors) {
            Write-Host "  Secret Recovery Phrase:" -ForegroundColor $colorWarning
            Write-Host "  $script:SuiMnemonicPhrase" -ForegroundColor $colorSuccess
        } else {
            Write-Host "  Secret Recovery Phrase:"
            Write-Host "  $script:SuiMnemonicPhrase"
        }
        Write-Host ""

        if ($useColors) {
            Write-Host "  ⚠️  Write this down and store it safely!" -ForegroundColor $colorError
            Write-Host "  ⚠️  You will need this phrase to recover your wallet." -ForegroundColor $colorError
            Write-Host "  ⚠️  Never share this phrase with anyone!" -ForegroundColor $colorError
        } else {
            Write-Host "  ⚠️  Write this down and store it safely!"
            Write-Host "  ⚠️  You will need this phrase to recover your wallet."
            Write-Host "  ⚠️  Never share this phrase with anyone!"
        }
        Write-Host ""

        if (-not $script:SuiAddress) {
            Write-Host "  Get your address: sui client active-address"
            Write-Host ""
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

