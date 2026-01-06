package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/selimozten/walgo/internal/compress"
	"github.com/selimozten/walgo/internal/config"
	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
)

// editProjectOptions contains metadata fields for project editing.
type editProjectOptions struct {
	Name        string
	Category    string
	Description string
	ImageURL    string
	SuiNS       string
}

// editProject updates project metadata in database and ws-resources.json.
func editProject(nameOrID string, opts editProjectOptions) error {
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

	if opts.Name == "" && opts.Category == "" && opts.Description == "" && opts.ImageURL == "" && opts.SuiNS == "" {
		return fmt.Errorf("no changes specified. Use --name, --category, --description, --image-url, or --suins flags")
	}

	fmt.Println()
	fmt.Printf("%s Editing project: %s\n", icons.Pencil, proj.Name)
	fmt.Println()

	changes := []string{}

	if opts.Name != "" && opts.Name != proj.Name {
		exists, err := pm.ProjectNameExists(opts.Name)
		if err == nil && exists {
			existingProj, _ := pm.GetProjectByName(opts.Name)
			if existingProj != nil && existingProj.ID != proj.ID {
				return fmt.Errorf("project name '%s' already exists. Choose a different name", opts.Name)
			}
		}
		changes = append(changes, fmt.Sprintf("Name: %s → %s", proj.Name, opts.Name))
		proj.Name = opts.Name
	}

	if opts.Category != "" && opts.Category != proj.Category {
		oldCat := proj.Category
		if oldCat == "" {
			oldCat = "(empty)"
		}
		changes = append(changes, fmt.Sprintf("Category: %s → %s", oldCat, opts.Category))
		proj.Category = opts.Category
	}

	if opts.Description != "" && opts.Description != proj.Description {
		oldDesc := proj.Description
		if oldDesc == "" {
			oldDesc = "(empty)"
		}
		displayNew := opts.Description
		if len(displayNew) > 50 {
			displayNew = displayNew[:47] + "..."
		}
		displayOld := oldDesc
		if len(displayOld) > 50 {
			displayOld = displayOld[:47] + "..."
		}
		changes = append(changes, fmt.Sprintf("Description: %s → %s", displayOld, displayNew))
		proj.Description = opts.Description
	}

	if opts.ImageURL != "" && opts.ImageURL != proj.ImageURL {
		oldURL := proj.ImageURL
		if oldURL == "" {
			oldURL = "(default)"
		}
		changes = append(changes, fmt.Sprintf("Image URL: %s → %s", oldURL, opts.ImageURL))
		proj.ImageURL = opts.ImageURL
	}

	if opts.SuiNS != "" && opts.SuiNS != proj.SuiNS {
		oldSuiNS := proj.SuiNS
		if oldSuiNS == "" {
			oldSuiNS = "(none)"
		}
		changes = append(changes, fmt.Sprintf("SuiNS: %s → %s", oldSuiNS, opts.SuiNS))
		proj.SuiNS = opts.SuiNS
	}

	if len(changes) == 0 {
		fmt.Println("No changes to apply (values are the same)")
		return nil
	}

	fmt.Println("Changes to apply:")
	for _, change := range changes {
		fmt.Printf("  %s %s\n", icons.Pencil, change)
	}
	fmt.Println()

	fmt.Printf("%s Step 1/2: Updating database...\n", icons.Database)
	if err := pm.UpdateProject(proj); err != nil {
		return fmt.Errorf("failed to update project in database: %w", err)
	}
	fmt.Printf("  %s Database updated\n", icons.Check)

	fmt.Printf("%s Step 2/2: Updating ws-resources.json...\n", icons.File)

	walgoCfg, err := config.LoadConfigFrom(proj.SitePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s Warning: Could not load config from %s: %v\n", icons.Warning, proj.SitePath, err)
		fmt.Fprintf(os.Stderr, "  %s Skipping ws-resources.json update\n", icons.Info)
	} else {
		publishDir := filepath.Join(proj.SitePath, walgoCfg.HugoConfig.PublishDir)
		wsResourcesPath := filepath.Join(publishDir, "ws-resources.json")

		if _, err := os.Stat(publishDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "  %s Warning: Publish directory does not exist: %s\n", icons.Warning, publishDir)
			fmt.Fprintf(os.Stderr, "  %s Run 'walgo build' first to create the publish directory\n", icons.Lightbulb)
		} else {
			metadataOpts := compress.MetadataOptions{
				SiteName:    proj.Name,
				Description: proj.Description,
				ImageURL:    proj.ImageURL,
				Category:    proj.Category,
				Creator:     compress.DefaultCreator,
			}

			if proj.ObjectID != "" {
				metadataOpts.ObjectID = proj.ObjectID
			}

			if err := compress.UpdateMetadata(wsResourcesPath, metadataOpts); err != nil {
				fmt.Fprintf(os.Stderr, "  %s Error: Failed to update ws-resources.json: %v\n", icons.Error, err)
				fmt.Fprintf(os.Stderr, "  %s Run 'walgo build' to regenerate the publish directory\n", icons.Lightbulb)
				return fmt.Errorf("failed to update ws-resources.json: %w", err)
			}
			fmt.Printf("  %s ws-resources.json updated\n", icons.Check)
		}
	}

	fmt.Println()
	fmt.Printf("%s Project metadata updated locally!\n", icons.Check)
	fmt.Println()
	fmt.Printf("%s To push changes to Walrus, run:\n", icons.Lightbulb)
	fmt.Printf("   walgo projects update %s\n", proj.Name)
	fmt.Println()

	return nil
}
