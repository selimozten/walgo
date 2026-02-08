package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/selimozten/walgo/pkg/api"
)

// AIProgressState holds the current AI pipeline progress for polling.
type AIProgressState struct {
	IsActive     bool    `json:"isActive"`
	Phase        string  `json:"phase"`
	Message      string  `json:"message"`
	PagePath     string  `json:"pagePath"`
	Progress     float64 `json:"progress"`
	Current      int     `json:"current"`
	Total        int     `json:"total"`
	Complete     bool    `json:"complete"`
	Success      bool    `json:"success"`
	SitePath     string  `json:"sitePath"`
	TotalPages   int     `json:"totalPages"`
	FilesCreated int     `json:"filesCreated"`
	Error        string  `json:"error"`
}

// App represents the main desktop application structure.
type App struct {
	ctx           context.Context
	mu            sync.Mutex // protects serveCmd, serverPort, serveSitePath
	serveCmd      *exec.Cmd
	serverPort    int
	serveSitePath string
	aiProgress    *AIProgressState
	aiProgressMu  sync.Mutex
	aiCancel      context.CancelFunc // cancels the running AI pipeline
	aiDone        chan struct{}      // closed when the AI goroutine exits
}

// NewApp initializes and returns a new App instance.
func NewApp() *App {
	return &App{}
}

// lookPath finds an executable in PATH or common installation directories
func lookPath(name string) (string, error) {
	// Try standard PATH lookup first
	if path, err := exec.LookPath(name); err == nil {
		return path, nil
	}

	// On Windows, binaries have .exe extension
	binaryName := name
	if runtime.GOOS == "windows" && filepath.Ext(name) == "" {
		binaryName = name + ".exe"
	}

	// Check common binary directories
	var extraDirs []string
	switch runtime.GOOS {
	case "darwin":
		extraDirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"}
	case "linux":
		extraDirs = []string{"/usr/local/bin", "/usr/bin", "/bin"}
	case "windows":
		home, _ := os.UserHomeDir()
		if home != "" {
			extraDirs = []string{
				filepath.Join(home, ".walgo", "bin"),
				filepath.Join(home, ".sui", "bin"),
				filepath.Join(home, ".local", "bin"),
			}
		}
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			extraDirs = append(extraDirs, filepath.Join(localAppData, "Walgo"))
		}
		if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
			extraDirs = append(extraDirs, filepath.Join(programFiles, "Walgo"))
		}
	}

	// Check extra directories
	for _, dir := range extraDirs {
		candidatePath := filepath.Join(dir, binaryName)
		if info, err := os.Stat(candidatePath); err == nil && !info.IsDir() {
			if runtime.GOOS == "windows" || info.Mode()&0111 != 0 {
				return candidatePath, nil
			}
		}
	}

	// Check ~/.local/bin (non-Windows, since Windows already checks it above)
	if runtime.GOOS != "windows" {
		if home, err := os.UserHomeDir(); err == nil {
			localBin := filepath.Join(home, ".local", "bin", binaryName)
			if info, err := os.Stat(localBin); err == nil && !info.IsDir() {
				if info.Mode()&0111 != 0 {
					return localBin, nil
				}
			}
		}
	}

	return "", fmt.Errorf("%s not found in PATH", name)
}

// startup is invoked when the application starts. The context is saved
// to enable runtime method calls throughout the application lifecycle.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Window Controls (for frameless window)

// Minimize minimizes application window.
func (a *App) Minimize() {
	windowMinimise(a.ctx)
}

// Maximize toggles between maximize and restore states for window.
func (a *App) Maximize() {
	windowToggleMaximise(a.ctx)
}

// Close stops any running server and terminates the application.
func (a *App) Close() {
	a.cancelAI()
	a.StopServe()
	appQuit(a.ctx)
}

// beforeClose is called by Wails before the window closes.
// It cancels any running AI pipeline so the defer cleanup runs,
// then allows the close to proceed.
func (a *App) beforeClose(_ context.Context) bool {
	a.cancelAI()
	return false // allow close
}

// shutdown is called by Wails when the application is shutting down.
func (a *App) shutdown(_ context.Context) {
	a.cancelAI()
	a.StopServe()
}

// File Dialogs

// SelectDirectory opens a directory selection dialog and returns the chosen path.
func (a *App) SelectDirectory(title string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	return openDirectoryDialog(a.ctx, title, homeDir)
}

// Version Management

// VersionResult contains version information for the application.
type VersionResult = api.VersionResult

// GetVersion retrieves and returns the current application version information.
func (a *App) GetVersion() VersionResult {
	return api.GetVersion()
}

