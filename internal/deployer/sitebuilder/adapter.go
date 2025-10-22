package sitebuilder

import (
	"context"
	"walgo/internal/deployer"
	"walgo/internal/walrus"
)

// Adapter implements deployer.WalrusDeployer via the site-builder CLI.
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Deploy(ctx context.Context, siteDir string, opts deployer.DeployOptions) (*deployer.Result, error) {
	walrus.SetVerbose(opts.Verbose)
	out, err := walrus.DeploySite(siteDir, opts.WalrusCfg, opts.Epochs)
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
	rc := 0
	if out != nil {
		rc = len(out.Resources)
	}
	return &deployer.Result{Success: out.Success, ObjectID: objectID, BrowseURLs: out.BrowseURLs, ResourceCount: rc}, nil
}
