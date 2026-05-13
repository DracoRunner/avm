package cmd

import (
	"fmt"
	"os"

	"github.com/PrajaNova/avm/internal/plugin"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		fmt.Printf("Installing plugin from %s...\n", source)
		if err := plugin.InstallPlugin(source); err != nil {
			return err
		}
		fmt.Println(color.GreenString("Plugin installed successfully."))
		return nil
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginDir := plugin.GetPluginDir()
		dirs, err := os.ReadDir(pluginDir)
		if err != nil {
			fmt.Println("No plugins installed.")
			return nil
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
		return nil
	},
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := plugin.RemovePlugin(name); err != nil {
			return err
		}
		fmt.Printf("Plugin '%s' removed.\n", name)
		return nil
	},
}

var pluginUpdateCmd = &cobra.Command{
	Use:   "update <name>|--all",
	Short: "Update a plugin or all plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			pluginDir := plugin.GetPluginDir()
			dirs, err := os.ReadDir(pluginDir)
			if err != nil {
				return nil
			}
			for _, d := range dirs {
				if d.IsDir() {
					fmt.Printf("Updating %s...\n", d.Name())
					if err := plugin.UpdatePlugin(d.Name()); err != nil {
						return err
					}
				}
			}
			return nil
		}

		if len(args) == 0 {
			return cmd.Help()
		}

		name := args[0]
		if err := plugin.UpdatePlugin(name); err != nil {
			return err
		}
		fmt.Printf("Plugin '%s' updated.\n", name)
		return nil
	},
}

func init() {
	pluginCmd.AddCommand(pluginAddCmd)
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	pluginCmd.AddCommand(pluginUpdateCmd)
	pluginUpdateCmd.Flags().BoolP("all", "a", false, "Update all plugins")
}
