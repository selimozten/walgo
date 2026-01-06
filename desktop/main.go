package main

import (
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Base application options
	appOptions := &options.App{
		Title:             "Walgo",
		Width:             1280,
		Height:            800,
		Fullscreen:        false, // Start windowed, user can toggle fullscreen
		Frameless:         true,
		MinWidth:          800,
		MinHeight:         600,
		DisableResize:     false,
		AlwaysOnTop:       false,
		StartHidden:       false,
		HideWindowOnClose: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	}

	// Configure platform-specific options
	configurePlatformOptions(appOptions)

	// Run the application
	err := wails.Run(appOptions)

	if err != nil {
		println("Error:", err.Error())
	}
}

// configurePlatformOptions sets platform-specific options for the application
func configurePlatformOptions(appOptions *options.App) {
	switch runtime.GOOS {
	case "darwin": // macOS
		appOptions.Frameless = false // Use native macOS window with traffic lights
		appOptions.Mac = &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(), // Hidden but shows traffic lights
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "Walgo",
				Message: "A modern desktop application for Walgo\n\n© 2026 Walgo Team",
				Icon:    nil, // Uses default app icon
			},
		}

	case "windows":
		// Windows-specific configuration
		appOptions.Windows = &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
			WebviewUserDataPath:               "", // Uses default
			Theme:                             windows.SystemDefault,
			BackdropType:                      windows.Auto,
		}

	case "linux":
		// Linux-specific configuration
		appOptions.Linux = &linux.Options{
			Icon:                nil, // Uses icon from wails.json
			WindowIsTranslucent: false,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways,
		}

	default:
		// Fallback for other platforms - use safe defaults
		appOptions.Frameless = false
	}
}
