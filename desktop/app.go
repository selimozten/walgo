package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/selimozten/walgo/pkg/api"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx        context.Context
	serveCmd   *exec.Cmd // Track running Hugo server
	serverPort int
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ====================
// Window Controls (for frameless window)
// ====================

// Minimize minimizes the window
func (a *App) Minimize() {
	runtime.WindowMinimise(a.ctx)
}

// Maximize toggles maximize/restore
func (a *App) Maximize() {
	runtime.WindowToggleMaximise(a.ctx)
}

// Close closes the application
func (a *App) Close() {
	// Stop any running server first
	a.StopServe()
	runtime.Quit(a.ctx)
}

// ====================
// File Dialogs
// ====================

// SelectDirectory opens a directory picker dialog
func (a *App) SelectDirectory(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// SelectFile opens a file picker dialog
func (a *App) SelectFile(title string, filters []string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
}

// ====================
// Version
// ====================

// VersionResult holds version information
type VersionResult = api.VersionResult

// GetVersion returns version information
func (a *App) GetVersion() VersionResult {
	return api.GetVersion()
}

// ====================
// Site Management
// ====================

// CreateSite initializes a new Hugo site with Walrus configuration
func (a *App) CreateSite(parentDir string, name string) error {
	return api.CreateSite(parentDir, name)
}

// BuildSite builds the Hugo site at the given path
func (a *App) BuildSite(sitePath string) error {
	return api.BuildSite(sitePath)
}

// DeployResult holds the result of a deployment
type DeployResult struct {
	Success  bool   `json:"success"`
	ObjectID string `json:"objectId"`
	Error    string `json:"error"`
}

// DeploySite deploys the site to Walrus
func (a *App) DeploySite(sitePath string, epochs int) DeployResult {
	res := api.DeploySite(sitePath, epochs)
	return DeployResult{
		Success:  res.Success,
		ObjectID: res.ObjectID,
		Error:    res.Error,
	}
}

// OptimizeSite optimizes HTML, CSS, and JS files
func (a *App) OptimizeSite(sitePath string) error {
	return api.OptimizeSite(sitePath)
}

// CompressSite compresses files with Brotli
func (a *App) CompressSite(sitePath string) error {
	return api.CompressSite(sitePath)
}

// ====================
// AI Features
// ====================

// AIConfigureParams holds AI configuration parameters
type AIConfigureParams = api.AIConfigureParams
type AIConfigResult = api.AIConfigResult

// ConfigureAI sets up AI provider credentials
func (a *App) ConfigureAI(params AIConfigureParams) error {
	return api.ConfigureAI(params)
}

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

// GenerateContentParams holds content generation parameters
type GenerateContentParams struct {
	SitePath    string `json:"sitePath"`
	ContentType string `json:"contentType"` // "post" or "page"
	Topic       string `json:"topic"`
	Context     string `json:"context"`
}

// GenerateContentResult holds the result of content generation
type GenerateContentResult struct {
	Success  bool   `json:"success"`
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
	Error    string `json:"error"`
}

// GenerateContent creates new content using AI
func (a *App) GenerateContent(params GenerateContentParams) GenerateContentResult {
	res := api.GenerateContent(api.GenerateContentParams(params))
	return GenerateContentResult{
		Success:  res.Success,
		FilePath: res.FilePath,
		Content:  res.Content,
		Error:    res.Error,
	}
}

// UpdateContentParams holds content update parameters
type UpdateContentParams struct {
	FilePath     string `json:"filePath"`
	Instructions string `json:"instructions"`
}

// UpdateContentResult holds the result of content update
type UpdateContentResult struct {
	Success        bool   `json:"success"`
	UpdatedContent string `json:"updatedContent"`
	Error          string `json:"error"`
}

// UpdateContent updates existing content using AI
func (a *App) UpdateContent(params UpdateContentParams) UpdateContentResult {
	res := api.UpdateContent(api.UpdateContentParams(params))
	return UpdateContentResult{
		Success:        res.Success,
		UpdatedContent: res.UpdatedContent,
		Error:          res.Error,
	}
}

// ====================
// Projects Management
// ====================

// Project represents a Walrus site project
type Project = api.Project

// ListProjects returns all projects
func (a *App) ListProjects() ([]Project, error) {
	return api.ListProjects()
}

// GetProject returns a single project by ID
func (a *App) GetProject(projectID int64) (*Project, error) {
	return api.GetProject(projectID)
}

// DeleteProject deletes a project by ID
func (a *App) DeleteProject(projectID int64) error {
	return api.DeleteProject(projectID)
}

// ====================
// Import
// ====================

// ImportObsidianParams holds import parameters
type ImportObsidianParams struct {
	SitePath      string `json:"sitePath"`
	VaultPath     string `json:"vaultPath"`
	IncludeDrafts bool   `json:"includeDrafts"`
	AttachmentDir string `json:"attachmentDir"`
}

// ImportObsidianResult holds import results
type ImportObsidianResult struct {
	Success       bool   `json:"success"`
	FilesImported int    `json:"filesImported"`
	Error         string `json:"error"`
}

// ImportObsidian imports content from Obsidian vault
func (a *App) ImportObsidian(params ImportObsidianParams) ImportObsidianResult {
	res := api.ImportObsidian(api.ImportObsidianParams(params))
	return ImportObsidianResult{
		Success:       res.Success,
		FilesImported: res.FilesImported,
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
	return api.QuickStart(api.QuickStartParams(params))
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

	// Start the server and track the command
	result := api.Serve(api.ServeParams(params))
	if result.Success {
		a.serverPort = params.Port
		if a.serverPort == 0 {
			a.serverPort = 1313
		}
	}
	return result
}

// StopServe stops the running Hugo development server
func (a *App) StopServe() bool {
	if a.serveCmd != nil && a.serveCmd.Process != nil {
		a.serveCmd.Process.Kill()
		a.serveCmd = nil
		return true
	}
	return false
}

// IsServing returns true if a server is running
func (a *App) IsServing() bool {
	return a.serveCmd != nil && a.serveCmd.Process != nil
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
	return api.NewContent(api.NewContentParams(params))
}

// ====================
// Update Deployment
// ====================

// UpdateDeploymentParams holds update parameters
type UpdateDeploymentParams = api.UpdateDeploymentParams
type UpdateDeploymentResult = api.UpdateDeploymentResult

// UpdateDeployment updates an existing Walrus Site
func (a *App) UpdateDeployment(params UpdateDeploymentParams) UpdateDeploymentResult {
	return api.UpdateDeployment(api.UpdateDeploymentParams(params))
}

// ====================
// Doctor
// ====================

// DoctorResult holds diagnostic results
type DoctorResult = api.DoctorResult
type DoctorCheck = api.DoctorCheck
type DoctorSummary = api.DoctorSummary

// Doctor runs environment diagnostics
func (a *App) Doctor() DoctorResult {
	return api.Doctor()
}

// ====================
// Status
// ====================

// StatusResult holds status check results
type StatusResult = api.StatusResult

// Status checks the status of a deployed site
func (a *App) Status(objectID string) StatusResult {
	return api.Status(objectID)
}

// ====================
// Project Metadata Operations
// ====================

// EditProjectParams holds project edit parameters
type EditProjectParams = api.EditProjectParams
type EditProjectResult = api.EditProjectResult

// EditProject updates project metadata
func (a *App) EditProject(params EditProjectParams) EditProjectResult {
	return api.EditProject(api.EditProjectParams(params))
}

// ArchiveProjectResult holds archive results
type ArchiveProjectResult = api.ArchiveProjectResult

// ArchiveProject archives a project
func (a *App) ArchiveProject(projectID int64) ArchiveProjectResult {
	return api.ArchiveProject(projectID)
}

// ====================
// Launch Wizard
// ====================

// LaunchStep represents a step in the launch wizard
type LaunchStep = api.LaunchStep
type LaunchWizardParams = api.LaunchWizardParams
type LaunchWizardResult = api.LaunchWizardResult

// LaunchWizard executes full launch wizard flow
func (a *App) LaunchWizard(params LaunchWizardParams) LaunchWizardResult {
	return api.LaunchWizard(api.LaunchWizardParams(params))
}

// ====================
// AI Create Site
// ====================

// AICreateSiteParams holds AI site creation parameters
type AICreateSiteParams = api.AICreateSiteParams
type AICreateSiteResult = api.AICreateSiteResult

// AICreateSite creates a complete Hugo site with AI-generated content
func (a *App) AICreateSite(params AICreateSiteParams) AICreateSiteResult {
	return api.AICreateSite(api.AICreateSiteParams(params))
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

// GetAICredentials returns the current AI provider credentials (masked)
func (a *App) GetAICredentials() AIConfigResult {
	result, err := api.GetAIConfig()
	if err != nil {
		return AIConfigResult{
			Success: false,
			Error:   err.Error(),
		}
	}
	return *result
}

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
	runtime.BrowserOpenURL(a.ctx, url)
}

// OpenInFinder opens a path in Finder (macOS) or file explorer
func (a *App) OpenInFinder(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Run()
}

// OpenInEditor opens a file in the default editor
func (a *App) OpenInEditor(filePath string) error {
	cmd := exec.Command("open", "-t", filePath)
	return cmd.Run()
}
