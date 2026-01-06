package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/selimozten/walgo/internal/ai"
	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/deployment"
	"github.com/selimozten/walgo/internal/deps"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/obsidian"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/utils"
	"github.com/selimozten/walgo/internal/version"
	"github.com/spf13/viper"
)

// =============================================================================
// Progress Handler Types (Public wrappers for desktop app)
// =============================================================================

// ProgressEvent represents a progress event from AI operations
type ProgressEvent struct {
	Phase     string  `json:"phase"`
	EventType string  `json:"eventType"`
	Message   string  `json:"message"`
	PagePath  string  `json:"pagePath,omitempty"`
	Progress  float64 `json:"progress"`
	Current   int     `json:"current"`
	Total     int     `json:"total"`
}

// ProgressHandler is a callback function for progress updates
type ProgressHandler func(event ProgressEvent)

// =============================================================================
// Utility Functions
// =============================================================================

// GetDefaultSitesDir returns the default walgo-sites directory in user's home
func GetDefaultSitesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	sitesDir := filepath.Join(homeDir, "walgo-sites")
	if err := os.MkdirAll(sitesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create walgo-sites directory: %w", err)
	}

	return sitesDir, nil
}

// saveDraftProject saves a newly created site as a draft project
// This allows users to manage and deploy the site later
func saveDraftProject(siteName, sitePath string) error {
	manager, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create project manager: %w", err)
	}
	defer manager.Close()

	// Create draft project
	err = manager.CreateDraftProject(siteName, sitePath)
	if err != nil {
		return fmt.Errorf("failed to create draft project: %w", err)
	}
	return nil
}

// =============================================================================
// Wallet Management
// =============================================================================

// WalletInfo holds wallet information
type WalletInfo struct {
	Address    string  `json:"address"`
	SuiBalance float64 `json:"suiBalance"`
	WalBalance float64 `json:"walBalance"`
	Network    string  `json:"network"`
	Active     bool    `json:"active"`
}

// GetWalletInfo returns current wallet information
func GetWalletInfo() (*WalletInfo, error) {
	// Get active network
	network, err := sui.GetActiveEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to get active env: %w", err)
	}

	// Get active address
	address, err := sui.GetActiveAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get active address: %w", err)
	}

	// Get balance
	balance, err := sui.GetBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return &WalletInfo{
		Address:    address,
		SuiBalance: balance.SUI,
		WalBalance: balance.WAL,
		Network:    network,
		Active:     true,
	}, nil
}

// AddressListResult holds list of addresses
type AddressListResult struct {
	Addresses []string `json:"addresses"`
	Error     string   `json:"error"`
}

// GetAddressList returns list of all wallet addresses
func GetAddressList() AddressListResult {
	addresses, err := sui.GetAddressList()
	if err != nil {
		return AddressListResult{Error: fmt.Sprintf("failed to get addresses: %v", err)}
	}
	return AddressListResult{Addresses: addresses}
}

// SwitchAddressParams holds parameters for switching address
type SwitchAddressParams struct {
	Address string `json:"address"`
}

// SwitchAddressResult holds result of switching address
type SwitchAddressResult struct {
	Success bool   `json:"success"`
	Address string `json:"address"`
	Error   string `json:"error"`
}

// SwitchAddress switches to a different wallet address
func SwitchAddress(address string) SwitchAddressResult {
	err := sui.SwitchAddress(address)
	if err != nil {
		return SwitchAddressResult{Error: fmt.Sprintf("failed to switch address: %v", err)}
	}
	return SwitchAddressResult{Success: true, Address: address}
}

// CreateAddressParams holds parameters for creating address
type CreateAddressParams struct {
	KeyScheme string `json:"keyScheme"` // ed25519, secp256k1, secp256r1
}

// CreateAddressResult holds result of creating address
type CreateAddressResult struct {
	Success        bool   `json:"success"`
	Address        string `json:"address"`
	Alias          string `json:"alias"`
	RecoveryPhrase string `json:"recoveryPhrase"`
	Error          string `json:"error"`
}

// CreateAddress creates a new wallet address
// keyScheme: ed25519, secp256k1, or secp256r1
// alias: optional alias for the address (can be empty string)
func CreateAddress(keyScheme string, alias string) CreateAddressResult {
	if keyScheme == "" {
		keyScheme = "ed25519" // default
	}

	result, err := sui.CreateAddressWithDetails(keyScheme, alias)
	if err != nil {
		return CreateAddressResult{Error: fmt.Sprintf("failed to create address: %v", err)}
	}

	return CreateAddressResult{
		Success:        true,
		Address:        result.Address,
		Alias:          result.Alias,
		RecoveryPhrase: result.RecoveryPhrase,
	}
}

// ImportAddressParams holds import parameters
type ImportAddressParams struct {
	Method    string `json:"method"`    // "mnemonic" or "key"
	KeyScheme string `json:"keyScheme"` // "ed25519", "secp256k1", "secp256r1"
	Input     string `json:"input"`     // Mnemonic phrase or private key
}

// ImportAddressResult holds import result
type ImportAddressResult struct {
	Success bool   `json:"success"`
	Address string `json:"address"`
	Error   string `json:"error"`
}

// ImportAddress imports a wallet address with provided mnemonic or private key
func ImportAddress(params ImportAddressParams) ImportAddressResult {
	if params.Input == "" {
		return ImportAddressResult{
			Success: false,
			Error:   "Input (mnemonic or private key) is required",
		}
	}

	method := sui.ImportFromMnemonic
	if params.Method == "key" || params.Method == "private-key" {
		method = sui.ImportFromPrivateKey
	}

	keyScheme := params.KeyScheme
	if keyScheme == "" {
		keyScheme = "ed25519"
	}

	address, err := sui.ImportAddressWithInput(method, keyScheme, params.Input)
	if err != nil {
		return ImportAddressResult{
			Success: false,
			Error:   fmt.Sprintf("failed to import address: %v", err),
		}
	}

	return ImportAddressResult{
		Success: true,
		Address: address,
	}
}

// SwitchNetworkResult holds result of switching network
type SwitchNetworkResult struct {
	Success bool   `json:"success"`
	Network string `json:"network"`
	Error   string `json:"error"`
}

// SwitchNetwork switches to a different network (testnet/mainnet)
func SwitchNetwork(network string) SwitchNetworkResult {
	err := sui.SwitchEnv(network)
	if err != nil {
		return SwitchNetworkResult{Error: fmt.Sprintf("failed to switch network: %v", err)}
	}
	return SwitchNetworkResult{Success: true, Network: network}
}

// =============================================================================
// Site Management
// =============================================================================

// InitSiteResult holds init site result
type InitSiteResult struct {
	Success  bool   `json:"success"`
	SitePath string `json:"sitePath"`
	Error    string `json:"error"`
}

// InitSite initializes a new Hugo site in current directory (like 'walgo init')
func InitSite(parentDir string, siteName string) InitSiteResult {
	// Keep original site name for database
	originalSiteName := siteName

	// Sanitize site name ONLY for directory creation
	sanitizedDirName := utils.SanitizeSiteName(siteName)

	// Use default walgo-sites directory if parentDir is empty
	if parentDir == "" {
		defaultDir, err := GetDefaultSitesDir()
		if err != nil {
			return InitSiteResult{
				Success: false,
				Error:   fmt.Sprintf("cannot determine walgo-sites directory: %v", err),
			}
		}
		parentDir = defaultDir
	}

	// Create site path using sanitized directory name
	sitePath, err := filepath.Abs(filepath.Join(parentDir, sanitizedDirName))
	if err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("invalid site path: %v", err),
		}
	}

	// Create site directory
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create site directory: %v", err),
		}
	}

	// Initialize Hugo site
	if err := hugo.InitializeSite(sitePath); err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("failed to initialize Hugo site: %v", err),
		}
	}

	// Create Walgo configuration
	if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create walgo.yaml: %v", err),
		}
	}

	if err := BuildSite(sitePath); err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("failed to build site: %v", err),
		}
	}

	// Save as draft project using ORIGINAL name
	if err := saveDraftProject(originalSiteName, sitePath); err != nil {
		return InitSiteResult{
			Success: false,
			Error:   fmt.Sprintf("failed to save draft project: %v", err),
		}
	}

	return InitSiteResult{
		Success:  true,
		SitePath: sitePath,
	}
}

