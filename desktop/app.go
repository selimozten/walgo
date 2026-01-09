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

	"github.com/selimozten/walgo/pkg/api"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App represents the main desktop application structure.
type App struct {
	ctx           context.Context
	serveCmd      *exec.Cmd // Track running Hugo server
	serverPort    int
	serveSitePath string // Track site path for cleanup
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

	// Check common binary directories
	var extraDirs []string
	switch runtime.GOOS {
	case "darwin":
		extraDirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin"}
	case "linux":
		extraDirs = []string{"/usr/local/bin", "/usr/bin", "/bin"}
	case "windows":
		// Windows installers typically add to PATH
		return "", fmt.Errorf("%s not found in PATH", name)
	}

	// Check extra directories
	for _, dir := range extraDirs {
		candidatePath := filepath.Join(dir, name)
		if info, err := os.Stat(candidatePath); err == nil && !info.IsDir() {
			if runtime.GOOS == "windows" || info.Mode()&0111 != 0 {
				return candidatePath, nil
			}
		}
	}

	// Check ~/.local/bin
	if home, err := os.UserHomeDir(); err == nil {
		localBin := filepath.Join(home, ".local", "bin", name)
		if info, err := os.Stat(localBin); err == nil && !info.IsDir() {
			if runtime.GOOS == "windows" || info.Mode()&0111 != 0 {
				return localBin, nil
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
	wruntime.WindowMinimise(a.ctx)
}

// Maximize toggles between maximize and restore states for window.
func (a *App) Maximize() {
	wruntime.WindowToggleMaximise(a.ctx)
}

// Close stops any running server and terminates the application.
func (a *App) Close() {
	// Stop any running server first
	a.StopServe()
	wruntime.Quit(a.ctx)
}

// File Dialogs

// SelectDirectory opens a directory selection dialog and returns the chosen path.
func (a *App) SelectDirectory(title string) (string, error) {
	homeDir, _ := os.UserHomeDir()
	return wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: homeDir,
	})
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

// GenerateContentParams holds content generation parameters (legacy - kept for compatibility)
type GenerateContentParams struct {
	SitePath     string `json:"sitePath"`
	FilePath     string `json:"filePath"`
	ContentType  string `json:"contentType"` // "post" or "page"
	Topic        string `json:"topic"`
	Context      string `json:"context"`
	Instructions string `json:"instructions"` // New: simplified instructions
}

// GenerateContentResult holds the result of content generation
type GenerateContentResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
	Error    string `json:"error"`
}

// ContentStructure holds information about the content directory structure
type ContentStructure = api.ContentStructure
type ContentTypeInfo = api.ContentTypeInfo

// GetContentStructure returns the content directory structure
func (a *App) GetContentStructure(sitePath string) ContentStructure {
	return api.GetContentStructure(sitePath)
}

func (a *App) GenerateContent(params GenerateContentParams) GenerateContentResult {
	apiParams := api.GenerateContentParams{
		SitePath:     params.SitePath,
		FilePath:     params.FilePath,
		ContentType:  params.ContentType,
		Topic:        params.Topic,
		Context:      params.Context,
		Instructions: params.Instructions,
	}
	res := api.GenerateContent(apiParams)
	return GenerateContentResult{
		Success:  res.Success,
		Content:  res.Content,
		FilePath: res.FilePath,
		Error:    res.Error,
	}
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

// ImportObsidianParams holds import parameters
type ImportObsidianParams struct {
	VaultPath     string `json:"vaultPath"`
	SiteName      string `json:"siteName"`      // Name for new site (defaults to vault name)
	ParentDir     string `json:"parentDir"`     // Parent directory for site (defaults to current dir)
	OutputDir     string `json:"outputDir"`     // Subdirectory in content for imported files
	DryRun        bool   `json:"dryRun"`        // Preview without creating site
	ConvertLinks  bool   `json:"convertLinks"`  // Convert wikilinks
	LinkStyle     string `json:"linkStyle"`     // "markdown" (default) or "relref"
	IncludeDrafts bool   `json:"includeDrafts"` // Include draft content
}

// ImportObsidianResult holds import results
type ImportObsidianResult struct {
	Success       bool   `json:"success"`
	FilesImported int    `json:"filesImported"`
	SitePath      string `json:"sitePath"` // Path to created site
	Error         string `json:"error"`
}

// ImportObsidian creates a new Hugo site and imports content from Obsidian vault
func (a *App) ImportObsidian(params ImportObsidianParams) ImportObsidianResult {
	res := api.ImportObsidian(api.ImportObsidianParams{
		VaultPath:    params.VaultPath,
		SiteName:     params.SiteName,
		ParentDir:    params.ParentDir,
		OutputDir:    params.OutputDir,
		DryRun:       params.DryRun,
		ConvertLinks: params.ConvertLinks,
	})
	return ImportObsidianResult{
		Success:       res.Success,
		FilesImported: res.FilesImported,
		SitePath:      res.SitePath,
		Error:         res.Error,
	}
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

// killExistingHugoProcesses finds and kills any existing 'hugo serve' processes
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

	// Build hugo arguments
	hugoArgs := []string{"server"}
	if params.Drafts {
		hugoArgs = append(hugoArgs, "-D")
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

	if err := cmd.Start(); err != nil {
		return ServeResult{Error: fmt.Sprintf("failed to start hugo server: %v", err)}
	}

	// Track the running command, port, and site path
	a.serveCmd = cmd
	a.serverPort = port
	a.serveSitePath = sitePath

	return ServeResult{
		Success: true,
		URL:     fmt.Sprintf("http://localhost:%d", port),
	}
}

// getProductionBaseURLFromHugoToml reads the baseURL from hugo.toml or config.toml
func getProductionBaseURLFromHugoToml(sitePath string) (string, error) {
	// Try hugo.toml first
	hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
	if baseURL, err := extractBaseURLFromTomlFile(hugoTomlPath); err == nil && baseURL != "" {
		return baseURL, nil
	}

	// Try config.toml
	configTomlPath := filepath.Join(sitePath, "config.toml")
	if baseURL, err := extractBaseURLFromTomlFile(configTomlPath); err == nil && baseURL != "" {
		return baseURL, nil
	}

	return "", fmt.Errorf("baseURL not found in hugo.toml or config.toml")
}

// extractBaseURLFromTomlFile extracts baseURL from a TOML file
func extractBaseURLFromTomlFile(tomlPath string) (string, error) {
	content, err := os.ReadFile(tomlPath)
	if err != nil {
		return "", err
	}

	// Simple parsing to extract baseURL (handles various formats)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "baseURL") || strings.HasPrefix(line, "baseurl") {
			// Extract value after = sign
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				baseURL := strings.TrimSpace(parts[1])
				// Remove quotes
				baseURL = strings.Trim(baseURL, `"'`)
				// Skip placeholder values
				if baseURL != "" && !strings.Contains(baseURL, "example.") && !strings.Contains(baseURL, "localhost") {
					return baseURL, nil
				}
			}
		}
	}

	return "", fmt.Errorf("baseURL not found in %s", tomlPath)
}

