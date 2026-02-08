# Walgo Installation Script for Windows PowerShell
# Usage: irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Enable TLS 1.2 for HTTPS connections (required for older Windows/PowerShell)
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$REPO = "selimozten/walgo"
$BINARY_NAME = "walgo"
$TEMP_DIR = [IO.Path]::GetTempPath()

# Global variables to store wallet info for display at end
$script:SuiMnemonicPhrase = ""
$script:SuiAddress = ""
$script:TempDirsToCleanup = [System.Collections.Generic.List[string]]::new()

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

# Cleanup function for temp directories
function Cleanup-TempDirs {
    foreach ($dir in $script:TempDirsToCleanup) {
        if (Test-Path $dir) {
            Remove-Item -Path $dir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Register cleanup on script exit
$null = Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Cleanup-TempDirs } -ErrorAction SilentlyContinue

# Helpers
$script:PathUpdates = [System.Collections.Generic.List[string]]::new()
$script:IsAdmin = $false
$script:WalgoInstallRoot = $null
$script:WalgoBinaryPath = $null
$script:WalgoToolsDir = Join-Path $env:USERPROFILE ".walgo\bin"
$script:SuiBinDir = Join-Path $env:USERPROFILE ".sui\bin"  # Suiup's actual install location
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
        # Split, trim whitespace, and filter empty entries
        $existingList = $existing -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    }

    # Case-insensitive comparison for Windows paths
    $alreadyExists = $existingList | Where-Object { $_.ToLower() -eq $resolved.ToLower() }

    if (-not $alreadyExists) {
        $newList = $existingList + $resolved
        $newPath = ($newList -join ';').Trim(';')
        [Environment]::SetEnvironmentVariable("Path", $newPath, $target)
        if (-not ($script:PathUpdates | Where-Object { $_.ToLower() -eq $resolved.ToLower() })) {
            $script:PathUpdates.Add($resolved)
        }
        Print-Success "Added $resolved to PATH"
    } else {
        Print-Info "$resolved already in PATH"
    }

    # Also update current session PATH (case-insensitive check)
    $currentPaths = $env:Path -split ';' | ForEach-Object { $_.Trim() } | Where-Object { $_ }
    $inCurrentPath = $currentPaths | Where-Object { $_.ToLower() -eq $resolved.ToLower() }
    if (-not $inCurrentPath) {
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

function Expand-TarGz {
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

        # Check if tar command is available (Windows 10 build 17063+ has built-in tar)
        $tarCommand = Get-Command tar.exe -ErrorAction SilentlyContinue
        if ($tarCommand) {
            # Use built-in tar command
            $result = & tar.exe -xzf $Archive -C $Destination 2>&1
            if ($LASTEXITCODE -ne 0) {
                Print-Error "tar extraction failed: $result"
                return $false
            }
        } else {
            # Fallback: Use .NET to decompress gzip, then use 7-zip or fail gracefully
            Print-Info "tar.exe not available, using .NET gzip decompression..."

            # First decompress .gz to get .tar
            $tarFile = $Archive -replace '\.gz$', ''
            try {
                $gzStream = New-Object System.IO.FileStream($Archive, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read)
                $tarStream = New-Object System.IO.FileStream($tarFile, [System.IO.FileMode]::Create, [System.IO.FileAccess]::Write)
                $gzipStream = New-Object System.IO.Compression.GzipStream($gzStream, [System.IO.Compression.CompressionMode]::Decompress)

                $gzipStream.CopyTo($tarStream)
                $gzipStream.Close()
                $tarStream.Close()
                $gzStream.Close()
            } catch {
                $errorMsg = $_.Exception.Message
                Print-Error "Failed to decompress gzip - $errorMsg"
                return $false
            }

            # Try to find 7-zip for tar extraction
            $sevenZipPaths = @(
                "C:\Program Files\7-Zip\7z.exe",
                "C:\Program Files (x86)\7-Zip\7z.exe",
                (Join-Path $env:LOCALAPPDATA "Programs\7-Zip\7z.exe")
            )

            $sevenZip = $null
            foreach ($path in $sevenZipPaths) {
                if (Test-Path $path) {
                    $sevenZip = $path
                    break
                }
            }

            if ($sevenZip) {
                Print-Info "Using 7-Zip for tar extraction..."
                $result = & $sevenZip x $tarFile -o"$Destination" -y 2>&1
                Remove-Item $tarFile -Force -ErrorAction SilentlyContinue
                if ($LASTEXITCODE -ne 0) {
                    Print-Error "7-Zip extraction failed: $result"
                    return $false
                }
            } else {
                Remove-Item $tarFile -Force -ErrorAction SilentlyContinue
                Print-Error "Cannot extract .tar.gz: tar.exe not found and 7-Zip not installed."
                Print-Info "Please install Windows 10/11 (has built-in tar) or install 7-Zip."
                Print-Info "Or download the .zip version manually from GitHub releases."
                return $false
            }
        }

        # Verify extraction succeeded
        $extractedItems = Get-ChildItem -Path $Destination -ErrorAction SilentlyContinue
        if ($null -eq $extractedItems -or $extractedItems.Count -eq 0) {
            Print-Error "Archive extraction completed but no files found"
            return $false
        }

        return $true
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Error "Failed to extract tar.gz archive - $errorMsg"
        Print-Info "The archive file may be corrupted. Please try again."
        return $false
    }
}

function New-TempDirectory {
    param([string]$Suffix)
    $folder = Join-Path $TEMP_DIR ("walgo-" + $Suffix + "-" + [Guid]::NewGuid().ToString("N").Substring(0, 8))
    New-Item -ItemType Directory -Force -Path $folder | Out-Null
    $script:TempDirsToCleanup.Add($folder)
    return $folder
}

# Platform helpers
function Get-Architecture {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Print-Error "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
            Cleanup-TempDirs
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
        $script:WalgoInstallRoot = "C:\Program Files\Walgo"
    } else {
        $script:WalgoInstallRoot = Join-Path $env:LOCALAPPDATA "Walgo"
    }

    $script:WalgoBinaryPath = Join-Path $script:WalgoInstallRoot "walgo.exe"

    Ensure-Directory $script:WalgoInstallRoot
    Ensure-Directory $script:WalgoToolsDir
    Ensure-Directory $script:SuiBinDir
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
        Cleanup-TempDirs
        exit 1
    }
    $version = $release.tag_name -replace '^v', ''
    Print-Success "Latest version: $version"
    return $version
}