// BuildSite builds the Hugo site at the given path
func BuildSite(sitePath string) error {
	// Change to site directory to ensure Hugo finds everything
	if err := os.Chdir(sitePath); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// Load config to ensure it exists and is valid
	viper.Reset()
	viper.SetConfigFile(filepath.Join(sitePath, "walgo.yaml"))
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read walgo.yaml: %w", err)
	}

	if err := hugo.BuildSite(sitePath); err != nil {
		return fmt.Errorf("hugo build failed: %w", err)
	}

	return nil
}

// NewContentParams holds parameters for creating new content
type NewContentParams struct {
	SitePath    string `json:"sitePath,omitempty"`
	Slug        string `json:"slug"`
	ContentType string `json:"contentType"`
	NoBuild     bool   `json:"noBuild"`
	Serve       bool   `json:"serve"`
}

// NewContentResult holds the result of creating new content
type NewContentResult struct {
	Success  bool   `json:"success"`
	Path     string `json:"path"`
	FilePath string `json:"filePath"` // Alias for Path (frontend compatibility)
	Error    string `json:"error"`
}

// NewContent creates new content in Hugo site
func NewContent(params NewContentParams) NewContentResult {
	sitePath := params.SitePath
	if sitePath == "" {
		var err error
		sitePath, err = os.Getwd()
		if err != nil {
			return NewContentResult{Error: fmt.Sprintf("cannot determine current directory: %v", err)}
		}
	}

	// Detect available content types
	defaultType := hugo.GetDefaultContentType(sitePath)

	// Determine content type
	selectedType := params.ContentType
	if selectedType == "" {
		selectedType = defaultType
	}

	// Get slug
	slug := params.Slug
	if slug == "" {
		return NewContentResult{Error: "slug is required"}
	}

	// Validate slug
	if !isValidSlug(slug) {
		return NewContentResult{Error: "invalid slug: use only letters, numbers, hyphens, and underscores"}
	}

	// Ensure .md extension
	if !strings.HasSuffix(slug, ".md") {
		slug += ".md"
	}

	// Build content path
	contentPath := filepath.Join(selectedType, slug)

	// Create content using Hugo
	if err := hugo.CreateContent(sitePath, contentPath); err != nil {
		return NewContentResult{Error: fmt.Sprintf("failed to create content: %v", err)}
	}

	createdFilePath := filepath.Join(sitePath, "content", contentPath)

	if err := BuildSite(sitePath); err != nil {
		return NewContentResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	return NewContentResult{
		Success:  true,
		Path:     createdFilePath,
		FilePath: createdFilePath,
	}
}

// isValidSlug validates slug format
func isValidSlug(slug string) bool {
	slug = strings.TrimSuffix(slug, ".md")
	validSlug := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return validSlug.MatchString(slug) && len(slug) > 0 && len(slug) < 100
}

// =============================================================================
// Diagnostics & Status
// =============================================================================

// VersionResult holds version information
type VersionResult struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
}

// GetVersion returns current version information
func GetVersion() VersionResult {
	return VersionResult{
		Version:   "0.3.0",
		GitCommit: "dev",
		BuildDate: "unknown",
	}
}

// CheckUpdatesResult holds update check results
type CheckUpdatesResult struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	UpdateURL      string `json:"updateUrl"`
}

// CheckUpdates checks for available updates
func CheckUpdates() CheckUpdatesResult {
	const githubReleasesAPI = "https://api.github.com/repos/selimozten/walgo/releases/latest"

	version := GetVersion()

	result := CheckUpdatesResult{
		CurrentVersion: version.Version,
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(githubReleasesAPI)
	if err != nil {
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return result
	}

	body, _ := io.ReadAll(resp.Body)

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return result
	}

	result.LatestVersion = strings.TrimPrefix(release.TagName, "v")
	result.UpdateURL = release.HTMLURL

	return result
}

// SystemHealth holds current system health status
type SystemHealth struct {
	NetOnline       bool   `json:"netOnline"`
	SuiInstalled    bool   `json:"suiInstalled"`
	SuiConfigured   bool   `json:"suiConfigured"`
	WalrusInstalled bool   `json:"walrusInstalled"`
	SiteBuilder     bool   `json:"siteBuilder"`
	HugoInstalled   bool   `json:"hugoInstalled"`
	Message         string `json:"message"`
}

// checkNetworkConnectivity checks internet connectivity using multiple endpoints
// Returns true if at least one endpoint is reachable
func checkNetworkConnectivity() bool {
	// Use multiple reliable endpoints to avoid rate limiting and false negatives
	endpoints := []string{
		"https://www.google.com",
		"https://1.1.1.1", // Cloudflare DNS
		"https://8.8.8.8", // Google DNS
	}

	client := &http.Client{
		Timeout: 3 * time.Second, // Short timeout to avoid blocking
	}

	// Try each endpoint, return true on first success
	for _, endpoint := range endpoints {
		resp, err := client.Get(endpoint)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return true
			}
		}
	}

	return false
}

// =============================================================================
// QuickStart
// =============================================================================

// QuickStartParams holds quickstart parameters
type QuickStartParams struct {
	ParentDir string `json:"parentDir,omitempty"`
	SiteName  string `json:"siteName"`
	Name      string `json:"name,omitempty"` // Alias for SiteName (frontend compatibility)
	SkipBuild bool   `json:"skipBuild"`
}

// QuickStartResult holds quickstart result
type QuickStartResult struct {
	Success  bool   `json:"success"`
	SitePath string `json:"sitePath"`
	Error    string `json:"error"`
}

// QuickStart creates a new Hugo site with quickstart flow
func QuickStart(params QuickStartParams) QuickStartResult {
	siteName := params.SiteName
	if siteName == "" {
		siteName = params.Name // Use Name as fallback
	}
	parentDir := params.ParentDir
	if parentDir == "" {
		// Use default walgo-sites directory in home
		defaultDir, err := GetDefaultSitesDir()
		if err != nil {
			return QuickStartResult{Error: fmt.Sprintf("cannot determine walgo-sites directory: %v", err)}
		}
		parentDir = defaultDir
	}

	// Sanitize site name ONLY for directory/path creation
	originalSiteName := siteName
	sanitizedDirName := utils.SanitizeSiteName(siteName)
	if sanitizedDirName != originalSiteName {
		fmt.Printf("Directory name sanitized: '%s' -> '%s'\n", originalSiteName, sanitizedDirName)
	}

	// Create site directory using sanitized name
	sitePath, err := filepath.Abs(filepath.Join(parentDir, sanitizedDirName))
	if err != nil {
		return QuickStartResult{Error: fmt.Sprintf("invalid site path: %v", err)}
	}

	if err := os.MkdirAll(sitePath, 0755); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to create site directory: %v", err)}
	}

	// Initialize Hugo site
	if err := hugo.InitializeSite(sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to initialize Hugo site: %v", err)}
	}

	// Create walgo.yaml config
	if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to create walgo.yaml: %v", err)}
	}

	// Setup site - use ORIGINAL name for Hugo config
	siteType := hugo.SiteTypeBlog
	if err := hugo.SetupSiteConfigWithName(sitePath, siteType, originalSiteName); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to set up config: %v", err)}
	}

	if err := hugo.SetupArchetypes(sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to set up archetypes: %v", err)}
	}

	if err := hugo.SetupFavicon(sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to set up favicon: %v", err)}
	}

	// Install theme
	_ = hugo.GetThemeInfo(siteType)
	if err := hugo.InstallTheme(sitePath, siteType); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to install theme: %v", err)}
	}

	// Create sample content
	contentDir := filepath.Join(sitePath, "content", "posts")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to create content directory: %v", err)}
	}

	// Create welcome.md sample post
	welcomePath := filepath.Join(contentDir, "welcome.md")
	welcomeContent := `---
title: "Welcome to Walrus Sites"
date: 2024-01-01T00:00:00Z
draft: false
---

Welcome to your new decentralized website powered by Walrus!

This site is hosted on the Walrus decentralized storage network, making it censorship-resistant and always available.

## Next Steps

1. Edit this content in ` + "`content/posts/welcome.md`" + `
2. Add more posts to ` + "`content/posts/`" + `
3. Customize your theme
4. Deploy with ` + "`walgo launch`" + `

Happy building! 🚀
`
	if err := os.WriteFile(welcomePath, []byte(welcomeContent), 0644); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to create welcome post: %v", err)}
	}

	if err := BuildSite(sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	// Save as draft project for later deployment - use ORIGINAL name
	if err := saveDraftProject(originalSiteName, sitePath); err != nil {
		return QuickStartResult{Error: fmt.Sprintf("failed to save draft project: %v", err)}
	}

	return QuickStartResult{
		Success:  true,
		SitePath: sitePath,
	}
}

