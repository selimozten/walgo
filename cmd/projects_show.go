package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/ui"
	"github.com/ganbitlabs/walgo/internal/walrus"
)

// showProjectDetails displays detailed information about a specific project.
func showProjectDetails(proj *projects.Project) error {
	icons := ui.GetIcons()
	pm, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

	stats, err := pm.GetProjectStats(proj.ID)
	if err != nil {
		return fmt.Errorf("failed to get project stats: %w", err)
	}

	deployments, err := pm.GetProjectDeployments(proj.ID)
	if err != nil {
		return fmt.Errorf("failed to get deployments: %w", err)
	}

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Printf("║  %s %s%s║\n", icons.Package, proj.Name, strings.Repeat(" ", 55-len(proj.Name)))
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("%s Project Information\n", icons.Info)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  ID:              %d\n", proj.ID)
	fmt.Printf("  Name:            %s\n", proj.Name)
	if proj.Category != "" {
		fmt.Printf("  Category:        %s\n", proj.Category)
	}
	if proj.Description != "" {
		desc := proj.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("  Description:     %s\n", desc)
	}
	if proj.ImageURL != "" {
		url := proj.ImageURL
		if len(url) > 60 {
			url = url[:57] + "..."
		}
		fmt.Printf("  Image URL:       %s\n", url)
	}
	fmt.Printf("  Network:         %s\n", proj.Network)
	fmt.Printf("  Status:          %s\n", proj.Status)
	fmt.Printf("  Object ID:       %s\n", proj.ObjectID)
	if proj.SuiNS != "" {
		fmt.Printf("  SuiNS:           %s\n", proj.SuiNS)
		fmt.Printf("  URL:             https://%s.wal.app\n", proj.SuiNS)
	}
	fmt.Printf("  Wallet:          %s\n", proj.WalletAddr)
	fmt.Printf("  Site path:       %s\n", proj.SitePath)
	fmt.Printf("  Created:         %s\n", proj.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("%s Storage & Deployment\n", icons.Info)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get epoch info from all deployments for accurate expiry calculation
	epochInfo, epochErr := pm.GetEpochInfo(proj.ID)
	_ = epochErr // Ignore error, will use fallback

	// Storage Epochs
	if epochInfo != nil && epochInfo.TotalEpochs > 0 {
		fmt.Printf("  Total Storage Epochs:  %s\n", formatEpochs(epochInfo.TotalEpochs, proj.Network))

		// Show last deployment epochs if different from total (indicates multiple deployments)
		if proj.Epochs > 0 && proj.Epochs != epochInfo.TotalEpochs {
			fmt.Printf("  Last Deploy Epochs:    %d\n", proj.Epochs)
		}

		// Calculate expiry from first deployment + total epochs
		if !epochInfo.FirstDeploymentAt.IsZero() {
			if expiryDate, ok := estimateExpiry(epochInfo.FirstDeploymentAt, epochInfo.TotalEpochs, proj.Network); ok {
				fmt.Printf("  Expires In (est.):     %s\n", formatExpiryDuration(expiryDate))
			}
		}
	} else if proj.Epochs > 0 {
		// Fallback to project epochs if epoch info not available
		fmt.Printf("  Storage Epochs:        %s\n", formatEpochs(proj.Epochs, proj.Network))

		if !proj.LastDeployAt.IsZero() {
			if expiryDate, ok := estimateExpiry(proj.LastDeployAt, proj.Epochs, proj.Network); ok {
				fmt.Printf("  Expires In (est.):     %s\n", formatExpiryDuration(expiryDate))
			}
		}
	}

	// The database only knows what walgo last did; the chain knows what storage
	// is actually paid for. Prefer the latter when the project is reachable from
	// the active Sui environment.
	if proj.ObjectID != "" {
		if _, err := resolveDeployNetwork(proj.Network); err == nil {
			if expiry, err := walrus.GetSiteExpiry(proj.ObjectID); err == nil {
				fmt.Printf("  On-chain Storage:      %s\n", expiry.Describe(time.Now()))
				if expiry.HasExpired() {
					fmt.Printf("  %s Expired storage cannot be extended; 'walgo update --epochs <n>'\n", icons.Warning)
					fmt.Printf("     re-uploads the content from the local build.\n")
				}
			}
		}
	}

	// Gas Fee (actual cost from last deployment)
	if proj.GasFee != "" {
		fmt.Printf("  Gas Fee:               %s\n", proj.GasFee)
	}

	// Last Deploy
	if !proj.LastDeployAt.IsZero() {
		fmt.Printf("  Last Deploy:           %s\n", proj.LastDeployAt.Format("2006-01-02 15:04"))
	}

	fmt.Printf("  Total Deployments:     %d\n", stats.TotalDeployments)
	fmt.Printf("  Successful:            %d\n", stats.SuccessfulDeploys)
	if stats.FailedDeploys > 0 {
		fmt.Printf("  Failed:                %d\n", stats.FailedDeploys)
	}
	if !stats.FirstDeployment.IsZero() {
		fmt.Printf("  First Deployment:      %s\n", stats.FirstDeployment.Format("2006-01-02 15:04"))
	}
	fmt.Println()

	if len(deployments) > 0 {
		fmt.Printf("%s Recent Deployments\n", icons.Info)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		maxShow := 5
		if len(deployments) < maxShow {
			maxShow = len(deployments)
		}

		for i := 0; i < maxShow; i++ {
			d := deployments[i]
			status := icons.Check
			if !d.Success {
				status = icons.Cross
			}

			fmt.Printf("  %s %s - %d epochs", status, d.CreatedAt.Format("2006-01-02 15:04"), d.Epochs)
			if d.GasFee != "" {
				fmt.Printf(" - %s", d.GasFee)
			}
			if d.Version != "" {
				fmt.Printf(" (v%s)", d.Version)
			}
			fmt.Println()

			if d.Notes != "" {
				fmt.Printf("     Notes: %s\n", d.Notes)
			}
			if !d.Success && d.Error != "" {
				fmt.Printf("     Error: %s\n", d.Error)
			}
		}

		if len(deployments) > maxShow {
			fmt.Printf("\n  ... and %d more deployments\n", len(deployments)-maxShow)
		}
		fmt.Println()
	}

	fmt.Printf("%s Available Actions:\n", icons.Lightbulb)
	fmt.Printf("   - Update site:   walgo projects update %s\n", proj.Name)
	fmt.Printf("   - Edit metadata: walgo projects edit %s --name 'New Name'\n", proj.Name)
	fmt.Printf("   - Archive:       walgo projects archive %s\n", proj.Name)
	fmt.Printf("   - Delete:        walgo projects delete %s\n", proj.Name)
	fmt.Println()

	return nil
}

// calculateExpiryDate calculates when the storage will expire based on last deployment
func calculateExpiryDate(lastDeploy time.Time, epochs int, network string) time.Time {
	// Mainnet: ~2 weeks per epoch, Testnet: ~1 day per epoch
	var daysPerEpoch int
	if network == "mainnet" {
		daysPerEpoch = 14
	} else {
		daysPerEpoch = 1
	}

	totalDays := epochs * daysPerEpoch
	return lastDeploy.Add(time.Duration(totalDays) * 24 * time.Hour)
}

// estimateExpiry is calculateExpiryDate with an honesty check: epoch length
// depends on the network, so without a recorded network the estimate would be
// off by 14x and is better left unsaid.
func estimateExpiry(lastDeploy time.Time, epochs int, network string) (time.Time, bool) {
	switch normalizeNetwork(network) {
	case "mainnet", "testnet":
		return calculateExpiryDate(lastDeploy, epochs, normalizeNetwork(network)), true
	default:
		return time.Time{}, false
	}
}

// formatEpochs renders an epoch count, adding the wall-clock duration only when
// the network is known — a mainnet epoch lasts two weeks and a testnet one a day.
func formatEpochs(epochs int, network string) string {
	switch normalizeNetwork(network) {
	case "mainnet", "testnet":
		return fmt.Sprintf("%d (~%s)", epochs, projects.CalculateStorageDuration(epochs, normalizeNetwork(network)))
	default:
		return fmt.Sprintf("%d (network unknown)", epochs)
	}
}

// formatExpiryDuration formats the time until expiry in a human-readable format
func formatExpiryDuration(expiryDate time.Time) string {
	now := time.Now()
	diff := expiryDate.Sub(now)

	if diff < 0 {
		return "Expired"
	}

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24

	if days == 0 {
		if hours == 0 {
			return "Expiring soon"
		}
		return fmt.Sprintf("%d hours", hours)
	}

	if days == 1 {
		if hours > 0 {
			return fmt.Sprintf("1 day, %d hours", hours)
		}
		return "1 day"
	}

	if days >= 7 {
		weeks := days / 7
		remainingDays := days % 7
		if remainingDays > 0 {
			return fmt.Sprintf("%d weeks, %d days", weeks, remainingDays)
		}
		return fmt.Sprintf("%d weeks", weeks)
	}

	return fmt.Sprintf("%d days", days)
}