// GetDefaultSitesDir returns the default walgo-sites directory path
func (a *App) GetDefaultSitesDir() string {
	sitesDir, err := api.GetDefaultSitesDir()
	if err != nil {
		return ""
	}
	return sitesDir
}

// SystemHealth holds system health information
type SystemHealth = api.SystemHealth

// GetSystemHealth returns current system health status
func (a *App) GetSystemHealth() SystemHealth {
	return api.GetSystemHealth()
}

// ToolVersionInfo represents version information for a tool
type ToolVersionInfo = api.ToolVersionInfo

// CheckToolVersionsResult holds the result of version checking
type CheckToolVersionsResult = api.CheckToolVersionsResult

// CheckToolVersions checks if installed tools have updates available
func (a *App) CheckToolVersions() CheckToolVersionsResult {
	return api.CheckToolVersions()
}

// UpdateToolsParams holds parameters for updating tools
type UpdateToolsParams = api.UpdateToolsParams

// UpdateToolsResult holds the result of updating tools
type UpdateToolsResult = api.UpdateToolsResult

// UpdateTools updates specified tools to their latest versions
func (a *App) UpdateTools(params UpdateToolsParams) UpdateToolsResult {
	return api.UpdateTools(params)
}

// ====================
// Site Management
// ====================

// InitSiteResult holds init site result
type InitSiteResult = api.InitSiteResult

// InitSite initializes a new Hugo site in current directory (like 'walgo init')
func (a *App) InitSite(parentDir string, siteName string) InitSiteResult {
	return api.InitSite(parentDir, siteName)
}

// BuildSite builds the Hugo site at the given path
func (a *App) BuildSite(sitePath string) error {
	return api.BuildSite(sitePath)
}

// ====================
// Theme Management
// ====================

// InstallThemeParams holds theme installation parameters
type InstallThemeParams = api.InstallThemeParams

// InstallThemeResult holds theme installation result
type InstallThemeResult = api.InstallThemeResult

// GetInstalledThemesResult holds list of installed themes
type GetInstalledThemesResult = api.GetInstalledThemesResult

// InstallTheme installs a Hugo theme from GitHub URL
func (a *App) InstallTheme(params InstallThemeParams) InstallThemeResult {
	return api.InstallTheme(params)
}

// GetInstalledThemes returns list of installed themes
func (a *App) GetInstalledThemes(sitePath string) GetInstalledThemesResult {
	return api.GetInstalledThemes(sitePath)
}

// ====================
// AI Features
// ====================

// AIConfigureParams holds AI configuration parameters
type AIConfigureParams = api.AIConfigureParams
type AIConfigResult = api.AIConfigResult

