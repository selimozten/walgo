package cmd

import (
	"fmt"
	"os"

	"github.com/selimozten/walgo/internal/ui"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage your Walrus site projects",
	Long: `View, manage, and redeploy your Walrus site projects.

The projects command shows all your deployed sites and allows you to:
  • View project history and statistics
  • Edit project metadata locally (name, category, description, etc.)
  • Update the site on Walrus (push changes on-chain)
  • Archive or delete projects

Workflow:
  1. Edit metadata:  walgo projects edit <name> --name "New Name"
  2. Update on-chain: walgo projects update <name>

Examples:
  walgo projects                           # List all projects (default)
  walgo projects list                      # List all projects
  walgo projects list --network mainnet    # Filter by network
  walgo projects show <name>               # Show project details
  walgo projects edit <name> --name "New Name"  # Edit metadata locally
  walgo projects update <name>             # Push changes to Walrus
  walgo projects update <name> --epochs 10  # Update with new epoch count`,
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
	Use:   "show <name|id>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		if err := showProject(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to show project: %w", err)
		}

		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <name|id>",
	Short: "Update the site on Walrus (push changes on-chain)",
	Long: `Update a project's site on Walrus blockchain.

This command pushes your local site changes to the Walrus network.
Use 'walgo projects edit' first to change metadata like name, description, etc.

Examples:
  walgo projects update mysite              # Update site on Walrus
  walgo projects update mysite --epochs 10  # Update with new epoch count`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		epochs, _ := cmd.Flags().GetInt("epochs")

		if err := updateProject(args[0], epochs); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to update project: %w", err)
		}

		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <name|id>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		if err := deleteProject(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to delete project: %w", err)
		}

		return nil
	},
}

var projectsEditCmd = &cobra.Command{
	Use:   "edit <name|id>",
	Short: "Edit project metadata locally",
	Long: `Edit project metadata (name, description, category, image URL, SuiNS).

This command updates metadata locally:
  • Local database
  • ws-resources.json in publish directory

After editing, use 'walgo projects update' to push changes to Walrus.

Examples:
  walgo projects edit mysite --name "New Name"
  walgo projects edit mysite --description "My awesome site"
  walgo projects edit mysite --category blog --image-url "https://example.com/logo.png"

Then push to Walrus:
  walgo projects update mysite`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		name, _ := cmd.Flags().GetString("name")
		category, _ := cmd.Flags().GetString("category")
		description, _ := cmd.Flags().GetString("description")
		imageURL, _ := cmd.Flags().GetString("image-url")
		suins, _ := cmd.Flags().GetString("suins")

		opts := editProjectOptions{
			Name:        name,
			Category:    category,
			Description: description,
			ImageURL:    imageURL,
			SuiNS:       suins,
		}

		if err := editProject(args[0], opts); err != nil {
			fmt.Fprintf(os.Stderr, "%s Error: %v\n", icons.Error, err)
			return fmt.Errorf("failed to edit project: %w", err)
		}

		return nil
	},
}

var projectsArchiveCmd = &cobra.Command{
	Use:   "archive <name|id>",
	Short: "Archive a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		icons := ui.GetIcons()
		if err := archiveProject(args[0]); err != nil {
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

	projectsCmd.Flags().StringP("network", "n", "", "Filter by network (testnet/mainnet)")
	projectsCmd.Flags().StringP("status", "s", "", "Filter by status (active/archived)")
	projectsListCmd.Flags().StringP("network", "n", "", "Filter by network (testnet/mainnet)")
	projectsListCmd.Flags().StringP("status", "s", "", "Filter by status (active/archived)")

	projectsUpdateCmd.Flags().IntP("epochs", "e", 0, "Number of epochs for storage duration")

	projectsEditCmd.Flags().String("name", "", "New project name")
	projectsEditCmd.Flags().String("category", "", "New project category")
	projectsEditCmd.Flags().String("description", "", "New project description")
	projectsEditCmd.Flags().String("image-url", "", "New image URL for the site")
	projectsEditCmd.Flags().String("suins", "", "New SuiNS domain")
}