// StopServe stops the running Hugo development server and cleans up localhost files
func (a *App) StopServe() bool {
	if a.serveCmd != nil && a.serveCmd.Process != nil {

		// Stop the server
		a.serveCmd.Process.Kill()
		a.serveCmd = nil
		a.serveSitePath = ""
		a.serverPort = 0
		return true
	}
	return false
}

// GetServerURL returns the current server URL
func (a *App) GetServerURL() string {
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

type SetStatusParams = api.SetStatusParams
type SetStatusResult = api.SetStatusResult

func (a *App) SetStatus(params SetStatusParams) SetStatusResult {
	return api.SetStatus(params)
}

// ====================
// Launch Wizard
// ====================

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

// AICreateSite creates a complete Hugo site with AI-generated content
func (a *App) AICreateSite(params AICreateSiteParams) AICreateSiteResult {
	// Create a progress handler that emits Wails runtime events
	progressHandler := func(event api.ProgressEvent) {
		// Emit progress event to frontend
		wruntime.EventsEmit(a.ctx, "ai:progress", map[string]interface{}{
			"phase":     event.Phase,
			"eventType": event.EventType,
			"message":   event.Message,
			"pagePath":  event.PagePath,
			"progress":  event.Progress,
			"current":   event.Current,
			"total":     event.Total,
		})
	}

	return api.AICreateSiteWithProgress(params, progressHandler)
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

// RemoveAICredentials removes AI credentials for a specific provider or all
func (a *App) RemoveAICredentials(provider string) error {
	if provider == "" {
		return api.CleanAIConfig()
	}
	// Remove specific provider - would need to add to api package
	return api.CleanAIConfig()
}

// ====================
// Open External
// ====================

// OpenInBrowser opens a URL in the default browser
func (a *App) OpenInBrowser(url string) {
	wruntime.BrowserOpenURL(a.ctx, url)
}

// OpenInFinder opens a path in Finder (macOS) or file explorer
func (a *App) OpenInFinder(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Run()
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

	// Read directory
	entries, err := os.ReadDir(dirPath)
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

		fullPath := filepath.Join(dirPath, entry.Name())

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
	Success    bool  `json:"success"`
	FileCount  int   `json:"fileCount"`
	FolderCount int  `json:"folderCount"`
	TotalSize  int64 `json:"totalSize"`
	Error      string `json:"error"`
}

// GetFolderStats recursively gets statistics for a folder
func (a *App) GetFolderStats(dirPath string) FolderStatsResult {
	result := FolderStatsResult{
		Success: false,
	}

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

	// Validate file exists
	info, err := os.Stat(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("File not accessible: %v", err)
		return result
	}

	if info.IsDir() {
		result.Error = "Path is a directory, not a file"
		return result
	}

	// Read file content
	content, err := os.ReadFile(filePath) // #nosec G304 - filePath comes from project directory
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

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create directory: %v", err)
		return result
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
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

	// Check if file/directory exists
	info, err := os.Stat(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("File not accessible: %v", err)
		return result
	}

	// Delete file or directory
	if info.IsDir() {
		if err := os.RemoveAll(filePath); err != nil {
			result.Error = fmt.Sprintf("Failed to delete directory: %v", err)
			return result
		}
	} else {
		if err := os.Remove(filePath); err != nil {
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

	// If file exists, find a unique name by adding (1), (2), etc.
	uniquePath := findUniquePath(filePath)
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

	// If directory exists, find a unique name by adding (1), (2), etc.
	uniquePath := findUniquePath(dirPath)
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

	// Check if source exists
	if _, err := os.Stat(oldPath); err != nil {
		result.Error = fmt.Sprintf("Source not found: %v", err)
		return result
	}

	// Check if destination already exists
	if _, err := os.Stat(newPath); err == nil {
		result.Error = "Destination already exists"
		return result
	}

	// Create destination directory if needed
	destDir := filepath.Dir(newPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		result.Error = fmt.Sprintf("Failed to create destination directory: %v", err)
		return result
	}

	// Move/rename the file or directory
	if err := os.Rename(oldPath, newPath); err != nil {
		result.Error = fmt.Sprintf("Failed to move: %v", err)
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

	// Calculate new path
	dir := filepath.Dir(oldPath)
	newPath := filepath.Join(dir, newName)
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

	info, err := os.Stat(path)
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

	depth, err := getDirectoryDepth(path)
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
