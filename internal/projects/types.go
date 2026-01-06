package projects

import (
	"fmt"
	"time"

	"github.com/selimozten/walgo/internal/walrus"
)

// Project represents a deployed Walrus site project
type Project struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	Network      string    `json:"network"`     // testnet or mainnet
	ObjectID     string    `json:"object_id"`   // Site object ID from site-builder
	SuiNS        string    `json:"suins"`       // SuiNS domain (optional)
	WalletAddr   string    `json:"wallet_addr"` // Wallet address used
	Epochs       int       `json:"epochs"`      // Number of epochs
	GasFee       string    `json:"gas_fee"`     // Gas fee paid
	SitePath     string    `json:"site_path"`   // Local site path
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastDeployAt time.Time `json:"last_deploy_at"`
	DeployCount  int       `json:"deploy_count"` // Number of times deployed
	Status       string    `json:"status"`       // active, archived, etc.
	// Metadata for ws-resources.json (displayed on wallets/explorers)
	Description string `json:"description"` // Site description
	ImageURL    string `json:"image_url"`   // Site logo/image URL
}

// DeploymentRecord represents a single deployment of a project
type DeploymentRecord struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	ObjectID  string    `json:"object_id"`
	Network   string    `json:"network"`
	Epochs    int       `json:"epochs"`
	GasFee    string    `json:"gas_fee"`
	Version   string    `json:"version"` // Optional version tag
	Notes     string    `json:"notes"`   // Optional deployment notes
	Success   bool      `json:"success"`
	Error     string    `json:"error"` // Error message if failed
	CreatedAt time.Time `json:"created_at"`
}

// ProjectStats provides statistics about a project
type ProjectStats struct {
	TotalDeployments  int
	SuccessfulDeploys int
	FailedDeploys     int
	TotalGasSpent     string
	FirstDeployment   time.Time
	LastDeployment    time.Time
	CurrentNetwork    string
	CurrentObjectID   string
}

// NetworkConfig holds network-specific configuration
type NetworkConfig struct {
	Name          string
	EpochDuration string // "1 day" for testnet, "2 weeks" for mainnet
	MaxEpochs     int    // Maximum epochs allowed
}

// GetNetworkConfig returns configuration for a network
func GetNetworkConfig(network string) NetworkConfig {
	switch network {
	case "mainnet":
		return NetworkConfig{
			Name:          "mainnet",
			EpochDuration: "2 weeks",
			MaxEpochs:     53,
		}
	case "testnet":
		return NetworkConfig{
			Name:          "testnet",
			EpochDuration: "1 day",
			MaxEpochs:     53,
		}
	default:
		return NetworkConfig{
			Name:          "testnet",
			EpochDuration: "1 day",
			MaxEpochs:     53,
		}
	}
}

// CalculateStorageDuration calculates total storage duration based on epochs and network
func CalculateStorageDuration(epochs int, network string) string {
	if network == "mainnet" {
		weeks := epochs * 2
		if weeks >= 4 {
			months := weeks / 4
			remainingWeeks := weeks % 4
			if remainingWeeks > 0 {
				return fmt.Sprintf("%d months, %d weeks", months, remainingWeeks)
			}
			return fmt.Sprintf("%d months", months)
		}
		return fmt.Sprintf("%d weeks", weeks)
	}

	// testnet
	if epochs > 7 {
		weeks := epochs / 7
		remainingDays := epochs % 7
		if remainingDays > 0 {
			return fmt.Sprintf("%d weeks, %d days", weeks, remainingDays)
		}
		return fmt.Sprintf("%d weeks", weeks)
	}
	return fmt.Sprintf("%d days", epochs)
}

// CostEstimate represents a detailed cost estimate for operations
type CostEstimate struct {
	WAL          float64 // WAL tokens for storage
	SUI          float64 // SUI tokens for transaction gas
	WALRange     string  // WAL cost range (e.g., "0.01 - 0.02")
	SUICostRange string  // SUI cost range
	Summary      string  // Human-readable summary
}

// EstimateGasFee provides a professional estimate of gas fees for deployment
// Uses real Sui RPC gas price and Walrus storage pricing
func EstimateGasFee(network string, siteSize int64) string {
	return EstimateGasFeeWithEpochs(network, siteSize, 1)
}