// GetAIConfig retrieves current AI configuration
func (a *App) GetAIConfig() AIConfigResult {
	result, err := api.GetAIConfig()
	if err != nil {
		return AIConfigResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	return *result
}

// UpdateAIConfig updates AI configuration
func (a *App) UpdateAIConfig(params AIConfigureParams) error {
	return api.UpdateAIConfig(params)
}

// CleanAIConfig removes all AI credentials
func (a *App) CleanAIConfig() error {
	return api.CleanAIConfig()
}

// CleanProviderConfig removes credentials for a specific provider
func (a *App) CleanProviderConfig(provider string) error {
	return api.CleanProviderConfig(provider)
}

// GetProviderCredentials returns credentials for a specific provider
func (a *App) GetProviderCredentials(provider string) ProviderCredentialsResult {
	result, err := api.GetProviderCredentialsAPI(provider)
	if err != nil {
		return ProviderCredentialsResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	return *result
}

// ProviderCredentialsResult holds provider credentials
type ProviderCredentialsResult = api.ProviderCredentialsResult

// ====================
// AI Content Generation
// ====================

type GenerateContentParams = api.GenerateContentParams
type GenerateContentResult = api.GenerateContentResult

// ContentStructure holds information about the content directory structure
type ContentStructure = api.ContentStructure
type ContentTypeInfo = api.ContentTypeInfo

// GetContentStructure returns the content directory structure
func (a *App) GetContentStructure(sitePath string) ContentStructure {
	return api.GetContentStructure(sitePath)
}

func (a *App) GenerateContent(params GenerateContentParams) GenerateContentResult {
	return api.GenerateContent(params)
}

// UpdateContentParams holds content update parameters
type UpdateContentParams = api.UpdateContentParams
type UpdateContentResult = api.UpdateContentResult

// UpdateContent updates existing content using AI
func (a *App) UpdateContent(params UpdateContentParams) UpdateContentResult {
	return api.UpdateContent(params)
}

// ====================
// Projects Management
// ====================

// Project represents a Walrus site project
type Project = api.Project

// =============================================================================
// Wallet Management
// =============================================================================

// WalletInfo holds wallet information
type WalletInfo = api.WalletInfo

// GetWalletInfo returns current wallet information
func (a *App) GetWalletInfo() (*WalletInfo, error) {
	return api.GetWalletInfo()
}

// AddressListResult holds list of addresses
type AddressListResult = api.AddressListResult

// GetAddressList returns list of all wallet addresses
func (a *App) GetAddressList() AddressListResult {
	return api.GetAddressList()
}

// SwitchAddressParams holds parameters for switching address
type SwitchAddressParams = api.SwitchAddressParams

// SwitchAddressResult holds result of switching address
type SwitchAddressResult = api.SwitchAddressResult

// SwitchAddress switches to a different wallet address
func (a *App) SwitchAddress(address string) SwitchAddressResult {
	return api.SwitchAddress(address)
}

// CreateAddressParams holds parameters for creating address
type CreateAddressParams = api.CreateAddressParams

// CreateAddressResult holds result of creating address
type CreateAddressResult = api.CreateAddressResult

// CreateAddress creates a new wallet address
func (a *App) CreateAddress(keyScheme string, alias string) CreateAddressResult {
	return api.CreateAddress(keyScheme, alias)
}

// ImportAddressParams holds import parameters
type ImportAddressParams = api.ImportAddressParams

// ImportAddressResult holds import result
type ImportAddressResult = api.ImportAddressResult

// ImportAddress imports a wallet address
func (a *App) ImportAddress(params ImportAddressParams) ImportAddressResult {
	return api.ImportAddress(params)
}

// SwitchNetworkResult holds result of switching network
type SwitchNetworkResult = api.SwitchNetworkResult

// SwitchNetwork switches to a different network (testnet/mainnet)
func (a *App) SwitchNetwork(network string) SwitchNetworkResult {
	return api.SwitchNetwork(network)
}

// ListProjects returns all projects
func (a *App) ListProjects() ([]Project, error) {
	return api.ListProjects()
}

// GetProject returns a single project by ID
func (a *App) GetProject(projectID int64) (*Project, error) {
	return api.GetProject(projectID)
}

// DeleteProjectParams holds delete project parameters
type DeleteProjectParams = api.DeleteProjectParams

// DeleteProjectResult holds delete result
type DeleteProjectResult = api.DeleteProjectResult

// DeleteProject deletes a project by ID (includes on-chain destruction if objectId exists)
func (a *App) DeleteProject(params DeleteProjectParams) DeleteProjectResult {
	return api.DeleteProject(params)
}

// ProjectNameExists checks if a project with the given name already exists
func (a *App) ProjectNameExists(name string) (bool, error) {
	return api.ProjectNameExists(name)
}

// ====================
// Import
// ====================

type ImportObsidianParams = api.ImportObsidianParams
type ImportObsidianResult = api.ImportObsidianResult

// ImportObsidian creates a new Hugo site and imports content from Obsidian vault
func (a *App) ImportObsidian(params ImportObsidianParams) ImportObsidianResult {
	return api.ImportObsidian(params)
}

// ====================
// QuickStart
// ====================

// QuickStartParams holds quickstart parameters
type QuickStartParams = api.QuickStartParams
type QuickStartResult = api.QuickStartResult

// QuickStart creates a new Hugo site with quickstart flow
func (a *App) QuickStart(params QuickStartParams) QuickStartResult {
	return api.QuickStart(params)
}

// ====================
// Serve
// ====================

// ServeParams holds serve parameters
type ServeParams = api.ServeParams
type ServeResult = api.ServeResult

// Serve starts local Hugo development server
func (a *App) Serve(params ServeParams) ServeResult {
	// Stop any existing server first
	a.StopServe()

	// Kill any existing Hugo serve processes system-wide
	if err := killExistingHugoProcesses(); err != nil {
		fmt.Printf("⚠️  Warning: Error cleaning up existing Hugo processes: %v\n", err)
	}

	// Check if hugo is installed
	hugoPath, err := lookPath("hugo")
	if err != nil {
		return ServeResult{Error: "hugo is not installed or not found in PATH"}
	}

	// Get the site path
	sitePath := params.SitePath
	if sitePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return ServeResult{Error: fmt.Sprintf("cannot determine current directory: %v", err)}
		}
		sitePath = cwd
	}

	// Build the site before serving
	buildCmd := exec.Command(hugoPath, "--source", sitePath)
	hideWindow(buildCmd)
	if output, err := buildCmd.CombinedOutput(); err != nil {
		return ServeResult{Error: fmt.Sprintf("build failed: %s", string(output))}
	}

	// Build hugo arguments
	hugoArgs := []string{"server"}
	if params.Drafts {
		hugoArgs = append(hugoArgs, "-D")
	}
	if params.Expired {
		hugoArgs = append(hugoArgs, "-E")
	}
	if params.Future {
		hugoArgs = append(hugoArgs, "-F")
	}

	// Set port
	port := params.Port
	if port == 0 {
		port = 1313
	}
	hugoArgs = append(hugoArgs, "--port", strconv.Itoa(port))

	// Create and start the command
	cmd := exec.Command(hugoPath, hugoArgs...)
	cmd.Dir = sitePath

	// Hide console window on Windows
	hideWindow(cmd)

	if err := cmd.Start(); err != nil {
		return ServeResult{Error: fmt.Sprintf("failed to start hugo server: %v", err)}
	}

	// Track the running command, port, and site path
	a.mu.Lock()
	a.serveCmd = cmd
	a.serverPort = port
	a.serveSitePath = sitePath
	a.mu.Unlock()

	return ServeResult{
		Success: true,
		URL:     fmt.Sprintf("http://localhost:%d", port),
	}
}