// =============================================================================
// Serve
// =============================================================================

// ServeParams holds serve parameters
type ServeParams struct {
	SitePath string `json:"sitePath"`
	Port     int    `json:"port"`
	Drafts   bool   `json:"drafts"`
	Expired  bool   `json:"expired"` // Include expired content
	Future   bool   `json:"future"`
}

// ServeResult holds serve result
type ServeResult struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Error   string `json:"error"`
}

// Serve starts local Hugo development server
func Serve(params ServeParams) ServeResult {
	sitePath := params.SitePath
	if sitePath == "" {
		return ServeResult{Error: "site path is required"}
	}

	if _, err := exec.LookPath("hugo"); err != nil {
		return ServeResult{Error: "hugo is not installed or not found in PATH"}
	}

	if err := hugo.ServeSite(sitePath); err != nil {
		return ServeResult{Error: fmt.Sprintf("failed to serve site: %v", err)}
	}

	return ServeResult{
		Success: true,
		URL:     "http://localhost:1313",
	}
}

// =============================================================================
// Projects Management
// =============================================================================

// Project represents a Walrus site project
type Project struct {
	ID           int64              `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Category     string             `json:"category"`
	ObjectID     string             `json:"objectId"`
	Network      string             `json:"network"`
	WalletAddr   string             `json:"wallet"`
	SitePath     string             `json:"sitePath"`
	ImageURL     string             `json:"imageUrl"`
	SuiNS        string             `json:"suins"`
	CreatedAt    string             `json:"createdAt"`
	UpdatedAt    string             `json:"updatedAt"`
	LastDeployAt string             `json:"lastDeployAt"`
	Epochs       int                `json:"epochs"`
	Status       string             `json:"status"`
	DeployCount  int                `json:"deployCount"`
	Deployments  []DeploymentRecord `json:"deployments,omitempty"`
	// Tool status for deployment
	SuiReady    bool `json:"suiReady,omitempty"`
	WalrusReady bool `json:"walrusReady,omitempty"`
	SiteBuilder bool `json:"siteBuilder,omitempty"`
	HugoReady   bool `json:"hugoReady,omitempty"`
}

// DeploymentRecord represents a single deployment of a project
type DeploymentRecord struct {
	ID        int64  `json:"id"`
	ProjectID int64  `json:"projectId"`
	ObjectID  string `json:"objectId"`
	Network   string `json:"network"`
	Epochs    int    `json:"epochs"`
	GasFee    string `json:"gasFee"`
	Version   string `json:"version,omitempty"`
	Notes     string `json:"notes,omitempty"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// ListProjects returns all projects
func ListProjects() ([]Project, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}
	defer pm.Close()

	projs, err := pm.ListProjects("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	// Convert to API Project type
	result := make([]Project, len(projs))
	for i, p := range projs {
		// Check tool status
		suiReady, _ := deps.LookPath("sui")
		walrusReady, _ := deps.LookPath("walrus")
		siteBuilder, _ := deps.LookPath("site-builder")
		hugoReady, _ := deps.LookPath("hugo")

		result[i] = Project{
			ID:           p.ID,
			Name:         p.Name,
			Description:  p.Description,
			Category:     p.Category,
			ObjectID:     p.ObjectID,
			Network:      p.Network,
			WalletAddr:   p.WalletAddr,
			SitePath:     p.SitePath,
			ImageURL:     p.ImageURL,
			SuiNS:        p.SuiNS,
			CreatedAt:    p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    p.UpdatedAt.Format(time.RFC3339),
			LastDeployAt: p.LastDeployAt.Format(time.RFC3339),
			Epochs:       p.Epochs,
			Status:       p.Status,
			DeployCount:  p.DeployCount,
			// Tool status
			SuiReady:    suiReady != "",
			WalrusReady: walrusReady != "",
			SiteBuilder: siteBuilder != "",
			HugoReady:   hugoReady != "",
		}
	}

	return result, nil
}

// GetProject returns a single project by ID with deployment history
func GetProject(projectID int64) (*Project, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create project manager: %w", err)
	}
	defer pm.Close()

	proj, err := pm.GetProject(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	if proj == nil {
		return nil, fmt.Errorf("project not found")
	}

	// Get deployment records
	deploymentRecords, err := pm.GetProjectDeployments(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployments: %w", err)
	}

	// Convert deployment records to API format
	deployments := make([]DeploymentRecord, 0, len(deploymentRecords))
	for _, dr := range deploymentRecords {
		deployments = append(deployments, DeploymentRecord{
			ID:        dr.ID,
			ProjectID: dr.ProjectID,
			ObjectID:  dr.ObjectID,
			Network:   dr.Network,
			Epochs:    dr.Epochs,
			GasFee:    dr.GasFee,
			Version:   dr.Version,
			Notes:     dr.Notes,
			Success:   dr.Success,
			Error:     dr.Error,
			CreatedAt: dr.CreatedAt.Format(time.RFC3339),
		})
	}

	return &Project{
		ID:           proj.ID,
		Name:         proj.Name,
		Description:  proj.Description,
		Category:     proj.Category,
		ObjectID:     proj.ObjectID,
		Network:      proj.Network,
		WalletAddr:   proj.WalletAddr,
		SitePath:     proj.SitePath,
		ImageURL:     proj.ImageURL,
		SuiNS:        proj.SuiNS,
		CreatedAt:    proj.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    proj.UpdatedAt.Format(time.RFC3339),
		LastDeployAt: proj.LastDeployAt.Format(time.RFC3339),
		Epochs:       proj.Epochs,
		Status:       proj.Status,
		DeployCount:  proj.DeployCount,
		Deployments:  deployments,
	}, nil
}

// DeleteProjectParams holds delete project parameters
type DeleteProjectParams struct {
	ProjectID int64 `json:"projectId"`
}

// DeleteProjectResult holds delete result
type DeleteProjectResult struct {
	Success          bool   `json:"success"`
	Message          string `json:"message"`
	Error            string `json:"error"`
	OnChainDestroyed bool   `json:"onChainDestroyed"`
}

// DeleteProject deletes a project by ID (includes on-chain destruction if objectId exists)
func DeleteProject(params DeleteProjectParams) DeleteProjectResult {
	result := DeleteProjectResult{
		Success: false,
	}

	pm, err := projects.NewManager()
	if err != nil {
		result.Error = fmt.Sprintf("failed to create project manager: %v", err)
		return result
	}
	defer pm.Close()

	// Get project details
	proj, err := pm.GetProject(params.ProjectID)
	if err != nil {
		result.Error = fmt.Sprintf("failed to get project: %v", err)
		return result
	}

	// If project has an object ID, destroy it on-chain first
	if proj.ObjectID != "" {
		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := d.Destroy(ctx, proj.ObjectID); err != nil {
			// Log warning but continue with local deletion
			result.Message = fmt.Sprintf("Warning: Failed to destroy site on-chain: %v. Continuing with local deletion.", err)
		} else {
			result.OnChainDestroyed = true
		}
	}

	// Delete from local database
	if err := pm.DeleteProject(params.ProjectID); err != nil {
		result.Error = fmt.Sprintf("failed to delete project: %v", err)
		return result
	}

	result.Success = true
	if result.OnChainDestroyed {
		result.Message = "Project deleted successfully (including on-chain destruction)"
	} else {
		result.Message = "Project deleted successfully from local database"
	}

	return result
}

// ProjectNameExists checks if a project with the given name already exists
func ProjectNameExists(name string) (bool, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return false, fmt.Errorf("failed to create project manager: %w", err)
	}
	defer pm.Close()

	exists, err := pm.ProjectNameExists(name)
	if err != nil {
		return false, fmt.Errorf("failed to check project name: %w", err)
	}

	return exists, nil
}

