package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/PrajaNova/avm/internal/tooling"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var toolGlobal bool

var toolCmd = &cobra.Command{
	Use:     "tool",
	Short:   "Manage project runtimes",
	Long:    `Manage runtime versions for the current directory or globally.`,
	Aliases: []string{"tools"},
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured and installed tools",
	RunE: func(cmd *cobra.Command, args []string) error {
		resolved, err := tooling.ResolveTools(".")
		if err != nil {
			return fmt.Errorf("avm tool list error: %w", err)
		}

		knownTools := tooling.KnownTools()
		if len(knownTools) == 0 {
			fmt.Println("No tool providers are available.")
			return nil
		}

		fmt.Println(color.CyanString("Tool providers:"))
		for _, tool := range knownTools {
			fmt.Printf("  %s\n", color.GreenString(tool))

			if selection, ok := resolved[tool]; ok {
				fmt.Printf("    active (%s): %s\n", selection.Source, selection.Version)
			} else {
				fmt.Printf("    active: none\n")
			}

			installed, err := tooling.InstalledVersions(tool)
			if err != nil {
				fmt.Printf("    installed: unable to read\n")
				continue
			}
			if len(installed) == 0 {
				fmt.Printf("    installed: none\n")
				continue
			}
			sort.Strings(installed)
			fmt.Printf("    installed: %s\n", strings.Join(installed, ", "))
		}
		return nil
	},
}

var toolUseCmd = &cobra.Command{
	Use:   "use <tool> <version>",
	Short: "Set the active version for a tool",
	Long:  `Set a local or global tool version override in .avm.json.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]
		version := args[1]

		_, ok := tooling.GetProvider(tool)
		if !ok {
			return fmt.Errorf("unsupported tool %q", tool)
		}

		var alias *config.Alias
		if toolGlobal {
			alias = &config.Alias{
				Root:      os.Getenv("HOME"),
				LocalFile: ".avm.json",
				Global:    true,
			}
		} else {
			alias = &config.Alias{
				Root:      ".",
				LocalFile: ".avm.json",
				Global:    false,
			}
		}

		if err := config.UseTool(alias, tool, version); err != nil {
			return err
		}

		if toolGlobal {
			fmt.Printf("✓ Set global %s version to %s\n", tool, version)
		} else {
			fmt.Printf("✓ Set local %s version to %s\n", tool, version)
		}
		return nil
	},
}

var toolInstallCmd = &cobra.Command{
	Use:   "install <tool> <version>",
	Short: "Install a tool version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]
		version := args[1]

		_, ok := tooling.GetProvider(tool)
		if !ok {
			return fmt.Errorf("unsupported tool %q", tool)
		}

		if err := tooling.InstallTool(tool, version); err != nil {
			return err
		}
		fmt.Printf("✓ Installed %s %s\n", tool, version)
		return nil
	},
}

var toolUninstallCmd = &cobra.Command{
	Use:   "uninstall <tool> <version>",
	Short: "Remove a tool version",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tool := args[0]
		version := args[1]
		if _, ok := tooling.GetProvider(tool); !ok {
			return fmt.Errorf("unsupported tool %q", tool)
		}

		if err := tooling.UninstallTool(tool, version); err != nil {
			return err
		}
		fmt.Printf("✓ Removed %s %s\n", tool, version)
		return nil
	},
}

func init() {
	toolCmd.AddCommand(toolListCmd)
	toolCmd.AddCommand(toolUseCmd)
	toolCmd.AddCommand(toolInstallCmd)
	toolCmd.AddCommand(toolUninstallCmd)

	toolUseCmd.Flags().BoolVarP(&toolGlobal, "global", "g", false, "Set global default tool version")
	rootCmd.AddCommand(toolCmd)
}
