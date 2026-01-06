package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
)

// deleteProject removes a project from Walrus blockchain and local database.
func deleteProject(nameOrID string) error {
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

	fmt.Println()
	fmt.Printf("%s Delete project: %s?\n", icons.Warning, proj.Name)
	fmt.Println()
	fmt.Println("This will:")
	fmt.Println("  - Delete the site from Walrus blockchain (on-chain deletion)")
	fmt.Println("  - Delete the project record from local database")
	fmt.Println("  - Delete all deployment history")
	fmt.Println()
	fmt.Printf("Object ID to destroy: %s\n", proj.ObjectID)

	destroyCost := projects.EstimateDestroyCost(proj.Network)
	fmt.Printf("Estimated gas cost: %s\n", destroyCost)
	fmt.Println()
	fmt.Print("Are you sure? [y/N]: ")

	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		fmt.Printf("%s Error reading input: %v\n", icons.Error, err)
		return fmt.Errorf("failed to read confirmation: %w", err)
	}
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Printf("%s Cancelled\n", icons.Cross)
		return nil
	}

	if proj.ObjectID != "" {
		if !strings.HasPrefix(proj.ObjectID, "0x") || len(proj.ObjectID) < 10 {
			fmt.Printf("%s Warning: Object ID '%s' appears invalid\n", icons.Warning, proj.ObjectID)
			fmt.Print("Continue anyway? [y/N]: ")
			var confirmInvalid string
			if _, err := fmt.Scanln(&confirmInvalid); err != nil {
				fmt.Printf("%s Error reading input: %v\n", icons.Error, err)
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			if strings.ToLower(strings.TrimSpace(confirmInvalid)) != "y" {
				fmt.Printf("%s Cancelled\n", icons.Cross)
				return nil
			}
		}

		fmt.Println()
		fmt.Printf("%s Step 1/2: Destroying site on Walrus blockchain...\n", icons.Garbage)

		d := sb.New()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := d.Destroy(ctx, proj.ObjectID); err != nil {
			fmt.Printf("\n%s Warning: Failed to destroy site on-chain: %v\n", icons.Warning, err)
			fmt.Println()
			fmt.Print("Continue with local deletion anyway? [y/N]: ")

			var continueConfirm string
			if _, err := fmt.Scanln(&continueConfirm); err != nil {
				fmt.Printf("%s Error reading input: %v\n", icons.Error, err)
				return fmt.Errorf("failed to read confirmation: %w", err)
			}
			continueConfirm = strings.TrimSpace(strings.ToLower(continueConfirm))

			if continueConfirm != "y" && continueConfirm != "yes" {
				fmt.Printf("%s Cancelled\n", icons.Cross)
				return nil
			}
		} else {
			fmt.Printf("%s Site destroyed on-chain\n", icons.Check)
		}
	}

	fmt.Println()
	fmt.Printf("%s Step 2/2: Deleting local project record...\n", icons.Garbage)

	if err := pm.DeleteProject(proj.ID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s Project deleted successfully\n", icons.Check)
	fmt.Println()

	return nil
}