// StopServe stops the running Hugo development server and cleans up localhost files
func (a *App) StopServe() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.serveCmd == nil || a.serveCmd.Process == nil {
		return false
	}

	_ = a.serveCmd.Process.Kill()
	_ = a.serveCmd.Wait() // reap the process to avoid zombies
	a.serveCmd = nil
	a.serveSitePath = ""
	a.serverPort = 0
	return true
}

// GetServerURL returns the current server URL
func (a *App) GetServerURL() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.serverPort > 0 {
		return fmt.Sprintf("http://localhost:%d", a.serverPort)
	}
	return ""
}

// ====================
// New Content
// ====================

// NewContentParams holds content creation parameters
type NewContentParams = api.NewContentParams
type NewContentResult = api.NewContentResult

// NewContent creates new content in Hugo site
func (a *App) NewContent(params NewContentParams) NewContentResult {
	return api.NewContent(params)
}

// ====================
// Project Metadata Operations
// ====================

// EditProjectParams holds project edit parameters
type EditProjectParams = api.EditProjectParams
type EditProjectResult = api.EditProjectResult

// EditProject updates project metadata
func (a *App) EditProject(params EditProjectParams) EditProjectResult {
	return api.EditProject(params)
}

// ArchiveProjectResult holds archive results
type ArchiveProjectResult = api.ArchiveProjectResult

// ArchiveProject archives a project
func (a *App) ArchiveProject(projectID int64) ArchiveProjectResult {
	return api.ArchiveProject(projectID)
}

// ====================
// Update Site (Re-deploy)
// ====================

// UpdateSiteParams holds update site parameters
type UpdateSiteParams = api.UpdateSiteParams

// UpdateSiteResult holds update site result
type UpdateSiteResult = api.UpdateSiteResult

// UpdateSite updates an existing project's site on Walrus (re-deploy)
func (a *App) UpdateSite(params UpdateSiteParams) UpdateSiteResult {
	return api.UpdateSite(params)
}

type SetStatusParams = api.SetStatusParams
type SetStatusResult = api.SetStatusResult

func (a *App) SetStatus(params SetStatusParams) SetStatusResult {
	return api.SetStatus(params)
}

// LaunchStep represents a step in the launch wizard
type LaunchStep = api.LaunchStep

// ====================
// Gas Estimation
// ====================

type GasEstimateParams = api.GasEstimateParams
type GasEstimateResult = api.GasEstimateResult

// EstimateGasFee estimates the gas fee for deploying a site
func (a *App) EstimateGasFee(params GasEstimateParams) GasEstimateResult {
	return api.EstimateGasFee(params)
}

// ====================
// Launch Wizard
// ====================

type LaunchWizardParams = api.LaunchWizardParams
type LaunchWizardResult = api.LaunchWizardResult

// LaunchWizard executes full launch wizard flow
func (a *App) LaunchWizard(params LaunchWizardParams) LaunchWizardResult {
	return api.LaunchWizard(params)
}

// ====================
// AI Create Site
// ====================

// AICreateSiteParams holds AI site creation parameters
type AICreateSiteParams = api.AICreateSiteParams
type AICreateSiteResult = api.AICreateSiteResult

// GetAIProgress returns the current AI pipeline progress state for polling.
func (a *App) GetAIProgress() AIProgressState {
	a.aiProgressMu.Lock()
	defer a.aiProgressMu.Unlock()
	if a.aiProgress == nil {
		return AIProgressState{}
	}
	return *a.aiProgress
}