// EditProjectParams holds project edit parameters
type EditProjectParams struct {
	ProjectID   int64  `json:"projectId"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	SuiNS       string `json:"suins"`
}

// EditProjectResult holds edit result
type EditProjectResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// EditProject updates project metadata
func EditProject(params EditProjectParams) EditProjectResult {
	pm, err := projects.NewManager()
	if err != nil {
		return EditProjectResult{Error: fmt.Sprintf("failed to create project manager: %v", err)}
	}
	defer pm.Close()

	proj, err := pm.GetProject(params.ProjectID)
	if err != nil || proj == nil {
		return EditProjectResult{Error: "project not found"}
	}

	// Update fields if provided
	if params.Name != "" {
		proj.Name = params.Name
	}
	if params.Category != "" {
		proj.Category = params.Category
	}
	if params.Description != "" {
		proj.Description = params.Description
	}
	if params.ImageURL != "" {
		proj.ImageURL = params.ImageURL
	}
	if params.SuiNS != "" {
		proj.SuiNS = params.SuiNS
	}

	if err := pm.UpdateProject(proj); err != nil {
		return EditProjectResult{Error: fmt.Sprintf("failed to update project: %v", err)}
	}

	// Update ws-resources.json in the publish directory (if it exists)
	walgoCfg, err := config.LoadConfigFrom(proj.SitePath)
	if err != nil {
		// Config not found - skip ws-resources.json update (not critical for API)
		return EditProjectResult{
			Success: true,
			Message: "project updated successfully (ws-resources.json not updated - run 'walgo build' first)",
		}
	}

	publishDir := filepath.Join(proj.SitePath, walgoCfg.HugoConfig.PublishDir)
	wsResourcesPath := filepath.Join(publishDir, "ws-resources.json")

	// Check if publish directory exists
	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		// Publish directory doesn't exist - skip ws-resources.json update
		return EditProjectResult{
			Success: true,
			Message: "project updated successfully (ws-resources.json not updated - run 'walgo build' first)",
		}
	}

	// Update ws-resources.json with metadata, preserving ObjectID
	metadataOpts := compress.MetadataOptions{
		SiteName:    proj.Name,
		Description: proj.Description,
		ImageURL:    proj.ImageURL,
		ProjectURL:  compress.DefaultProjectURL,
		Creator:     compress.DefaultCreator,
		Link:        compress.DefaultLink,
		Category:    proj.Category,
	}

	// Preserve existing ObjectID
	if proj.ObjectID != "" {
		metadataOpts.ObjectID = proj.ObjectID
	}

	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		// For API, treat this as non-fatal but inform the user
		return EditProjectResult{
			Success: true,
			Message: fmt.Sprintf("project updated successfully, but ws-resources.json update failed: %v", err),
		}
	}

	return EditProjectResult{
		Success: true,
		Message: "project updated successfully",
	}
}

// ArchiveProjectResult holds archive result
type ArchiveProjectResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// ArchiveProject archives a project
func ArchiveProject(projectID int64) ArchiveProjectResult {
	pm, err := projects.NewManager()
	if err != nil {
		return ArchiveProjectResult{Error: fmt.Sprintf("failed to create project manager: %v", err)}
	}
	defer pm.Close()

	if err := pm.ArchiveProject(projectID); err != nil {
		return ArchiveProjectResult{Error: fmt.Sprintf("failed to archive project: %v", err)}
	}

	return ArchiveProjectResult{
		Success: true,
		Message: "project archived successfully",
	}
}

// SetStatusParams holds parameters for setting project status
type SetStatusParams struct {
	ProjectID int    `json:"projectId"`
	Status    string `json:"status"` // "draft", "active", or "archived"
}

// SetStatusResult holds the result of setting project status
type SetStatusResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// SetStatus sets the status of a project
func SetStatus(params SetStatusParams) SetStatusResult {
	pm, err := projects.NewManager()
	if err != nil {
		return SetStatusResult{Error: fmt.Sprintf("failed to create project manager: %v", err)}
	}
	defer pm.Close()

	if err := pm.SetStatus(params.ProjectID, params.Status); err != nil {
		return SetStatusResult{Error: fmt.Sprintf("failed to set status: %v", err)}
	}

	return SetStatusResult{
		Success: true,
		Message: fmt.Sprintf("project status set to %s successfully", params.Status),
	}
}

// =============================================================================
// Import (Obsidian)
// =============================================================================

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
func ImportObsidian(params ImportObsidianParams) ImportObsidianResult {
	// Validate vault path
	absVaultPath, err := filepath.Abs(params.VaultPath)
	if err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("invalid vault path: %v", err)}
	}
	params.VaultPath = filepath.Clean(absVaultPath)

	// Verify vault exists
	if _, err := os.Stat(params.VaultPath); os.IsNotExist(err) {
		return ImportObsidianResult{Error: fmt.Sprintf("vault path does not exist: %s", params.VaultPath)}
	}

	// Determine site name
	siteName := params.SiteName
	if siteName == "" {
		siteName = filepath.Base(params.VaultPath)
	}

	// Keep original site name for database
	originalSiteName := siteName

	// Sanitize site name ONLY for directory creation
	sanitizedDirName := utils.SanitizeSiteName(siteName)

	// Determine parent directory
	parentDir := params.ParentDir
	if parentDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return ImportObsidianResult{Error: fmt.Sprintf("cannot determine current directory: %v", err)}
		}
		parentDir = wd
	}

	// Use sanitized name for directory path
	sitePath := filepath.Join(parentDir, sanitizedDirName)

	// Check if site already exists
	if _, err := os.Stat(sitePath); err == nil {
		return ImportObsidianResult{Error: fmt.Sprintf("site directory already exists: %s", sitePath)}
	}

	// Step 1: Create site directory
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to create site directory: %v", err)}
	}

	// Step 2: Initialize Hugo site
	if err := hugo.InitializeSite(sitePath); err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to initialize Hugo site: %v", err)}
	}

	// Step 3: Create walgo.yaml
	if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to create walgo.yaml: %v", err)}
	}

	// Step 4: Import vault
	cfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("error loading config: %v", err)}
	}

	hugoContentDir := filepath.Join(sitePath, cfg.HugoConfig.ContentDir)
	if params.OutputDir != "" {
		hugoContentDir = filepath.Join(hugoContentDir, params.OutputDir)
	}

	// Build obsidian config
	obsidianCfg := cfg.ObsidianConfig
	if params.ConvertLinks {
		obsidianCfg.ConvertWikilinks = true
	}
	if params.LinkStyle != "" {
		obsidianCfg.LinkStyle = params.LinkStyle
	}
	if params.IncludeDrafts {
		obsidianCfg.IncludeDrafts = true
	}

	// Import vault content
	stats, err := obsidian.ImportVault(params.VaultPath, hugoContentDir, obsidianCfg)
	if err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("import failed: %v", err)}
	}

	if err := BuildSite(sitePath); err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	// Step 5: Create draft project using ORIGINAL name
	manager, err := projects.NewManager()
	if err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to create project manager: %v", err)}
	}
	defer manager.Close()

	if err := manager.CreateDraftProject(originalSiteName, sitePath); err != nil {
		return ImportObsidianResult{Error: fmt.Sprintf("failed to create draft project: %v", err)}
	}

	return ImportObsidianResult{
		Success:       true,
		FilesImported: stats.FilesProcessed,
		SitePath:      sitePath,
	}
}

// =============================================================================
// AI Configuration
// =============================================================================

// AIConfigureParams holds parameters for AI configuration
type AIConfigureParams struct {
	Provider string `json:"provider"` // "openai" or "openrouter"
	APIKey   string `json:"apiKey"`
	BaseURL  string `json:"baseURL,omitempty"`
	Model    string `json:"model,omitempty"`
}

// AIConfigResult holds the result of AI configuration
type AIConfigResult struct {
	Configured          bool     `json:"configured"`
	Enabled             bool     `json:"enabled"`
	Provider            string   `json:"provider,omitempty"`
	CurrentProvider     string   `json:"currentProvider,omitempty"`
	Model               string   `json:"model,omitempty"`
	CurrentModel        string   `json:"currentModel,omitempty"`
	ConfiguredProviders []string `json:"configuredProviders,omitempty"`
	Success             bool     `json:"success"`
	Error               string   `json:"error,omitempty"`
}

// GetAIConfig returns the current AI configuration
func GetAIConfig() (*AIConfigResult, error) {
	// Get list of all configured providers
	configuredProviders, err := ai.ListProviders()
	if err != nil {
		return &AIConfigResult{
			Configured: false,
			Enabled:    false,
			Success:    false,
			Error:      err.Error(),
		}, nil
	}

	// If no providers configured, return unconfigured state
	if len(configuredProviders) == 0 {
		return &AIConfigResult{
			Configured:          false,
			Enabled:             false,
			ConfiguredProviders: []string{},
			Success:             true,
		}, nil
	}

	// Get credentials for first provider as current provider
	currentProvider := configuredProviders[0]
	creds, err := ai.GetProviderCredentials(currentProvider)
	if err != nil {
		return &AIConfigResult{
			Configured:          true,
			Enabled:             true,
			ConfiguredProviders: configuredProviders,
			Success:             true,
		}, nil
	}

	return &AIConfigResult{
		Configured:          true,
		Enabled:             true,
		CurrentProvider:     currentProvider,
		Provider:            currentProvider,
		CurrentModel:        creds.Model,
		ConfiguredProviders: configuredProviders,
		Success:             true,
	}, nil
}

// UpdateAIConfig updates AI configuration and removes other providers
func UpdateAIConfig(params AIConfigureParams) error {
	// When saving new provider credentials, remove all other providers first
	if params.Provider != "" {
		_ = ai.RemoveProviderCredentials(params.Provider)
	}
	return ai.SetProviderCredentials(params.Provider, params.APIKey, params.BaseURL, params.Model)
}

// CleanAIConfig removes all AI credentials
func CleanAIConfig() error {
	return ai.RemoveAllCredentials()
}

// CleanProviderConfig removes credentials for a specific provider
func CleanProviderConfig(provider string) error {
	return ai.RemoveProviderCredentials(provider)
}

// GetProviderCredentialsAPI returns credentials for a specific provider
func GetProviderCredentialsAPI(provider string) (*ProviderCredentialsResult, error) {
	creds, err := ai.GetProviderCredentials(provider)
	if err != nil {
		return &ProviderCredentialsResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &ProviderCredentialsResult{
		Success: true,
		APIKey:  creds.APIKey,
		BaseURL: creds.BaseURL,
		Model:   creds.Model,
	}, nil
}

// ProviderCredentialsResult holds provider credentials
type ProviderCredentialsResult struct {
	Success bool   `json:"success"`
	APIKey  string `json:"apiKey"`
	BaseURL string `json:"baseURL"`
	Model   string `json:"model"`
	Error   string `json:"error,omitempty"`
}

// =============================================================================
// AI Content Generation
// =============================================================================

// GenerateContentParams holds content generation parameters
type GenerateContentParams struct {
	SitePath     string `json:"sitePath"`
	FilePath     string `json:"filePath,omitempty"` // Optional file path
	ContentType  string `json:"contentType"`        // "post" or "page"
	Topic        string `json:"topic"`
	Context      string `json:"context"`
	Instructions string `json:"instructions"` // New: simplified instructions
}

// GenerateContentResult holds the result of content generation
type GenerateContentResult struct {
	Success  bool   `json:"success"`
	Content  string `json:"content"`
	FilePath string `json:"filePath"`
	Error    string `json:"error"`
}

// ContentStructure holds information about the content directory structure
type ContentStructure = ai.ContentStructure

// ContentTypeInfo holds information about a content type
type ContentTypeInfo = ai.ContentTypeInfo

// GetContentStructure returns the content directory structure
func GetContentStructure(sitePath string) ContentStructure {
	structure, err := ai.GetContentStructure(sitePath)
	if err != nil {
		// Return empty structure on error
		return ai.ContentStructure{
			ContentTypes: []ai.ContentTypeInfo{},
			DefaultType:  "posts",
		}
	}
	return *structure
}

// GenerateContent creates new content using AI
func GenerateContent(params GenerateContentParams) GenerateContentResult {
	client, _, _, err := ai.LoadClient(0)
	if err != nil {
		return GenerateContentResult{Error: fmt.Sprintf("failed to load AI client: %v", err)}
	}

	if params.Instructions != "" {
		generator := ai.NewContentGenerator(client)

		result := generator.GenerateContent(ai.ContentGenerationParams{
			SitePath:     params.SitePath,
			Instructions: params.Instructions,
			Context:      context.Background(),
		})

		if !result.Success {
			return GenerateContentResult{
				Success: false,
				Error:   result.ErrorMessage,
			}
		}

		// Apply content fixer to ensure YAML frontmatter is correct
		siteType := hugo.DetermineSiteTypeFromPath(params.SitePath)
		fixer := ai.NewContentFixer(params.SitePath, ai.SiteType(siteType))
		if err := fixer.FixAll(); err != nil {
			// Warning only, don't fail the generation
			fmt.Fprintf(os.Stderr, "Warning: Content fix failed: %v\n", err)
		}

		if err := BuildSite(params.SitePath); err != nil {
			return GenerateContentResult{Error: fmt.Sprintf("failed to build site: %v", err)}
		}

		return GenerateContentResult{
			Success:  true,
			Content:  result.Content,
			FilePath: result.FilePath,
		}
	}

	isBlogContent := strings.Contains(strings.ToLower(params.ContentType), "post") ||
		strings.Contains(strings.ToLower(params.ContentType), "blog") ||
		strings.Contains(strings.ToLower(params.ContentType), "article") ||
		strings.Contains(strings.ToLower(params.ContentType), "news")

	var systemPrompt string
	if isBlogContent {
		systemPrompt = ai.SystemPromptBlogPost
	} else {
		systemPrompt = ai.SystemPromptPageGeneration
	}

	userPrompt := ai.BuildUserPrompt(params.Topic, params.Context)

	content, err := client.GenerateContent(systemPrompt, userPrompt)
	if err != nil {
		return GenerateContentResult{Error: fmt.Sprintf("generating content: %v", err)}
	}

	// Apply content fixer to ensure YAML frontmatter is correct
	fixer := ai.NewContentFixer(params.SitePath, ai.SiteType(hugo.DetermineSiteTypeFromPath(params.SitePath)))
	if err := fixer.FixAll(); err != nil {
		return GenerateContentResult{Error: fmt.Sprintf("failed to fix YAML frontmatter: %v", err)}
	}

	if err := BuildSite(params.SitePath); err != nil {
		return GenerateContentResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	return GenerateContentResult{
		Success: true,
		Content: ai.CleanGeneratedContent(content),
	}
}

// UpdateContentParams holds content update parameters
type UpdateContentParams struct {
	FilePath     string `json:"filePath"`
	Instructions string `json:"instructions"`
	SitePath     string `json:"sitePath"`
}

// UpdateContentResult holds the result of content update
type UpdateContentResult struct {
	Success        bool   `json:"success"`
	UpdatedContent string `json:"updatedContent"`
	Error          string `json:"error"`
}

// UpdateContent updates existing content using AI
func UpdateContent(params UpdateContentParams) UpdateContentResult {
	client, _, _, err := ai.LoadClient(0)
	if err != nil {
		return UpdateContentResult{Error: fmt.Sprintf("failed to load AI client: %v", err)}
	}

	existingContent, err := os.ReadFile(params.FilePath)
	if err != nil {
		return UpdateContentResult{Error: fmt.Sprintf("reading file: %v", err)}
	}

	userPrompt := ai.BuildUpdatePrompt(params.Instructions, string(existingContent))

	updatedContent, err := client.GenerateContent(ai.SystemPromptContentUpdate, userPrompt)
	if err != nil {
		return UpdateContentResult{Error: fmt.Sprintf("updating content: %v", err)}
	}

	updatedContent = ai.CleanGeneratedContent(updatedContent)

	if err := os.WriteFile(params.FilePath, []byte(updatedContent), 0644); err != nil {
		return UpdateContentResult{Error: fmt.Sprintf("saving file: %v", err)}
	}

	// Apply content fixer to ensure YAML frontmatter is correct
	if params.SitePath != "" {
		siteType := hugo.DetermineSiteTypeFromPath(params.SitePath)
		fixer := ai.NewContentFixer(params.SitePath, ai.SiteType(siteType))
		if err := fixer.FixAll(); err != nil {
			return UpdateContentResult{Error: fmt.Sprintf("failed to fix YAML frontmatter: %v", err)}
		}
	}

	if err := BuildSite(params.SitePath); err != nil {
		return UpdateContentResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	return UpdateContentResult{
		Success:        true,
		UpdatedContent: updatedContent,
	}
}

// =============================================================================
// Gas Estimation
// =============================================================================

// GasEstimateParams holds parameters for gas estimation
type GasEstimateParams struct {
	SitePath string `json:"sitePath"`
	Network  string `json:"network"`
	Epochs   int    `json:"epochs"`
}

// GasEstimateResult holds the gas estimation result
type GasEstimateResult struct {
	Success   bool    `json:"success"`
	WAL       float64 `json:"wal"`
	SUI       float64 `json:"sui"`
	WALRange  string  `json:"walRange"`
	SUIRange  string  `json:"suiRange"`
	Summary   string  `json:"summary"`
	SiteSize  int64   `json:"siteSize"`
	FileCount int     `json:"fileCount"`
	Error     string  `json:"error,omitempty"`
}

// EstimateGasFee estimates the gas fee for deploying a site
func EstimateGasFee(params GasEstimateParams) GasEstimateResult {
	result := GasEstimateResult{
		Success: false,
	}

	// Load config to get publish directory
	walgoCfg, err := config.LoadConfigFrom(params.SitePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load config: %v", err)
		return result
	}

	publishDir := filepath.Join(params.SitePath, walgoCfg.HugoConfig.PublishDir)

	// Check if publish directory exists
	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		result.Error = "publish directory not found. Please build the site first."
		return result
	}

	// Calculate site size and file count
	var siteSize int64
	var fileCount int

	if err := filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			siteSize += info.Size()
			fileCount++
		}
		return nil
	}); err != nil {
		result.Error = fmt.Sprintf("failed to calculate site size: %v", err)
		return result
	}

	// Get detailed cost estimate
	costEstimate, err := projects.EstimateGasFeeDetailed(params.Network, siteSize, params.Epochs, fileCount)
	if err != nil {
		result.Error = fmt.Sprintf("failed to estimate gas: %v", err)
		return result
	}

	result.Success = true
	result.WAL = costEstimate.WAL
	result.SUI = costEstimate.SUI
	result.WALRange = costEstimate.WALRange
	result.SUIRange = costEstimate.SUICostRange
	result.Summary = costEstimate.Summary
	result.SiteSize = siteSize
	result.FileCount = fileCount

	return result
}

// =============================================================================
// Launch Wizard
// =============================================================================

// LaunchStep represents a step in the launch wizard
type LaunchStep struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// LaunchWizardParams holds launch wizard parameters
type LaunchWizardParams struct {
	SitePath    string `json:"sitePath,omitempty"`
	Network     string `json:"network"`
	ProjectName string `json:"projectName"`
	Category    string `json:"category"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Epochs      int    `json:"epochs"`
	SkipConfirm bool   `json:"skipConfirm"`
}

