package deployment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/selimozten/walgo/internal/cache"
	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/deployer"
	sb "github.com/selimozten/walgo/internal/deployer/sitebuilder"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/sui"
	"github.com/selimozten/walgo/internal/ui"
)

// DeploymentOptions contains all options for deployment
type DeploymentOptions struct {
	SitePath    string
	PublishDir  string
	Epochs      int
	WalgoCfg    *config.WalgoConfig
	Quiet       bool
	Verbose     bool
	ForceNew    bool
	DryRun      bool
	SaveProject bool
	ProjectName string // Also used as site_name in ws-resources.json
	Category    string
	Network     string
	WalletAddr  string
	// Metadata for ws-resources.json (displayed on wallets/explorers)
	Description string
	ImageURL    string
}

// DeploymentResult contains the result of a deployment
type DeploymentResult struct {
	Success      bool
	ObjectID     string
	IsUpdate     bool
	IsNewProject bool
	SiteSize     int64
	Error        error
}

// PerformDeployment handles the complete site deployment workflow
func PerformDeployment(ctx context.Context, opts DeploymentOptions) (*DeploymentResult, error) {
	icons := ui.GetIcons()
	result := &DeploymentResult{}

	if !opts.Quiet {
		fmt.Printf("%s Ensuring production URLs...\n", icons.Spinner)
	}

	var siteSize int64
	var walkErrors []string
	_ = filepath.Walk(opts.PublishDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			walkErrors = append(walkErrors, fmt.Sprintf("%s: %v", path, err))
			return nil
		}
		if !info.IsDir() {
			siteSize += info.Size()
		}
		return nil
	})

	if len(walkErrors) > 0 && !opts.Quiet {
		fmt.Fprintf(os.Stderr, "%s Warning: Encountered errors while calculating site size:\n", icons.Warning)
		for _, errMsg := range walkErrors {
			fmt.Fprintf(os.Stderr, "    - %s\n", errMsg)
		}
		fmt.Fprintf(os.Stderr, "  Size calculation may be incomplete.\n")
	}

	result.SiteSize = siteSize

	if !opts.Quiet {
		fmt.Printf("  %s Site ready: %s (%.2f MB)\n", icons.Check, opts.PublishDir, float64(siteSize)/(1024*1024))
	}

	var cacheHelper *cache.DeployHelper
	if !opts.Quiet {
		fmt.Println("  [1/5] Initializing cache...")
	}
	var err error
	cacheHelper, err = cache.NewDeployHelper(opts.SitePath)
	if err != nil {
		if !opts.Quiet {
			fmt.Fprintf(os.Stderr, "%s Warning: Cache initialization failed: %v\n", icons.Warning, err)
			fmt.Fprintf(os.Stderr, "  Continuing without incremental build optimization...\n")
		}
		cacheHelper = nil
	} else {
		defer func() {
			if closeErr := cacheHelper.Close(); closeErr != nil && !opts.Quiet {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to close cache: %v\n", icons.Warning, closeErr)
			}
		}()
		if !opts.Quiet {
			fmt.Printf("%s Cache initialized\n", icons.Check)
		}
	}

	if cacheHelper != nil {
		if !opts.Quiet {
			fmt.Println("  [2/5] Analyzing changes...")
		}
		plan, err := cacheHelper.PrepareDeployment(opts.PublishDir)
		if err != nil {
			if !opts.Quiet {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to analyze changes: %v\n", icons.Warning, err)
			}
		} else {
			if opts.Verbose {
				plan.PrintVerboseSummary()
			} else if !opts.Quiet {
				plan.PrintSummary()
			}

			if opts.DryRun {
				if !opts.Quiet {
					fmt.Printf("\n%s Dry-run mode: No files will be uploaded\n", icons.Info)
					fmt.Printf("%s Deployment plan complete!\n", icons.Check)
					fmt.Printf("\n%s To actually deploy, run without --dry-run flag\n", icons.Lightbulb)
				}
				result.Success = true
				return result, nil
			}
		}
	}

	// Detect if this is an update or new deployment
	var existingObjectID string
	var isUpdate bool

	// Check 1: walgo.yaml projectID
	if opts.WalgoCfg.WalrusConfig.ProjectID != "" && opts.WalgoCfg.WalrusConfig.ProjectID != "YOUR_WALRUS_PROJECT_ID" {
		existingObjectID = opts.WalgoCfg.WalrusConfig.ProjectID
		if !opts.Quiet {
			fmt.Printf("  %s Found objectID in walgo.yaml: %s\n", icons.Info, existingObjectID)
		}
	}

	// Check 2: ws-resources.json objectId
	if existingObjectID == "" {
		wsResourcesPath := filepath.Join(opts.PublishDir, "ws-resources.json")
		if wsConfig, err := compress.ReadWSResourcesConfig(wsResourcesPath); err == nil && wsConfig.ObjectID != "" {
			existingObjectID = wsConfig.ObjectID
			if !opts.Quiet {
				fmt.Printf("  %s Found objectID in ws-resources.json: %s\n", icons.Info, existingObjectID)
			}
		}
	}

	// Check 3: Database for existing project
	if existingObjectID == "" {
		pm, err := projects.NewManager()
		if err == nil {
			defer func() {
				if closeErr := pm.Close(); closeErr != nil && !opts.Quiet {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to close database: %v\n", icons.Warning, closeErr)
				}
			}()
			if proj, err := pm.GetProjectBySitePath(opts.SitePath); err == nil && proj != nil && proj.ObjectID != "" {
				existingObjectID = proj.ObjectID
				if !opts.Quiet {
					fmt.Printf("  %s Found existing project in database: %s (objectID: %s)\n", icons.Info, proj.Name, existingObjectID)
				}
			}
		}
	}

	if existingObjectID != "" && !opts.ForceNew {
		isUpdate = true
		result.IsUpdate = true
		if !opts.Quiet {
			fmt.Printf("  %s This site was already deployed - will UPDATE existing site\n", icons.Info)
			if opts.ForceNew {
				fmt.Printf("  %s To deploy as new site instead, use: --force-new\n", icons.Lightbulb)
			}
		}
	}

	// Update ws-resources.json with metadata BEFORE deployment (so it's included in the upload)
	stepNum := 3
	if cacheHelper == nil {
		stepNum = 2
	}
	if !opts.Quiet {
		fmt.Printf("  [%d/5] Preparing metadata...\n", stepNum)
	}

	wsResourcesPath := filepath.Join(opts.PublishDir, "ws-resources.json")
	metadataOpts := compress.MetadataOptions{
		SiteName:    opts.ProjectName,
		Description: opts.Description,
		ImageURL:    opts.ImageURL,
		Category:    opts.Category,
	}
	// For updates, preserve existing objectID
	if isUpdate && existingObjectID != "" {
		metadataOpts.ObjectID = existingObjectID
	}
	if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
		result.Error = fmt.Errorf("failed to prepare ws-resources.json metadata: %w", err)
		return result, result.Error
	}
	if !opts.Quiet {
		fmt.Printf("%s Metadata prepared in ws-resources.json\n", icons.Check)
	}

	// Deploy or update the site
	stepNum++
	if !opts.Quiet {
		if isUpdate {
			fmt.Printf("  [%d/5] Updating site...\n", stepNum)
		} else {
			fmt.Printf("  [%d/5] Uploading site...\n", stepNum)
		}
	}

	d := sb.New()
	uploadStart := time.Now()

	var output *deployer.Result

	if isUpdate {
		// Update existing site
		output, err = d.Update(ctx, opts.PublishDir, existingObjectID, deployer.DeployOptions{
			Epochs:    opts.Epochs,
			Verbose:   opts.Verbose && !opts.Quiet,
			WalrusCfg: opts.WalgoCfg.WalrusConfig,
		})
	} else {
		// Deploy new site
		output, err = d.Deploy(ctx, opts.PublishDir, deployer.DeployOptions{
			Epochs:    opts.Epochs,
			Verbose:   opts.Verbose && !opts.Quiet,
			WalrusCfg: opts.WalgoCfg.WalrusConfig,
		})
	}

	if err != nil {
		result.Error = err
		return result, fmt.Errorf("deployment failed: %w", err)
	}

	if !output.Success || output.ObjectID == "" {
		result.Error = fmt.Errorf("deployment failed: no object ID returned")
		return result, result.Error
	}

	result.Success = true
	result.ObjectID = output.ObjectID

	if !opts.Quiet {
		fmt.Printf("  %s Deployment completed in %v\n", icons.Check, time.Since(uploadStart).Round(time.Second))
	}

	// Update cache with deployment info
	if cacheHelper != nil {
		stepNum++
		if !opts.Quiet {
			fmt.Printf("  [%d/5] Updating cache...\n", stepNum)
		}
		err := cacheHelper.FinalizeDeployment(opts.PublishDir, output.ObjectID, output.ObjectID, output.FileToBlobID)
		if err != nil {
			if !opts.Quiet {
				fmt.Fprintf(os.Stderr, "%s Warning: Failed to update cache: %v\n", icons.Warning, err)
			}
		} else if !opts.Quiet {
			fmt.Printf("%s Cache updated\n", icons.Check)
		}
	}

	// Save object_id to local ws-resources.json (for reference, not on-chain)
	stepNum++
	if !opts.Quiet {
		fmt.Printf("  [%d/5] Saving deployment info...\n", stepNum)
	}

	// Update local ws-resources.json with object_id
	if err := compress.UpdateObjectID(wsResourcesPath, output.ObjectID); err != nil {
		result.Error = fmt.Errorf("failed to save object_id to ws-resources.json: %w", err)
		return result, result.Error
	}
	if !opts.Quiet {
		fmt.Printf("%s Saved object_id to ws-resources.json\n", icons.Check)
	}

	// Update walgo.yaml with projectID
	if err := config.UpdateWalgoYAMLProjectID(opts.SitePath, output.ObjectID); err != nil {
		result.Error = fmt.Errorf("failed to update walgo.yaml with Object ID: %w", err)
		return result, result.Error
	}
	if !opts.Quiet {
		fmt.Printf("%s Updated walgo.yaml with Object ID\n", icons.Check)
	}

	// Optionally save to projects database
	if opts.SaveProject && !opts.Quiet {
		fmt.Printf("\n%s Saving project...\n", icons.Database)

		pm, err := projects.NewManager()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Warning: Failed to save project: %v\n", icons.Warning, err)
		} else {
			defer func() {
				if closeErr := pm.Close(); closeErr != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to close database: %v\n", icons.Warning, closeErr)
				}
			}()

			// Get network and wallet if not provided
			network := opts.Network
			walletAddr := opts.WalletAddr

			if network == "" {
				network, err = sui.GetActiveEnv()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to get active network: %v\n", icons.Warning, err)
					network = "testnet"
				}
			}

			if walletAddr == "" {
				walletAddr, err = sui.GetActiveAddress()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to get active address: %v\n", icons.Warning, err)
					walletAddr = ""
				}
			}

			// Check if project already exists in database
			if !opts.Quiet {
				fmt.Printf("%s Checking for existing project with path: %s\n", icons.Info, opts.SitePath)
			}
			existingProj, err := pm.GetProjectBySitePath(opts.SitePath)
			if err != nil && !opts.Quiet {
				fmt.Fprintf(os.Stderr, "%s Warning: Error checking for existing project: %v\n", icons.Warning, err)
			}
			if existingProj != nil && !opts.Quiet {
				fmt.Printf("%s Found existing project: ID=%d, Name=%s, Status=%s, ObjectID=%s\n",
					icons.Info, existingProj.ID, existingProj.Name, existingProj.Status, existingProj.ObjectID)
			} else if !opts.Quiet {
				fmt.Printf("%s No existing project found for this path\n", icons.Info)
			}

			if existingProj != nil {
				// UPDATE existing project (whether it's draft, active, or archived)
				// If it was a draft, activate it now
				if existingProj.Status == "draft" {
					existingProj.Status = "active"
					if !opts.Quiet {
						fmt.Printf("%s Activating draft project: %s\n", icons.Info, existingProj.Name)
					}
				}

				existingProj.ObjectID = output.ObjectID
				existingProj.Network = network
				existingProj.WalletAddr = walletAddr
				existingProj.Epochs = opts.Epochs
				existingProj.LastDeployAt = time.Now()

				// Update metadata if provided
				if opts.Description != "" {
					existingProj.Description = opts.Description
				}
				if opts.ImageURL != "" {
					existingProj.ImageURL = opts.ImageURL
				}

				if err := pm.UpdateProject(existingProj); err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to update project: %v\n", icons.Warning, err)
				} else {
					// Record the deployment with epoch-aware cost estimation
					estimatedGas := projects.EstimateGasFeeWithEpochs(network, siteSize, opts.Epochs)
					deployment := &projects.DeploymentRecord{
						ProjectID: existingProj.ID,
						ObjectID:  output.ObjectID,
						Network:   network,
						Epochs:    opts.Epochs,
						GasFee:    estimatedGas,
						Success:   true,
					}
					if err := pm.RecordDeployment(deployment); err != nil {
						fmt.Fprintf(os.Stderr, "%s Warning: Failed to record deployment history: %v\n", icons.Warning, err)
					}

					fmt.Printf("%s Project updated in database\n", icons.Check)
					result.IsNewProject = false
				}
			} else {
				// CREATE new project
				projectName := opts.ProjectName
				if projectName == "" {
					projectName = filepath.Base(opts.SitePath)
					if projectName == "" || projectName == "." || projectName == "/" {
						projectName = "my-walgo-site"
					}
				}

				category := opts.Category
				if category == "" {
					category = "website"
				}

				project := &projects.Project{
					Name:        projectName,
					Category:    category,
					Network:     network,
					ObjectID:    output.ObjectID,
					WalletAddr:  walletAddr,
					Epochs:      opts.Epochs,
					SitePath:    opts.SitePath,
					Description: opts.Description,
					ImageURL:    opts.ImageURL,
				}

				if err := pm.CreateProject(project); err != nil {
					fmt.Fprintf(os.Stderr, "%s Warning: Failed to create project: %v\n", icons.Warning, err)
				} else {
					// Record the deployment with epoch-aware cost estimation
					estimatedGas := projects.EstimateGasFeeWithEpochs(network, siteSize, opts.Epochs)
					deployment := &projects.DeploymentRecord{
						ProjectID: project.ID,
						ObjectID:  output.ObjectID,
						Network:   network,
						Epochs:    opts.Epochs,
						GasFee:    estimatedGas,
						Success:   true,
					}
					if err := pm.RecordDeployment(deployment); err != nil {
						fmt.Fprintf(os.Stderr, "%s Warning: Failed to record deployment history: %v\n", icons.Warning, err)
					}

					fmt.Printf("%s Project saved - manage with 'walgo projects'\n", icons.Check)
					result.IsNewProject = true
				}
			}
		}
	}

	return result, nil
}
