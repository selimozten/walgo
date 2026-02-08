package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/hugo"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/selimozten/walgo/internal/walrus"
)

// updateProjectByRef pushes local site changes to Walrus blockchain.
func updateProjectByRef(proj *projects.Project, epochs int) error {
	icons := ui.GetIcons()
	pm, err := projects.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

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

	// Calculate site size for gas estimation
	var siteSize int64
	_ = filepath.Walk(publishDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			siteSize += info.Size()
		}
		return nil
	})

	// Try to get actual gas from blockchain, fallback to estimate
	var gasFee string
	gasInfo, err := walrus.GetLatestTransactionGas(proj.WalletAddr, proj.Network)
	if err == nil && gasInfo != nil {
		if gasInfo.TotalWAL > 0 && gasInfo.TotalGasSUI > 0 {
			gasFee = fmt.Sprintf("%.6f WAL + %.6f SUI", gasInfo.TotalWAL, gasInfo.TotalGasSUI)
		} else if gasInfo.TotalWAL > 0 {
			gasFee = fmt.Sprintf("%.6f WAL", gasInfo.TotalWAL)
		} else if gasInfo.TotalGasSUI > 0 {
			gasFee = fmt.Sprintf("%.6f SUI", gasInfo.TotalGasSUI)
		}
	}
	// Fallback to estimate if actual gas not available
	if gasFee == "" {
		gasFee = projects.EstimateGasFeeWithEpochs(proj.Network, siteSize, epochs)
	}

	proj.Epochs = epochs
	proj.LastDeployAt = time.Now()
	proj.GasFee = gasFee

	if err := pm.UpdateProject(proj); err != nil {
		return fmt.Errorf("failed to update project record: %w", err)
	}

	deployment := &projects.DeploymentRecord{
		ProjectID: proj.ID,
		ObjectID:  proj.ObjectID,
		Network:   proj.Network,
		Epochs:    epochs,
		GasFee:    gasFee,
		Success:   true,
	}
	if err := pm.RecordDeployment(deployment); err != nil {
		fmt.Fprintf(os.Stderr, "%s Warning: Failed to record deployment: %v\n", icons.Warning, err)
	}

	fmt.Println()
	fmt.Printf("%s Update successful!\n", icons.Check)
	fmt.Printf("%s Object ID: %s\n", icons.Info, proj.ObjectID)
	if gasFee != "" {
		fmt.Printf("%s Gas Fee: %s\n", icons.Info, gasFee)
	}
	fmt.Println()

	return nil
}
