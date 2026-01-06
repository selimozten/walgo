package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
)

// updateProject pushes local site changes to Walrus blockchain.
func updateProject(nameOrID string, epochs int) error {
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
	fmt.Printf("%s Updating project on Walrus: %s\n", icons.Rocket, proj.Name)
	fmt.Println()

	if epochs == 0 {
		epochs = proj.Epochs
	} else if epochs != proj.Epochs {
		fmt.Printf("  %s Epochs: %d → %d\n", icons.Info, proj.Epochs, epochs)
	}

	walgoCfg, err := config.LoadConfigFrom(proj.SitePath)
	if err != nil {
		return fmt.Errorf("failed to load config from %s: %w", proj.SitePath, err)
	}

	publishDir := filepath.Join(proj.SitePath, walgoCfg.HugoConfig.PublishDir)

	if _, err := os.Stat(publishDir); os.IsNotExist(err) {
		return fmt.Errorf("site not built - run 'walgo build' first")
	}

	fmt.Printf("%s Pushing changes to Walrus...\n", icons.Spinner)

	err = hugo.BuildSite(proj.SitePath)
	if err != nil {
		return fmt.Errorf("failed to build site: %w", err)
	}

	d := sb.New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	output, err := d.Update(ctx, publishDir, proj.ObjectID, deployer.DeployOptions{
		Epochs:    epochs,
		Verbose:   true,
		WalrusCfg: walgoCfg.WalrusConfig,
	})

	if err != nil {
		deployment := &projects.DeploymentRecord{
			ProjectID: proj.ID,
			ObjectID:  proj.ObjectID,
			Network:   proj.Network,
			Epochs:    epochs,
			Success:   false,
			Error:     err.Error(),
		}
		_ = pm.RecordDeployment(deployment)

		return fmt.Errorf("update failed: %w", err)
	}

	if !output.Success {
		return fmt.Errorf("update failed: operation unsuccessful")
	}

	proj.Epochs = epochs
	proj.LastDeployAt = time.Now()

	if err := pm.UpdateProject(proj); err != nil {
		return fmt.Errorf("failed to update project record: %w", err)
	}

	deployment := &projects.DeploymentRecord{
		ProjectID: proj.ID,
		ObjectID:  proj.ObjectID,
		Network:   proj.Network,
		Epochs:    epochs,
		Success:   true,
	}
	if err := pm.RecordDeployment(deployment); err != nil {
		fmt.Fprintf(os.Stderr, "%s Warning: Failed to record deployment: %v\n", icons.Warning, err)
	}

	fmt.Println()
	fmt.Printf("%s Update successful!\n", icons.Check)
	fmt.Printf("%s Object ID: %s\n", icons.Info, proj.ObjectID)
	fmt.Println()

	return nil
}
