//go:build desktop

package main

import (
	"context"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func windowMinimise(ctx context.Context) {
	wruntime.WindowMinimise(ctx)
}

func windowToggleMaximise(ctx context.Context) {
	wruntime.WindowToggleMaximise(ctx)
}

func appQuit(ctx context.Context) {
	wruntime.Quit(ctx)
}

func openDirectoryDialog(ctx context.Context, title, defaultDir string) (string, error) {
	return wruntime.OpenDirectoryDialog(ctx, wruntime.OpenDialogOptions{
		Title:            title,
		DefaultDirectory: defaultDir,
	})
}

func browserOpenURL(ctx context.Context, url string) {
	wruntime.BrowserOpenURL(ctx, url)
}