function Install-WalgoBinary {
    param([Parameter(Mandatory = $true)][string]$Version)

    $arch = Get-Architecture
    $extractDir = New-TempDirectory "cli"

    # Try .zip first (newer releases), then .tar.gz (v0.3.2 and earlier)
    $formats = @(
        @{ Ext = "zip"; Func = "Expand-ArchiveSafe" },
        @{ Ext = "tar.gz"; Func = "Expand-TarGz" }
    )

    $success = $false
    $tempFile = $null
    foreach ($format in $formats) {
        $filename = "${BINARY_NAME}_${Version}_windows_${arch}.$($format.Ext)"
        $downloadUrl = "https://github.com/$REPO/releases/download/v$Version/$filename"
        $tempFile = Join-Path $TEMP_DIR $filename

        Print-Info "Trying to download $filename..."
        if (Download-File -Url $downloadUrl -Destination $tempFile -Description "$BINARY_NAME CLI") {
            Print-Info "Extracting archive..."
            $extractResult = & $format.Func -Archive $tempFile -Destination $extractDir
            if ($extractResult) {
                $success = $true
                Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
                break
            }
        }
        Remove-Item -Path $tempFile -Force -ErrorAction SilentlyContinue
    }

    if (-not $success) {
        Print-Error "Failed to download or extract walgo binary"
        Print-Info "Tried formats: .zip, .tar.gz"
        Cleanup-TempDirs
        exit 1
    }

    $binary = Get-ChildItem -Path $extractDir -Filter "${BINARY_NAME}.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Print-Error "walgo.exe not found in archive"
        Print-Info "The downloaded archive may be corrupted. Please try again."
        Cleanup-TempDirs
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
        Cleanup-TempDirs
        exit 1
    }

    if ($script:IsAdmin) {
        Add-PathEntry -PathValue $script:WalgoInstallRoot -Machine
    } else {
        Add-PathEntry -PathValue $script:WalgoInstallRoot
    }

    Remove-Item -Path $extractDir -Recurse -Force -ErrorAction SilentlyContinue
}

