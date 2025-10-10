package sitebuilder

import (
	"context"
	"walgo/internal/config"
	"walgo/internal/deployer"
	"walgo/internal/walrus"
)

// Adapter implements deployer.WalrusDeployer via the site-builder CLI.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Deploy(ctx context.Context, siteDir string, opts deployer.DeployOptions) (*deployer.Result, error) {
	walrus.SetVerbose(opts.Verbose)
	// Load config for walrus settings if available
	// Note: commands typically load config, but adapter users may pass dir directly.
	// We construct a minimal config for compatibility; epochs is taken from opts.
	cfg, _ := config.LoadConfig()
	out, err := walrus.DeploySite(siteDir, cfg.WalrusConfig, opts.Epochs)
	if err != nil {
		return nil, err
	}
	return &deployer.Result{Success: out.Success, ObjectID: out.ObjectID, BrowseURLs: out.BrowseURLs}, nil
}

func (a *Adapter) Update(ctx context.Context, siteDir string, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	out, err := walrus.UpdateSite(siteDir, objectID, opts.Epochs)
	if err != nil {
		return nil, err
	}
	return &deployer.Result{Success: out.Success, ObjectID: objectID, BrowseURLs: out.BrowseURLs}, nil
}

func (a *Adapter) Status(ctx context.Context, objectID string, opts deployer.DeployOptions) (*deployer.Result, error) {
	out, err := walrus.GetSiteStatus(objectID)
	if err != nil {
		return nil, err
	}
	return &deployer.Result{Success: out.Success, ObjectID: objectID, BrowseURLs: out.BrowseURLs}, nil
}