// AICreateSite creates a complete Hugo site with AI-generated content.
// Runs the pipeline in a background goroutine. The frontend polls
// GetAIProgress() to track progress and detect completion.
// The pipeline is cancelled automatically if the app is closed.
func (a *App) AICreateSite(params AICreateSiteParams) {
	// Cancel any previous in-flight pipeline
	a.cancelAI()

	// Initialize progress state and cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	a.aiProgressMu.Lock()
	a.aiProgress = &AIProgressState{IsActive: true}
	a.aiCancel = cancel
	a.aiDone = done
	a.aiProgressMu.Unlock()

	progressHandler := func(event api.ProgressEvent) {
		a.aiProgressMu.Lock()
		defer a.aiProgressMu.Unlock()
		if a.aiProgress == nil {
			return
		}
		a.aiProgress.Phase = event.Phase
		a.aiProgress.PagePath = event.PagePath
		a.aiProgress.Progress = event.Progress
		a.aiProgress.Current = event.Current
		a.aiProgress.Total = event.Total

		// Show user-friendly messages for retries and errors
		switch event.EventType {
		case "retry":
			a.aiProgress.Message = fmt.Sprintf("Retrying %s (rate limited, waiting...)", event.PagePath)
		case "error":
			a.aiProgress.Message = fmt.Sprintf("Failed: %s", event.PagePath)
		default:
			a.aiProgress.Message = event.Message
		}
	}

	go func() {
		defer close(done)
		result := api.AICreateSiteWithProgress(ctx, params, progressHandler)

		a.aiProgressMu.Lock()
		defer a.aiProgressMu.Unlock()
		if a.aiProgress == nil {
			a.aiProgress = &AIProgressState{}
		}
		a.aiProgress.Complete = true
		a.aiProgress.Success = result.Success
		a.aiProgress.SitePath = result.SitePath
		a.aiProgress.TotalPages = result.TotalPages
		a.aiProgress.FilesCreated = result.FilesCreated
		a.aiProgress.Error = result.Error
		a.aiProgress.IsActive = false
		a.aiCancel = nil
		a.aiDone = nil
	}()
}

