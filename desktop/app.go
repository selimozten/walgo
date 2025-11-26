package main

import (
	"context"
	"walgo/pkg/api"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// CreateSite initializes a new Hugo site with Walrus configuration
func (a *App) CreateSite(parentDir string, name string) error {
	return api.CreateSite(parentDir, name)
}

// BuildSite builds the Hugo site at the given path
func (a *App) BuildSite(sitePath string) error {
	return api.BuildSite(sitePath)
}

// DeployResult holds the result of a deployment
type DeployResult struct {
	Success  bool   `json:"success"`
	ObjectID string `json:"objectId"`
	Error    string `json:"error"`
}

// DeploySite deploys the site to Walrus
func (a *App) DeploySite(sitePath string, epochs int) DeployResult {
	res := api.DeploySite(sitePath, epochs)
	return DeployResult{
		Success:  res.Success,
		ObjectID: res.ObjectID,
		Error:    res.Error,
	}
}

// SelectDirectory opens a dialog to select a directory (helper for frontend)
// Note: Wails runtime has dialogs, but we can expose a helper if needed.
// For now, we'll let frontend use runtime.WindowOpenDirectoryDialog via Wails JS runtime.