// EstimateGasFeeWithEpochs provides a detailed estimate with epoch count
func EstimateGasFeeWithEpochs(network string, siteSize int64, epochs int) string {
	if epochs <= 0 {
		epochs = 1
	}

	breakdown, err := walrus.CalculateCost(walrus.CostOptions{
		SiteSize: siteSize,
		Epochs:   epochs,
		Network:  network,
	})
	if err != nil {
		// Fallback to simple estimation
		return walrus.EstimateCostSimple(siteSize, epochs, network)
	}

	// Format with both WAL and SUI costs
	return fmt.Sprintf("~%.4f WAL + ~%.4f SUI", breakdown.TotalWAL, breakdown.GasCostSUI)
}

// formatSmallValue formats a value, showing "< 0.0001" for very small non-zero values
func formatSmallValue(val float64) string {
	if val == 0 {
		return "0.0000"
	}
	if val > 0 && val < 0.0001 {
		return "< 0.0001"
	}
	return fmt.Sprintf("%.4f", val)
}

// formatSmallRange formats a range, handling small values properly
func formatSmallRange(min, max float64, unit string) string {
	minStr := formatSmallValue(min)
	maxStr := formatSmallValue(max)
	return fmt.Sprintf("%s - %s %s", minStr, maxStr, unit)
}

// EstimateGasFeeDetailed returns a detailed cost breakdown for display
func EstimateGasFeeDetailed(network string, siteSize int64, epochs int, fileCount int) (*CostEstimate, error) {
	if epochs <= 0 {
		epochs = 1
	}

	breakdown, err := walrus.CalculateCost(walrus.CostOptions{
		SiteSize:  siteSize,
		Epochs:    epochs,
		Network:   network,
		FileCount: fileCount,
	})
	if err != nil {
		return nil, err
	}

	return &CostEstimate{
		WAL:          breakdown.TotalWAL,
		SUI:          breakdown.GasCostSUI,
		WALRange:     formatSmallRange(breakdown.MinTotalWAL, breakdown.MaxTotalWAL, "WAL"),
		SUICostRange: formatSmallRange(breakdown.MinTotalSUI, breakdown.MaxTotalSUI, "SUI"),
		Summary: fmt.Sprintf("~%s WAL + ~%.4f SUI for %d files (%d epochs)",
			formatSmallValue(breakdown.TotalWAL), breakdown.GasCostSUI, breakdown.FileCount, epochs),
	}, nil
}

// EstimateUpdateCost estimates the cost for updating an existing site
func EstimateUpdateCost(network string, changedSize int64, newFiles int, epochs int) string {
	if changedSize <= 0 && newFiles <= 0 {
		// Only metadata update
		return "~0.001 SUI (metadata only)"
	}

	breakdown, err := walrus.CalculateUpdateCost(changedSize, newFiles, epochs, network)
	if err != nil {
		return "Unable to estimate"
	}

	if breakdown.TotalWAL == 0 {
		return fmt.Sprintf("~%.4f SUI", breakdown.GasCostSUI)
	}

	return fmt.Sprintf("~%.4f WAL + ~%.4f SUI", breakdown.TotalWAL, breakdown.GasCostSUI)
}

// EstimateDestroyCost estimates the cost for destroying a site
func EstimateDestroyCost(network string) string {
	breakdown, err := walrus.CalculateDestroyCost(network)
	if err != nil {
		return "~0.01 SUI"
	}

	return fmt.Sprintf("~%.4f SUI", breakdown.GasCostSUI)
}

// FormatCostBreakdownStr returns a formatted string for display
func FormatCostBreakdownStr(network string, siteSize int64, epochs int, fileCount int) string {
	breakdown, err := walrus.CalculateCost(walrus.CostOptions{
		SiteSize:  siteSize,
		Epochs:    epochs,
		Network:   network,
		FileCount: fileCount,
	})
	if err != nil {
		return fmt.Sprintf("Estimated cost: %s", EstimateGasFee(network, siteSize))
	}

	return walrus.FormatCostBreakdown(*breakdown)
}
