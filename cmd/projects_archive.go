package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
)

// archiveProject marks a project as archived without deleting it.
func archiveProject(nameOrID string) error {
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

	if err := pm.ArchiveProject(proj.ID); err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	fmt.Println()
	fmt.Printf("%s Project '%s' archived\n", icons.Check, proj.Name)
	fmt.Println()
	fmt.Printf("%s To restore: walgo projects restore <name>\n", icons.Lightbulb)
	fmt.Println()

	return nil
}

// formatDuration returns a human-readable duration string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