// LaunchWizardResult holds launch wizard result
type LaunchWizardResult struct {
	Success  bool         `json:"success"`
	ObjectID string       `json:"objectId"`
	Steps    []LaunchStep `json:"steps"`
	Error    string       `json:"error"`
}

// LaunchWizard executes full launch wizard flow
func LaunchWizard(params LaunchWizardParams) LaunchWizardResult {
	result := LaunchWizardResult{
		Steps: []LaunchStep{},
	}

	sitePath := params.SitePath
	if sitePath == "" {
		result.Error = "site path is required"
		return result
	}

	walgoCfg, err := config.LoadConfigFrom(sitePath)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load config: %v", err)
		return result
	}

	if err := BuildSite(sitePath); err != nil {
		result.Error = fmt.Sprintf("failed to build site: %v", err)
		return result
	}

	publishDir := filepath.Join(sitePath, walgoCfg.HugoConfig.PublishDir)
	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		result.Error = "publish directory not found. Please build first."
		return result
	}

	// Prepare deployment options
	opts := deployment.DeploymentOptions{
		SitePath:    sitePath,
		PublishDir:  publishDir,
		Epochs:      params.Epochs,
		WalgoCfg:    walgoCfg,
		Quiet:       false,
		Verbose:     true,
		ForceNew:    false,
		DryRun:      false,
		SaveProject: true,
		ProjectName: params.ProjectName,
		Category:    params.Category,
		Network:     params.Network,
		Description: params.Description,
		ImageURL:    params.ImageURL,
	}

	// Perform deployment
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	deployResult, err := deployment.PerformDeployment(ctx, opts)
	if err != nil {
		result.Error = fmt.Sprintf("deployment failed: %v", err)
		return result
	}

	if !deployResult.Success {
		result.Error = "deployment failed"
		return result
	}

	result.Success = true
	result.ObjectID = deployResult.ObjectID

	return result
}