function Create-DesktopShortcut {
    param(
        [Parameter(Mandatory = $true)][string]$TargetPath,
        [Parameter(Mandatory = $true)][string]$ShortcutName,
        [string]$Description = "",
        [string]$IconPath = $null
    )

    try {
        $desktopPath = [Environment]::GetFolderPath("Desktop")
        $shortcutPath = Join-Path $desktopPath "$ShortcutName.lnk"

        $shell = New-Object -ComObject WScript.Shell
        $shortcut = $shell.CreateShortcut($shortcutPath)
        $shortcut.TargetPath = $TargetPath
        $shortcut.Description = $Description
        $shortcut.WorkingDirectory = Split-Path -Parent $TargetPath

        if ($IconPath -and (Test-Path $IconPath)) {
            $shortcut.IconLocation = $IconPath
        } elseif (Test-Path $TargetPath) {
            $shortcut.IconLocation = $TargetPath
        }

        $shortcut.Save()

        # Release COM object
        [System.Runtime.Interopservices.Marshal]::ReleaseComObject($shell) | Out-Null

        return $true
    } catch {
        $errorMsg = $_.Exception.Message
        Print-Warning "Failed to create desktop shortcut - $errorMsg"
        return $false
    }
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

        # Create desktop shortcut
        if (Create-DesktopShortcut -TargetPath $destination -ShortcutName "Walgo" -Description "Walgo - Ship static sites to Walrus") {
            Print-Success "Desktop shortcut created"
        }

        Print-Info "Launch via: walgo desktop (or use the desktop shortcut)"
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

function Test-Dependencies {
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
        Print-Info "suiup already installed at $target"
        return $target
    }

    Print-Info "Installing suiup..."
    $release = Get-LatestRelease -Repository "MystenLabs/suiup"
    if (-not $release) {
        return $null
    }

    $tag = $release.tag_name
    $arch = Get-Architecture

    # Suiup Windows releases use format: suiup-Windows-msvc-x86_64.zip or suiup-Windows-msvc-aarch64.zip
    $suiupArch = switch ($arch) {
        "amd64" { "x86_64" }
        "arm64" { "aarch64" }
        default { $arch }
    }

    $filename = "suiup-Windows-msvc-${suiupArch}.zip"
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

        # Add all potential binary directories to PATH
        Add-PathEntry -PathValue $script:WalgoToolsDir
        Add-PathEntry -PathValue $script:SuiBinDir      # ~/.sui/bin - suiup's default on Windows
        Add-PathEntry -PathValue $script:LocalBinDir    # ~/.local/bin - fallback

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

function Get-SuiupBinDir {
    param([string]$SuiupPath)

    # Try to detect where suiup installs binaries
    # Check in order: ~/.sui/bin, ~/.local/bin
    $possiblePaths = @(
        $script:SuiBinDir,
        $script:LocalBinDir
    )

    # If suiup is installed, try running it to see where it installs
    if ($SuiupPath -and (Test-Path $SuiupPath)) {
        try {
            $output = & $SuiupPath show 2>&1
            # Parse output to find bin directory if possible
            if ($output -match 'bin.*?([A-Za-z]:\\[^\s]+)') {
                $detectedPath = $matches[1]
                if (Test-Path $detectedPath) {
                    return $detectedPath
                }
            }
        } catch { }
    }

    # Return first existing path or default
    foreach ($path in $possiblePaths) {
        if (Test-Path $path) {
            return $path
        }
    }

    return $script:SuiBinDir
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
    param(
        [string]$Network = "testnet",
        [string]$BinDir
    )

    Print-Info "Configuring Sui client..."

    $suiConfigPath = Join-Path $env:USERPROFILE ".sui\sui_config\client.yaml"

    # Find sui.exe - check multiple locations
    $suiPath = $null
    $possibleSuiPaths = @(
        (Join-Path $BinDir "sui.exe"),
        (Join-Path $script:SuiBinDir "sui.exe"),
        (Join-Path $script:LocalBinDir "sui.exe")
    )

    foreach ($path in $possibleSuiPaths) {
        if (Test-Path $path) {
            $suiPath = $path
            break
        }
    }

    if (-not $suiPath) {
        # Try to find in PATH
        $suiCmd = Get-Command sui.exe -ErrorAction SilentlyContinue
        if ($suiCmd) {
            $suiPath = $suiCmd.Source
        }
    }

    if (Test-Path $suiConfigPath) {
        Print-Success "Sui client already configured"

        # Even if already configured, ensure both environments exist
        if ($suiPath) {
            Print-Info "Verifying testnet and mainnet environments..."

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

    if (-not $suiPath -or -not (Test-Path $suiPath)) {
        Print-Warning "Sui CLI not found. Skipping client initialization."
        Print-Info "After installation completes, run: sui client"
        return
    }

    Print-Info "Initializing Sui client for $Network..."
    Print-Info "Creating new Sui address..."

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
            # Run directly in foreground so progress bars are visible to user.
            # Sui binary is ~230 MB and can take 10+ minutes on slow connections.
            & $suiupPath install $($target.Spec)
            if ($LASTEXITCODE -eq 0) {
                Print-Success "$($target.Label) installed"
            } else {
                Print-Warning "$($target.Label) installation may have failed (exit code: $LASTEXITCODE)"
                Print-Info "  Install manually later: suiup install $($target.Spec)"
            }
        } catch {
            $errorMsg = $_.Exception.Message
            Print-Warning "Failed to install $($target.Label) - $errorMsg"
            Print-Info "  Install manually later: suiup install $($target.Spec)"
        }
    }

    # Set defaults with proper error reporting
    $defaults = @(
        @{ Spec = "sui@testnet"; Label = "Sui CLI default" },
        @{ Spec = "walrus@testnet"; Label = "Walrus CLI default" },
        @{ Spec = "site-builder@mainnet"; Label = "site-builder default" }
    )

    foreach ($default in $defaults) {
        try {
            $result = & $suiupPath default set $default.Spec 2>&1
            if ($LASTEXITCODE -ne 0) {
                Print-Warning "Failed to set $($default.Label)"
            }
        } catch {
            Print-Warning "Failed to set $($default.Label)"
        }
    }

    # Detect where suiup installed binaries
    $binDir = Get-SuiupBinDir -SuiupPath $suiupPath
    Print-Info "Checking for binaries in: $binDir"

    $binaries = @(
        @{ Name = "Sui CLI"; Exe = "sui.exe" },
        @{ Name = "Walrus CLI"; Exe = "walrus.exe" },
        @{ Name = "site-builder"; Exe = "site-builder.exe" }
    )

    # Check multiple possible locations for each binary
    foreach ($binary in $binaries) {
        $found = $false
        $foundPath = $null

        $searchPaths = @(
            (Join-Path $binDir $binary.Exe),
            (Join-Path $script:SuiBinDir $binary.Exe),
            (Join-Path $script:LocalBinDir $binary.Exe)
        )

        foreach ($path in $searchPaths) {
            if (Test-Path $path) {
                $found = $true
                $foundPath = $path
                break
            }
        }

        if ($found) {
            try {
                $output = & $foundPath --version 2>&1 | Select-Object -First 1
                Print-Success "$($binary.Name) ready: $output"
                Print-Info "  Location: $foundPath"
            } catch {
                Print-Info "$($binary.Name) installed at $foundPath"
            }
        } else {
            Print-Warning "$($binary.Name) not found. Searched:"
            foreach ($path in $searchPaths) {
                Print-Info "  - $path"
            }
            Print-Info "Try running: suiup show"
        }
    }

    Download-WalrusConfigs
    Initialize-SuiClient -Network "testnet" -BinDir $binDir
    Print-Success "Walrus dependencies installation completed"
    Print-Info "Verify with: sui --version && walrus --version && site-builder --version"
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
    Write-Host "     - Use the desktop shortcut on your Desktop"
    Write-Host "     - Or run: walgo desktop"
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
    Test-Dependencies
    Install-WalrusDependencies
    Show-NextSteps

    # Final cleanup
    Cleanup-TempDirs
}

Main