// cancelAI cancels any running AI pipeline and waits for it to finish.
func (a *App) cancelAI() {
	a.aiProgressMu.Lock()
	cancel := a.aiCancel
	done := a.aiDone
	a.aiProgressMu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

// ====================
// Setup Dependencies
// ====================

// SetupDepsResult holds setup dependencies result
type SetupDepsResult = api.SetupDepsResult

// CheckSetupDeps checks if all required dependencies are installed
func (a *App) CheckSetupDeps() SetupDepsResult {
	return api.CheckSetupDeps()
}

// ====================
// AI Credentials Management
// ====================

// RemoveAICredentials removes all AI credentials.
// The provider parameter is accepted for future per-provider support but currently ignored.
func (a *App) RemoveAICredentials(_ string) error {
	return api.CleanAIConfig()
}

// ====================
// Open External
// ====================

// OpenInBrowser opens a URL in the default browser
func (a *App) OpenInBrowser(url string) {
	browserOpenURL(a.ctx, url)
}

// OpenInFinder opens a path in Finder (macOS), Explorer (Windows), or file manager (Linux)
func (a *App) OpenInFinder(path string) error {
	return openFileExplorer(path)
}

// ====================
// File Management
// ====================

// FileInfo holds file metadata
type FileInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	IsDir    bool   `json:"isDir"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

// ListFilesResult holds the result of listing files
type ListFilesResult struct {
	Success bool       `json:"success"`
	Files   []FileInfo `json:"files"`
	Error   string     `json:"error"`
}

// ListFiles lists all files in a directory
func (a *App) ListFiles(dirPath string) ListFilesResult {
	result := ListFilesResult{
		Success: false,
		Files:   []FileInfo{},
	}

	safePath, err := validateFilePath(dirPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	info, err := os.Stat(safePath)
	if err != nil {
		result.Error = fmt.Sprintf("Directory not accessible: %v", err)
		return result
	}

	if !info.IsDir() {
		result.Error = "Path is not a directory"
		return result
	}

	// Read directory
	entries, err := os.ReadDir(safePath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read directory: %v", err)
		return result
	}

	files := []FileInfo{}
	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		fullPath := filepath.Join(safePath, entry.Name())

		files = append(files, FileInfo{
			Name:     entry.Name(),
			Path:     fullPath,
			IsDir:    entry.IsDir(),
			Size:     fileInfo.Size(),
			Modified: fileInfo.ModTime().Unix(),
		})
	}

	result.Files = files
	result.Success = true
	return result
}

// FolderStatsResult holds statistics for a folder
type FolderStatsResult struct {
	Success     bool   `json:"success"`
	FileCount   int    `json:"fileCount"`
	FolderCount int    `json:"folderCount"`
	TotalSize   int64  `json:"totalSize"`
	Error       string `json:"error"`
}

// GetFolderStats recursively gets statistics for a folder
func (a *App) GetFolderStats(dirPath string) FolderStatsResult {
	result := FolderStatsResult{
		Success: false,
	}

	safePath, err := validateFilePath(dirPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	dirPath = safePath

	// Validate directory exists
	info, err := os.Stat(dirPath)
	if err != nil {
		result.Error = fmt.Sprintf("Directory not accessible: %v", err)
		return result
	}

	if !info.IsDir() {
		result.Error = "Path is not a directory"
		return result
	}

	// Recursively count files and calculate size
	fileCount, folderCount, totalSize := countFilesRecursive(dirPath)
	result.FileCount = fileCount
	result.FolderCount = folderCount
	result.TotalSize = totalSize
	result.Success = true
	return result
}

// countFilesRecursive recursively counts files, folders, and total size
func countFilesRecursive(dirPath string) (fileCount int, folderCount int, totalSize int64) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, 0, 0
	}

	for _, entry := range entries {
		// Skip hidden files and directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(dirPath, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			folderCount++
			fc, dc, ts := countFilesRecursive(fullPath)
			fileCount += fc
			folderCount += dc
			totalSize += ts
		} else {
			fileCount++
			totalSize += fileInfo.Size()
		}
	}

	return fileCount, folderCount, totalSize
}

// validateFilePath ensures the path resolves inside the user's home directory,
// preventing path traversal attacks from the frontend.
func validateFilePath(filePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	absPath, err := filepath.Abs(filepath.Clean(filePath))
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}
	absHome, err := filepath.Abs(homeDir)
	if err != nil {
		return "", fmt.Errorf("invalid home path: %w", err)
	}
	if !strings.HasPrefix(absPath, absHome+string(filepath.Separator)) && absPath != absHome {
		return "", fmt.Errorf("access denied: path is outside home directory")
	}
	return absPath, nil
}

// ReadFileResult holds the result of reading a file
type ReadFileResult struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Error   string `json:"error"`
}

// ReadFile reads the content of a file
func (a *App) ReadFile(filePath string) ReadFileResult {
	result := ReadFileResult{
		Success: false,
		Content: "",
	}

	safePath, err := validateFilePath(filePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	info, err := os.Stat(safePath)
	if err != nil {
		result.Error = fmt.Sprintf("File not accessible: %v", err)
		return result
	}

	if info.IsDir() {
		result.Error = "Path is a directory, not a file"
		return result
	}

	content, err := os.ReadFile(safePath) // #nosec G304 - path validated by validateFilePath
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read file: %v", err)
		return result
	}

	result.Content = string(content)
	result.Success = true
	return result
}

// WriteFileResult holds the result of writing a file
type WriteFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// WriteFile writes content to a file
func (a *App) WriteFile(filePath string, content string) WriteFileResult {
	result := WriteFileResult{
		Success: false,
	}

	safePath, err := validateFilePath(filePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	dir := filepath.Dir(safePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return result
	}

	if err := os.WriteFile(safePath, []byte(content), 0644); err != nil {
		result.Error = fmt.Sprintf("Failed to write file: %v", err)
		return result
	}

	result.Success = true
	return result
}

// DeleteFileResult holds the result of deleting a file
type DeleteFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// DeleteFile deletes a file or directory
func (a *App) DeleteFile(filePath string) DeleteFileResult {
	result := DeleteFileResult{
		Success: false,
	}

	safePath, err := validateFilePath(filePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	info, err := os.Stat(safePath)
	if err != nil {
		result.Error = fmt.Sprintf("File not accessible: %v", err)
		return result
	}

	if info.IsDir() {
		if err := os.RemoveAll(safePath); err != nil {
			result.Error = fmt.Sprintf("Failed to delete directory: %v", err)
			return result
		}
	} else {
		if err := os.Remove(safePath); err != nil {
			result.Error = fmt.Sprintf("Failed to delete file: %v", err)
			return result
		}
	}

	result.Success = true
	return result
}

// CreateFileResult holds the result of creating a new file
type CreateFileResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}

// CreateFile creates a new file with optional content
func (a *App) CreateFile(filePath string, content string) CreateFileResult {
	result := CreateFileResult{
		Success: false,
		Path:    filePath,
	}

	safePath, err := validateFilePath(filePath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// If file exists, find a unique name by adding (1), (2), etc.
	uniquePath := findUniquePath(safePath)
	result.Path = uniquePath

	// Create directory if it doesn't exist
	dir := filepath.Dir(uniquePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return result
	}

	// Create and write file
	if err := os.WriteFile(uniquePath, []byte(content), 0644); err != nil {
		result.Error = fmt.Sprintf("Failed to create file: %v", err)
		return result
	}

	result.Success = true
	return result
}

// CreateDirectoryResult holds the result of creating a directory
type CreateDirectoryResult struct {
	Success bool   `json:"success"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(dirPath string) CreateDirectoryResult {
	result := CreateDirectoryResult{
		Success: false,
		Path:    dirPath,
	}

	safePath, err := validateFilePath(dirPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	// If directory exists, find a unique name by adding (1), (2), etc.
	uniquePath := findUniquePath(safePath)
	result.Path = uniquePath

	// Create directory
	if err := os.MkdirAll(uniquePath, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return result
	}

	result.Success = true
	return result
}

// MoveFileResult holds the result of a move operation
type MoveFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// MoveFile moves a file or directory from oldPath to newPath
func (a *App) MoveFile(oldPath string, newPath string) MoveFileResult {
	result := MoveFileResult{
		Success: false,
		OldPath: oldPath,
		NewPath: newPath,
	}

	safeOldPath, err := validateFilePath(oldPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	safeNewPath, err := validateFilePath(newPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	oldPath = safeOldPath
	newPath = safeNewPath

	// Check if source exists
	if _, err := os.Stat(oldPath); err != nil {
		result.Error = fmt.Sprintf("Source not found: %v", err)
		return result
	}

	// Create destination directory if needed
	destDir := filepath.Dir(newPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create destination directory: %v", err)
		return result
	}

	// Check if destination already exists right before rename to minimize race window
	if _, err := os.Stat(newPath); err == nil {
		result.Error = "Destination already exists"
		return result
	}

	// Move/rename the file or directory
	if err := os.Rename(oldPath, newPath); err != nil {
		// Handle race: another process may have created the destination
		if _, statErr := os.Stat(newPath); statErr == nil {
			result.Error = "Destination already exists"
		} else {
			result.Error = fmt.Sprintf("Failed to move: %v", err)
		}
		return result
	}

	result.Success = true
	return result
}

// RenameFileResult holds the result of a rename operation
type RenameFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	OldPath string `json:"oldPath"`
	NewPath string `json:"newPath"`
}

// RenameFile renames a file or directory
func (a *App) RenameFile(oldPath string, newName string) RenameFileResult {
	result := RenameFileResult{
		Success: false,
		OldPath: oldPath,
	}

	safeOldPath, err := validateFilePath(oldPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	oldPath = safeOldPath

	// Calculate new path
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
	result.NewPath = newPath

	safeNewPath, err := validateFilePath(newPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	newPath = safeNewPath
	result.NewPath = newPath

	// Check if source exists
	if _, err := os.Stat(oldPath); err != nil {
		result.Error = fmt.Sprintf("Source not found: %v", err)
		return result
	}

	// Check if destination already exists
	if _, err := os.Stat(newPath); err == nil {
		result.Error = "A file or directory with this name already exists"
		return result
	}

	// Rename the file or directory
	if err := os.Rename(oldPath, newPath); err != nil {
		result.Error = fmt.Sprintf("Failed to rename: %v", err)
		return result
	}

	result.Success = true
	return result
}

// CopyFileResult holds the result of a copy operation
type CopyFileResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	SrcPath string `json:"srcPath"`
	DstPath string `json:"dstPath"`
}

// CopyFile copies a file from srcPath to dstPath
func (a *App) CopyFile(srcPath string, dstPath string) CopyFileResult {
	result := CopyFileResult{
		Success: false,
		SrcPath: srcPath,
		DstPath: dstPath,
	}

	safeSrcPath, err := validateFilePath(srcPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	safeDstPath, err := validateFilePath(dstPath)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	srcPath = safeSrcPath
	dstPath = safeDstPath

	// Check if source exists
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		result.Error = fmt.Sprintf("Source not found: %v", err)
		return result
	}

	// If destination exists, find a unique name by adding (1), (2), etc.
	uniqueDstPath := findUniquePath(dstPath)
	result.DstPath = uniqueDstPath

	// Handle directory copy recursively
	if srcInfo.IsDir() {
		if err := copyDir(srcPath, uniqueDstPath); err != nil {
			result.Error = fmt.Sprintf("Failed to copy directory: %v", err)
			return result
		}
		result.Success = true
		return result
	}

	// Handle file copy
	// Create destination directory if needed
	destDir := filepath.Dir(uniqueDstPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create destination directory: %v", err)
		return result
	}

	// Read source file
	data, err := os.ReadFile(srcPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read source: %v", err)
		return result
	}

	// Write to destination
	if err := os.WriteFile(uniqueDstPath, data, srcInfo.Mode()); err != nil {
		result.Error = fmt.Sprintf("Failed to write destination: %v", err)
		return result
	}

	result.Success = true
	return result
}

// getDirectoryDepth calculates the depth of a directory tree
func getDirectoryDepth(path string) (int, error) {
	return getDirectoryDepthRecursive(path, 0)
}

// getDirectoryDepthRecursive recursively calculates directory depth
func getDirectoryDepthRecursive(path string, currentDepth int) (int, error) {
	const maxCheckDepth = 100
	if currentDepth > maxCheckDepth {
		return currentDepth, nil // Return early if too deep
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return currentDepth, err
	}

	maxDepth := currentDepth
	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(path, entry.Name())

			// Check for symlinks
			entryInfo, err := os.Lstat(entryPath)
			if err != nil {
				continue
			}
			if entryInfo.Mode()&os.ModeSymlink != 0 {
				continue
			}

			depth, err := getDirectoryDepthRecursive(entryPath, currentDepth+1)
			if err != nil {
				continue
			}
			if depth > maxDepth {
				maxDepth = depth
			}
		}
	}

	return maxDepth, nil
}

// CheckDirectoryDepthResult holds the result of directory depth check
type CheckDirectoryDepthResult struct {
	Success bool   `json:"success"`
	Depth   int    `json:"depth"`
	TooDeep bool   `json:"tooDeep"`
	Error   string `json:"error"`
}

// CheckDirectoryDepth checks if a directory is too deep for operations
func (a *App) CheckDirectoryDepth(path string) CheckDirectoryDepthResult {
	result := CheckDirectoryDepthResult{
		Success: false,
	}

	validPath, err := validateFilePath(path)
	if err != nil {
		result.Error = fmt.Sprintf("Invalid path: %v", err)
		return result
	}

	info, err := os.Stat(validPath)
	if err != nil {
		result.Error = fmt.Sprintf("Path not found: %v", err)
		return result
	}

	if !info.IsDir() {
		// Files are fine, depth is 0
		result.Success = true
		result.Depth = 0
		result.TooDeep = false
		return result
	}

	depth, err := getDirectoryDepth(validPath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to check depth: %v", err)
		return result
	}

	const maxSafeDepth = 50 // Lower threshold to prevent operations before they cause issues
	result.Success = true
	result.Depth = depth
	result.TooDeep = depth > maxSafeDepth

	return result
}

// copyDir recursively copies a directory tree
func copyDir(src string, dst string) error {
	return copyDirWithDepth(src, dst, 0)
}

// copyDirWithDepth recursively copies a directory tree with depth limit
func copyDirWithDepth(src string, dst string, depth int) error {
	// Prevent infinite recursion
	const maxDepth = 100
	if depth > maxDepth {
		return fmt.Errorf("maximum directory depth exceeded (possible circular reference)")
	}

	// Get absolute paths to prevent circular copies
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absDst, err := filepath.Abs(dst)
	if err != nil {
		return err
	}

	// Check if destination is inside source (would create circular copy)
	if strings.HasPrefix(absDst, absSrc+string(filepath.Separator)) {
		return fmt.Errorf("cannot copy directory into itself")
	}

	// Check if paths are the same
	if absSrc == absDst {
		return fmt.Errorf("source and destination are the same")
	}

	// Get source directory info
	srcInfo, err := os.Stat(absSrc)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(absDst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(absSrc)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(absSrc, entry.Name())
		dstPath := filepath.Join(absDst, entry.Name())

		// Get entry info to check for symlinks
		entryInfo, err := os.Lstat(srcPath) // Use Lstat to not follow symlinks
		if err != nil {
			continue // Skip entries we can't read
		}

		// Skip symlinks to prevent circular references
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}

		if entry.IsDir() {
			// Recursively copy subdirectory with incremented depth
			if err := copyDirWithDepth(srcPath, dstPath, depth+1); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, srcInfo.Mode())
}

// findUniquePath finds a unique path by adding (1), (2), etc. if the path exists
func findUniquePath(path string) string {
	// If path doesn't exist, use it as-is
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	// Extract directory, base name, and extension
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Try adding (1), (2), (3), etc.
	counter := 1
	for {
		var newName string
		if ext != "" {
			newName = fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		} else {
			newName = fmt.Sprintf("%s (%d)", nameWithoutExt, counter)
		}
		newPath := filepath.Join(dir, newName)

		// Check if this path exists
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}

		counter++
		// Safety limit to avoid infinite loop
		if counter > 1000 {
			return path
		}
	}
}