// =============================================================================
// AI Create Site (Pipeline)
// =============================================================================

// AICreateSiteParams holds AI site creation parameters
type AICreateSiteParams struct {
	ParentDir   string `json:"parentDir,omitempty"`
	SiteName    string `json:"siteName"`
	SiteType    string `json:"siteType"` // "blog", "portfolio", "docs", "business"
	Description string `json:"description,omitempty"`
	Audience    string `json:"audience,omitempty"`
}

// AICreateSiteResult holds AI site creation result
type AICreateSiteResult struct {
	Success      bool         `json:"success"`
	SitePath     string       `json:"sitePath"`
	TotalPages   int          `json:"totalPages"`
	FilesCreated int          `json:"filesCreated"`
	Steps        []LaunchStep `json:"steps"`
	Error        string       `json:"error"`
}

// AICreateSiteWithProgress creates a site with a custom progress handler (for desktop app)
func AICreateSiteWithProgress(params AICreateSiteParams, progressHandler ProgressHandler) AICreateSiteResult {
	result := AICreateSiteResult{}

	// Keep original site name for database and Hugo config
	originalSiteName := params.SiteName

	// Sanitize site name ONLY for directory creation
	sanitizedDirName := utils.SanitizeSiteName(params.SiteName)
	if sanitizedDirName != originalSiteName {
		fmt.Printf("Directory name sanitized: '%s' -> '%s'\n", originalSiteName, sanitizedDirName)
	}

	// Use default walgo-sites directory if parentDir is empty
	parentDir := params.ParentDir
	if parentDir == "" {
		defaultDir, err := GetDefaultSitesDir()
		if err != nil {
			result.Error = fmt.Sprintf("cannot determine walgo-sites directory: %v", err)
			return result
		}
		parentDir = defaultDir
	}

	// Create site directory using sanitized name
	sitePath, err := filepath.Abs(filepath.Join(parentDir, sanitizedDirName))
	if err != nil {
		result.Error = fmt.Sprintf("invalid site path: %v", err)
		return result
	}

	// Create the site directory if it doesn't exist
	if err := os.MkdirAll(sitePath, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create site directory: %v", err)
		return result
	}

	client, _, _, err := ai.LoadClient(0)
	if err != nil {
		result.Error = fmt.Sprintf("failed to load AI client: %v", err)
		return result
	}

	// Map site type string to ai.SiteType
	var siteType ai.SiteType
	switch params.SiteType {
	case "blog":
		siteType = ai.SiteTypeBlog
	case "portfolio":
		siteType = ai.SiteTypePortfolio
	case "docs":
		siteType = ai.SiteTypeDocs
	case "business":
		siteType = ai.SiteTypeBusiness
	default:
		siteType = ai.SiteTypeBlog
	}

	walgoConfigPath := filepath.Join(sitePath, config.DefaultConfigFileName)
	if _, err := os.Stat(walgoConfigPath); os.IsNotExist(err) {
		fmt.Printf("\nwalgo.yaml not found, creating default configuration...\n")
		if err := config.CreateDefaultWalgoConfig(sitePath); err != nil {
			result.Error = fmt.Sprintf("failed to create walgo.yaml: %v", err)
			return result
		}
		fmt.Printf("   Created walgo.yaml configuration\n")
	}

	// Setup Hugo site - use ORIGINAL name
	hugoSiteType := hugo.SiteType(siteType)
	if err := hugo.SetupSiteConfigWithName(sitePath, hugoSiteType, originalSiteName); err != nil {
		result.Error = fmt.Sprintf("failed to set up config: %v", err)
		return result
	}

	if err := hugo.SetupArchetypes(sitePath); err != nil {
		// Warning but not fatal
	}

	if err := hugo.SetupFaviconForTheme(sitePath, hugoSiteType); err != nil {
		// Warning but not fatal
	}

	themeInfo := hugo.GetThemeInfo(hugoSiteType)
	if err := hugo.InstallTheme(sitePath, hugoSiteType); err != nil {
		// Warning but not fatal
	}

	if hugoSiteType == hugo.SiteTypeBusiness {
		if err := hugo.SetupBusinessThemeOverrides(sitePath); err != nil {
			return AICreateSiteResult{Error: fmt.Sprintf("failed to set up theme overrides: %v", err)}
		}
	}
	if hugoSiteType == hugo.SiteTypePortfolio {
		if err := hugo.SetupPortfolioThemeOverrides(sitePath); err != nil {
			return AICreateSiteResult{Error: fmt.Sprintf("failed to set up theme overrides: %v", err)}
		}
	}
	if hugoSiteType == hugo.SiteTypeDocs {
		if err := hugo.SetupDocsThemeOverrides(sitePath); err != nil {
			return AICreateSiteResult{Error: fmt.Sprintf("failed to set up theme overrides: %v", err)}
		}
	}

	// Run AI pipeline with absolute paths to ensure content is created in the correct location
	pipelineConfig := ai.DefaultPipelineConfig()
	pipelineConfig.ContentDir = filepath.Join(sitePath, "content")
	pipelineConfig.PlanPath = filepath.Join(sitePath, ".walgo", "plan.json")
	pipeline := ai.NewPipeline(client, pipelineConfig)

	// Use custom progress handler if provided, otherwise use console handler
	if progressHandler != nil {
		// Convert our public ProgressHandler to internal ai.ProgressHandler
		internalHandler := func(event ai.ProgressEvent) {
			// Convert internal event to public event
			publicEvent := ProgressEvent{
				Phase:     string(event.Phase),
				EventType: string(event.EventType),
				Message:   event.Message,
				PagePath:  event.PagePath,
				Progress:  event.Progress,
				Current:   event.Current,
				Total:     event.Total,
			}
			progressHandler(publicEvent)
		}
		pipeline.SetProgressHandler(internalHandler)
	} else {
		pipeline.SetProgressHandler(ai.ConsoleProgressHandler(false))
	}

	input := &ai.PlannerInput{
		SiteName:    originalSiteName, // Use original name for AI content
		SiteType:    siteType,
		Description: params.Description,
		Audience:    params.Audience,
		Theme:       themeInfo.Name,
	}

	ctx := context.Background()
	pipelineResult, err := pipeline.Run(ctx, input)
	if err != nil {
		result.Error = fmt.Sprintf("pipeline error: %v", err)
		return result
	}

	if pipelineResult.Plan != nil {
		hugoTomlPath := filepath.Join(sitePath, "hugo.toml")
		if _, err := os.Stat(hugoTomlPath); os.IsNotExist(err) {
			hugoTomlPath = filepath.Join(sitePath, "config.toml")
		}
		if _, err := os.Stat(hugoTomlPath); err == nil {
			fmt.Printf("Updating Hugo menu configuration...\n")
			if err := hugo.ApplyMenuFromSitePlan(pipelineResult.Plan, hugoTomlPath); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to apply menu: %v", err)}
			}
		}
		switch siteType {
		case ai.SiteTypeBusiness:
			fmt.Printf("\n%s Validating and fixing content for Ananke theme...\n", ui.GetIcons().Gear)
			fixer := ai.NewContentFixer(sitePath, siteType)
			if err := fixer.FixAll(); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to fix content: %v", err)}
			} else {
				fmt.Printf("Content validated and fixed\n")
			}

			issues := ai.ValidateBusinessContent(sitePath)
			if len(issues) > 0 {
				fmt.Printf("Remaining issues:\n")
				for _, issue := range issues {
					fmt.Printf("      - %s\n", issue)
				}
			}

		case ai.SiteTypeBlog:
			fmt.Printf("\nValidating and fixing content for Ananke theme...\n")
			fixer := ai.NewContentFixer(sitePath, siteType)
			if err := fixer.FixAll(); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to fix content: %v", err)}
			} else {
				fmt.Printf("Content validated and fixed\n")
			}

			issues := ai.ValidateBlogContent(sitePath)
			if len(issues) > 0 {
				fmt.Printf("Remaining issues:\n")
				for _, issue := range issues {
					fmt.Printf("      - %s\n", issue)
				}
			}

		case ai.SiteTypePortfolio:
			if err := hugo.UpdatePortfolioParams(sitePath, pipelineResult.Plan.Description, pipelineResult.Plan.Audience); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to update portfolio params: %v", err)}
			} else {
				fmt.Printf("Updated portfolio params\n")
			}

			fmt.Printf("\nValidating and fixing content for Ananke theme...\n")
			fixer := ai.NewContentFixer(sitePath, siteType)
			if err := fixer.FixAll(); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to fix content: %v", err)}
			} else {
				fmt.Printf("Content validated and fixed\n")
			}

			issues := ai.ValidatePortfolioContent(sitePath)
			if len(issues) > 0 {
				fmt.Printf("Remaining issues:\n")
				for _, issue := range issues {
					fmt.Printf("      - %s\n", issue)
				}
			}

		case ai.SiteTypeDocs:
			if err := hugo.UpdateDocsParams(sitePath, pipelineResult.Plan.Description); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to update docs params: %v", err)}
			}

			fmt.Printf("\nValidating and fixing content for Hugo Book theme...\n")
			fixer := ai.NewContentFixer(sitePath, siteType)
			if err := fixer.FixAll(); err != nil {
				return AICreateSiteResult{Error: fmt.Sprintf("failed to fix content: %v", err)}
			} else {
				fmt.Printf("Content validated and fixed\n")
			}

			issues := ai.ValidateDocsContent(sitePath)
			if len(issues) > 0 {
				fmt.Printf("Remaining issues:\n")
				for _, issue := range issues {
					fmt.Printf("      - %s\n", issue)
				}
			}
		}
	}

	if err := BuildSite(sitePath); err != nil {
		return AICreateSiteResult{Error: fmt.Sprintf("failed to build site: %v", err)}
	}

	result.Success = true
	result.SitePath = sitePath
	result.TotalPages = pipelineResult.Plan.Stats.TotalPages
	result.FilesCreated = pipelineResult.Plan.Stats.TotalPages
	result.Steps = []LaunchStep{
		{Name: "plan", Status: "completed", Message: "Content plan created"},
		{Name: "generate", Status: "completed", Message: fmt.Sprintf("%d pages generated", pipelineResult.Plan.Stats.TotalPages)},
	}

	// Save as draft project using ORIGINAL name
	if err := saveDraftProject(originalSiteName, sitePath); err != nil {
		result.Error = fmt.Sprintf("failed to save draft project: %v", err)
		return result
	}

	return result
}

