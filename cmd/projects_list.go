package cmd

import (
	"fmt"
	"time"

	"github.com/ganbitlabs/walgo/internal/projects"
	"github.com/ganbitlabs/walgo/internal/ui"
)

// listProjects displays all projects matching the given filters.
func listProjects(network, status string) error {
	icons := ui.GetIcons()
	pm, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

	projectList, err := pm.ListProjects(network, status)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projectList) == 0 {
		fmt.Println()
		fmt.Println("No projects found.")
		fmt.Println()
		fmt.Printf("%s Deploy your first site with: walgo launch\n", icons.Lightbulb)
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  Your Walrus Projects                     ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()

	for i, proj := range projectList {
		if i > 0 {
			fmt.Println("─────────────────────────────────────────────────────────────")
		}

		fmt.Printf("%s %s", icons.Package, proj.Name)
		if proj.Category != "" {
			fmt.Printf(" (%s)", proj.Category)
		}
		fmt.Println()

		fmt.Printf("   ID:           %d\n", proj.ID)
		fmt.Printf("   Network:      %s\n", proj.Network)
		fmt.Printf("   Status:       %s\n", proj.Status)
		fmt.Printf("   Object ID:    %s\n", proj.ObjectID)

		if proj.SuiNS != "" {
			fmt.Printf("   SuiNS:        %s\n", proj.SuiNS)
		}

		fmt.Printf("   Deployments:  %d\n", proj.DeployCount)
		fmt.Printf("   Last deploy:  %s\n", proj.LastDeployAt.Format("2006-01-02 15:04"))

		// Show epoch and expiry info
		epochInfo, err := pm.GetEpochInfo(proj.ID)
		if err == nil && epochInfo != nil && epochInfo.TotalEpochs > 0 {
			fmt.Printf("   Epochs:       %s\n", formatEpochs(epochInfo.TotalEpochs, proj.Network))

			// Estimated from local deployment records; `walgo projects show` and
			// `walgo status` read the authoritative expiry from chain.
			if !epochInfo.FirstDeploymentAt.IsZero() {
				if expiryDate, ok := estimateExpiry(epochInfo.FirstDeploymentAt, epochInfo.TotalEpochs, proj.Network); ok {
					fmt.Printf("   Expires:      %s (est.)\n", formatExpiryDuration(expiryDate))
				}
			}
		} else if proj.Epochs > 0 {
			fmt.Printf("   Epochs:       %s\n", formatEpochs(proj.Epochs, proj.Network))
		}

		since := time.Since(proj.LastDeployAt)
		if since < 24*time.Hour {
			fmt.Printf("   Updated:      %s ago\n", formatDuration(since))
		} else {
			fmt.Printf("   Updated:      %d days ago\n", int(since.Hours()/24))
		}
	}

	fmt.Println()
	fmt.Printf("%s Commands:\n", icons.Lightbulb)
	fmt.Println("   • Show details:  walgo projects show <name>")
	fmt.Println("   • Update site:   walgo projects update <name>")
	fmt.Println("   • Edit metadata: walgo projects edit <name> --name 'New Name'")
	fmt.Println("   • Archive:       walgo projects archive <name>")
	fmt.Println()

	return nil
}
