package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/selimozten/walgo/internal/projects"
	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

// resolveProject resolves a project by --name, --id flags, or positional argument.
// Priority: --id flag > --name flag > positional argument
// Returns the resolved project or an error.
func resolveProject(cmd *cobra.Command, args []string) (*projects.Project, error) {
	pm, err := projects.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize project manager: %w", err)
	}
	defer pm.Close()

	// Get flags
	idFlag, _ := cmd.Flags().GetInt64("id")
	nameFlag, _ := cmd.Flags().GetString("name")

	// Priority 1: --id flag
	if idFlag > 0 {
		proj, err := pm.GetProject(idFlag)
		if err != nil {
			return nil, fmt.Errorf("project with ID %d not found: %w", idFlag, err)
		}
		return proj, nil
	}

	// Priority 2: --name flag
	if nameFlag != "" {
		proj, err := pm.GetProjectByName(nameFlag)
		if err != nil {
			return nil, fmt.Errorf("project with name '%s' not found: %w", nameFlag, err)
		}
		return proj, nil
	}

	// Priority 3: positional argument (backward compatibility)
	if len(args) > 0 {
		nameOrID := args[0]
		// Try parsing as ID first
		if id, err := strconv.ParseInt(nameOrID, 10, 64); err == nil {
			proj, err := pm.GetProject(id)
			if err != nil {
				return nil, fmt.Errorf("project with ID %d not found: %w", id, err)
			}
			return proj, nil
		}
		// Try as name
		proj, err := pm.GetProjectByName(nameOrID)
		if err != nil {
			return nil, fmt.Errorf("project '%s' not found: %w", nameOrID, err)
		}
		return proj, nil
	}

	return nil, fmt.Errorf("please specify a project using --name, --id, or as a positional argument")
}

// addProjectIdentifierFlags adds --name and --id flags to a command
func addProjectIdentifierFlags(cmd *cobra.Command) {
	cmd.Flags().Int64("id", 0, "Project ID")
	cmd.Flags().String("name", "", "Project name (supports names with spaces)")
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage your Walrus site projects",
	Long: `View, manage, and redeploy your Walrus site projects.

The projects command shows all your deployed sites and allows you to:
  • View project history and statistics
  • Edit project metadata locally (name, category, description, etc.)
  • Update the site on Walrus (push changes on-chain)
  • Archive or delete projects

Project Identification:
  You can identify projects using:
  • --id=<number>     Project ID (unambiguous)
  • --name="<name>"   Project name (supports spaces)
  • <name|id>         Positional argument (legacy, no spaces)

Workflow:
  1. Edit metadata:  walgo projects edit --name="My Site" --description="New desc"
  2. Update on-chain: walgo projects update --name="My Site"

Examples:
  walgo projects                              # List all projects (default)
  walgo projects list                         # List all projects
  walgo projects list --network mainnet       # Filter by network
  walgo projects show --name="My Site"        # Show project with spaces in name
  walgo projects show --id=5                  # Show project by ID
  walgo projects show mysite                  # Show project (legacy syntax)
  walgo projects edit --id=5 --description="New description"
  walgo projects update --name="My Site" --epochs 10`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		network, _ := cmd.Flags().GetString("network")
		status, _ := cmd.Flags().GetString("status")

		if err := listProjects(network, status); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to list projects: %w", err)
		}

		return nil
	},
}

var projectsShowCmd = &cobra.Command{
	Use:   "show [name|id]",
	Short: "Show project details",
	Long: `Show detailed information about a project.

Project Identification:
  --id=<number>     Project ID (unambiguous)
  --name="<name>"   Project name (supports spaces)
  <name|id>         Positional argument (legacy, no spaces)

Examples:
  walgo projects show --name="My Site"    # Name with spaces
  walgo projects show --id=5              # By ID
  walgo projects show mysite              # Legacy syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		proj, err := resolveProject(cmd, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}
		if err := showProjectDetails(proj); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to show project: %w", err)
		}

		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update [name|id]",
	Short: "Update the site on Walrus (push changes on-chain)",
	Long: `Update a project's site on Walrus blockchain.

This command pushes your local site changes to the Walrus network.
Use 'walgo projects edit' first to change metadata like name, description, etc.

Project Identification:
  --id=<number>     Project ID (unambiguous)
  --name="<name>"   Project name (supports spaces)
  <name|id>         Positional argument (legacy, no spaces)

Examples:
  walgo projects update --name="My Site"              # Update site with spaces in name
  walgo projects update --id=5 --epochs 10            # Update by ID with epochs
  walgo projects update mysite                        # Legacy syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		proj, err := resolveProject(cmd, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}
		epochs, _ := cmd.Flags().GetInt("epochs")

		if err := updateProjectByRef(proj, epochs); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to update project: %w", err)
		}

		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete [name|id]",
	Short: "Delete a project",
	Long: `Delete a project from Walrus blockchain and local database.

Project Identification:
  --id=<number>     Project ID (unambiguous)
  --name="<name>"   Project name (supports spaces)
  <name|id>         Positional argument (legacy, no spaces)

Examples:
  walgo projects delete --name="My Site"    # Delete by name with spaces
  walgo projects delete --id=5              # Delete by ID
  walgo projects delete mysite              # Legacy syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		proj, err := resolveProject(cmd, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}
		if err := deleteProjectByRef(proj); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	},
}

var projectsEditCmd = &cobra.Command{
	Use:   "edit [name|id]",
	Short: "Edit project metadata locally",
	Long: `Edit project metadata (new-name, description, category, image URL, SuiNS).