// AICreateSite creates a complete Hugo site with AI-generated content (CLI version with console output)
func AICreateSite(params AICreateSiteParams) AICreateSiteResult {
	return AICreateSiteWithProgress(params, nil)
}

// =============================================================================
// Setup Dependencies
// =============================================================================

// SetupDepsResult holds setup dependencies result
type SetupDepsResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// CheckSetupDeps checks if all required dependencies are installed
func CheckSetupDeps() SetupDepsResult {
	result := SetupDepsResult{
		Success: true,
	}

	missingTools := deps.GetMissingTools()
	if len(missingTools) > 0 {
		result.Success = false
		result.Message = fmt.Sprintf("missing required tools: %s", strings.Join(missingTools, ", "))
		result.Error = deps.InstallInstructions("testnet")
		return result
	}

	result.Message = "All required dependencies are installed"
	return result
}

// GetSystemHealth returns current system health status
func GetSystemHealth() SystemHealth {
	health := SystemHealth{
		NetOnline:       false,
		SuiInstalled:    false,
		SuiConfigured:   false,
		WalrusInstalled: false,
		SiteBuilder:     false,
		HugoInstalled:   false,
		Message:         "Checking...",
	}

	// Check network connectivity with timeout
	// Use multiple endpoints to avoid rate limiting and increase reliability
	health.NetOnline = checkNetworkConnectivity()

	// Check Sui CLI
	if _, err := deps.LookPath("sui"); err == nil {
		health.SuiInstalled = true
		// Check if Sui is configured (has active address)
		if _, err := sui.GetActiveAddress(); err == nil {
			health.SuiConfigured = true
		}
	}

	// Check Walrus CLI
	if _, err := deps.LookPath("walrus"); err == nil {
		health.WalrusInstalled = true
	}

	// Check site-builder
	if _, err := deps.LookPath("site-builder"); err == nil {
		health.SiteBuilder = true
	}

	// Check Hugo
	if _, err := deps.LookPath("hugo"); err == nil {
		health.HugoInstalled = true
	}

	// Set message based on status
	allReady := health.NetOnline && health.SuiInstalled && health.SuiConfigured &&
		health.WalrusInstalled && health.SiteBuilder && health.HugoInstalled

	if allReady {
		health.Message = "Ready to deploy"
	} else if !health.NetOnline {
		health.Message = "Network offline"
	} else if !health.SuiInstalled {
		health.Message = "Sui not installed"
	} else if !health.SuiConfigured {
		health.Message = "Sui not configured"
	} else if !health.WalrusInstalled {
		health.Message = "Walrus not installed"
	} else if !health.SiteBuilder {
		health.Message = "site-builder not installed"
	} else if !health.HugoInstalled {
		health.Message = "Hugo not installed"
	}

	return health
}

