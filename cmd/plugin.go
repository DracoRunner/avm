package cmd

import (
	"avm/internal/plugin"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage avm plugins",
}

var pluginAddCmd = &cobra.Command{
	Use:   "add <url|path>",
	Short: "Add a plugin from a git URL or local path",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		fmt.Printf("Installing plugin from %s...\n", source)
		if err := plugin.InstallPlugin(source); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(color.GreenString("Plugin installed successfully."))
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed plugins",
	Run: func(cmd *cobra.Command, args []string) {
		pluginDir := plugin.GetPluginDir()
		dirs, err := os.ReadDir(pluginDir)
		if err != nil {
			fmt.Println("No plugins installed.")
			return
		}

		fmt.Println(color.CyanString("Installed Plugins:"))
		for _, d := range dirs {
			if d.IsDir() || (d.Type()&os.ModeSymlink != 0) {
				manifest, err := plugin.GetManifest(d.Name())
				if err != nil {
					fmt.Printf("  %s (error reading manifest)\n", d.Name())
					continue
				}
				fmt.Printf("  %s (%s) - %s\n", color.GreenString(manifest.Name), manifest.Version, manifest.Description)
			}
		}
	},
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if err := plugin.RemovePlugin(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Plugin '%s' removed.\n", name)
	},
}

var pluginUpdateCmd = &cobra.Command{
	Use:   "update <name>|--all",
	Short: "Update a plugin or all plugins",
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			pluginDir := plugin.GetPluginDir()
			dirs, err := os.ReadDir(pluginDir)
			if err != nil {
				return
			}
			for _, d := range dirs {
				if d.IsDir() {
					fmt.Printf("Updating %s...\n", d.Name())
					plugin.UpdatePlugin(d.Name())
				}
			}
			return
		}

		if len(args) == 0 {
			cmd.Help()
			return
		}

		name := args[0]
		if err := plugin.UpdatePlugin(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Plugin '%s' updated.\n", name)
	},
}

func init() {
	pluginCmd.AddCommand(pluginAddCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	pluginCmd.AddCommand(pluginUpdateCmd)
	pluginUpdateCmd.Flags().BoolP("all", "a", false, "Update all plugins")
}