This command updates metadata locally:
  • Local database
  • ws-resources.json in publish directory

After editing, use 'walgo projects update' to push changes to Walrus.

Project Identification:
  --id=<number>     Project ID (unambiguous)
  --name="<name>"   Project name (supports spaces) - identifies the project
  <name|id>         Positional argument (legacy, no spaces)

Note: Use --new-name to rename the project, --name is for identification.

Examples:
  walgo projects edit --id=5 --new-name="New Name"
  walgo projects edit --name="My Site" --description="My awesome site"
  walgo projects edit --id=5 --category blog --image-url "https://example.com/logo.png"

Then push to Walrus:
  walgo projects update --id=5`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		proj, err := resolveProject(cmd, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}

		newName, _ := cmd.Flags().GetString("new-name")
		category, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		imageURL, _ := cmd.Flags().GetString("image-url")
		suins, _ := cmd.Flags().GetString("suins")

		opts := editProjectOptions{
			Name:        newName,
			Category:    category,
			Description: description,
			ImageURL:    imageURL,
			SuiNS:       suins,
		}

		if err := editProjectByRef(proj, opts); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to edit project: %w", err)
		}

		return nil
	},
}

var projectsArchiveCmd = &cobra.Command{
	Use:   "archive [name|id]",
	Short: "Archive a project",
	Long: `Archive a project without deleting it.

Project Identification:
  --id=<number>     Project ID (unambiguous)
  --name="<name>"   Project name (supports spaces)
  <name|id>         Positional argument (legacy, no spaces)

Examples:
  walgo projects archive --name="My Site"   # Archive by name with spaces
  walgo projects archive --id=5             # Archive by ID
  walgo projects archive mysite             # Legacy syntax`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		proj, err := resolveProject(cmd, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return err
		}
		if err := archiveProjectByRef(proj); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to archive project: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsShowCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	projectsCmd.AddCommand(projectsEditCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
	projectsCmd.AddCommand(projectsArchiveCmd)

	projectsCmd.RunE = func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		network, _ := cmd.Flags().GetString("network")
		status, _ := cmd.Flags().GetString("status")

		if err := listProjects(network, status); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to list projects: %w", err)
		}

		return nil
	}

	// List command flags
	projectsCmd.Flags().StringP("network", "n", "", "Filter by network (testnet/mainnet)")
	projectsCmd.Flags().StringP("status", "s", "", "Filter by status (active/archived)")
	projectsListCmd.Flags().StringP("network", "n", "", "Filter by network (testnet/mainnet)")
	projectsListCmd.Flags().StringP("status", "s", "", "Filter by status (active/archived)")

	// Add project identifier flags to all subcommands
	addProjectIdentifierFlags(projectsShowCmd)
	addProjectIdentifierFlags(projectsUpdateCmd)
	addProjectIdentifierFlags(projectsDeleteCmd)
	addProjectIdentifierFlags(projectsEditCmd)
	addProjectIdentifierFlags(projectsArchiveCmd)

	// Update command specific flags
	projectsUpdateCmd.Flags().IntP("epochs", "e", 0, "Number of epochs for storage duration")

	// Edit command specific flags
	projectsEditCmd.Flags().String("new-name", "", "New project name (rename)")
	projectsEditCmd.Flags().String("category", "", "New project category")
	projectsEditCmd.Flags().String("description", "", "New project description")
	projectsEditCmd.Flags().String("image-url", "", "New image URL for the site")
	projectsEditCmd.Flags().String("suins", "", "New SuiNS domain")
}