// ToolVersionInfo represents version information for a tool
type ToolVersionInfo struct {
	Tool           string `json:"tool"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	UpdateRequired bool   `json:"updateRequired"`
	Installed      bool   `json:"installed"`
}

// CheckToolVersionsResult holds the result of version checking
type CheckToolVersionsResult struct {
	Success bool              `json:"success"`
	Tools   []ToolVersionInfo `json:"tools"`
	Message string            `json:"message"`
	Error   string            `json:"error,omitempty"`
}

// CheckToolVersions checks if installed tools have updates available
func CheckToolVersions() CheckToolVersionsResult {
	result := CheckToolVersionsResult{
		Success: true,
		Tools:   []ToolVersionInfo{},
		Message: "Checking versions...",
	}

	// Check all versions using internal/version package
	versionResult, err := version.CheckAllVersions()
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "Failed to check versions"
		return result
	}

	// Convert to API format
	if versionResult.Sui != nil {
		result.Tools = append(result.Tools, ToolVersionInfo{
			Tool:           versionResult.Sui.Tool,
			CurrentVersion: versionResult.Sui.CurrentVersion,
			LatestVersion:  versionResult.Sui.LatestVersion,
			UpdateRequired: versionResult.Sui.UpdateRequired,
			Installed:      true,
		})
	}

	if versionResult.Walrus != nil {
		result.Tools = append(result.Tools, ToolVersionInfo{
			Tool:           versionResult.Walrus.Tool,
			CurrentVersion: versionResult.Walrus.CurrentVersion,
			LatestVersion:  versionResult.Walrus.LatestVersion,
			UpdateRequired: versionResult.Walrus.UpdateRequired,
			Installed:      true,
		})
	}

	if versionResult.SiteBuilder != nil {
		result.Tools = append(result.Tools, ToolVersionInfo{
			Tool:           versionResult.SiteBuilder.Tool,
			CurrentVersion: versionResult.SiteBuilder.CurrentVersion,
			LatestVersion:  versionResult.SiteBuilder.LatestVersion,
			UpdateRequired: versionResult.SiteBuilder.UpdateRequired,
			Installed:      true,
		})
	}

	result.Message = "Version check complete"
	return result
}

// UpdateToolsParams holds parameters for updating tools
type UpdateToolsParams struct {
	Tools   []string `json:"tools"`   // List of tools to update (e.g., ["sui", "walrus"])
	Network string   `json:"network"` // Network for suiup tools (testnet/mainnet)
}

// UpdateToolsResult holds the result of updating tools
type UpdateToolsResult struct {
	Success      bool              `json:"success"`
	UpdatedTools []string          `json:"updatedTools"`
	FailedTools  map[string]string `json:"failedTools"` // tool -> error message
	Message      string            `json:"message"`
}

// UpdateTools updates specified tools to their latest versions
func UpdateTools(params UpdateToolsParams) UpdateToolsResult {
	result := UpdateToolsResult{
		Success:      true,
		UpdatedTools: []string{},
		FailedTools:  make(map[string]string),
		Message:      "Updating tools...",
	}

	if len(params.Tools) == 0 {
		result.Success = false
		result.Message = "No tools specified"
		return result
	}

	// Default to testnet if not specified
	network := params.Network
	if network == "" {
		network = "testnet"
	}

	// Update each tool
	for _, tool := range params.Tools {
		err := version.UpdateTool(tool, network)
		if err != nil {
			result.FailedTools[tool] = err.Error()
			result.Success = false
		} else {
			result.UpdatedTools = append(result.UpdatedTools, tool)
		}
	}

	// Set final message
	if result.Success {
		result.Message = fmt.Sprintf("Successfully updated: %s", strings.Join(result.UpdatedTools, ", "))
	} else {
		if len(result.UpdatedTools) > 0 {
			result.Message = fmt.Sprintf("Partially updated. Success: %s, Failed: %d",
				strings.Join(result.UpdatedTools, ", "), len(result.FailedTools))
		} else {
			result.Message = "All updates failed"
		}
	}

	return result
}
