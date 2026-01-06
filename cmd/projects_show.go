package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
)

// showProject displays detailed information about a specific project.
func showProject(nameOrID string) error {
	icons := ui.GetIcons()
	pm, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

	var proj *projects.Project
	if id, err := strconv.ParseInt(nameOrID, 10, 64); err == nil {
		proj, err = pm.GetProject(id)
		if err != nil {
			return err
		}
	} else {
		proj, err = pm.GetProjectByName(nameOrID)
		if err != nil {
			return err
		}
	}

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
		fmt.Printf("  URL:             https://%s.walrus.site\n", proj.SuiNS)
	}
	fmt.Printf("  Wallet:          %s\n", proj.WalletAddr)
	fmt.Printf("  Site path:       %s\n", proj.SitePath)
	fmt.Printf("  Created:         %s\n", proj.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("%s Statistics\n", icons.Info)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("  Total deployments:     %d\n", stats.TotalDeployments)
	fmt.Printf("  Successful:            %d\n", stats.SuccessfulDeploys)
	if stats.FailedDeploys > 0 {
		fmt.Printf("  Failed:                %d\n", stats.FailedDeploys)
	}
	if !stats.FirstDeployment.IsZero() {
		fmt.Printf("  First deployment:      %s\n", stats.FirstDeployment.Format("2006-01-02 15:04"))
		fmt.Printf("  Last deployment:       %s\n", stats.LastDeployment.Format("2006-01-02 15:04"))
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
