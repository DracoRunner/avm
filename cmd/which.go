package cmd

import (
	"fmt"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/spf13/cobra"
)

var whichCmd = &cobra.Command{
	Use:   "which [command]",
	Short: "Show what a command resolves to",
	Long:  `Show what command a key resolves to and whether it comes from local or global configuration.`,
	Example: `  avm which start
  avm which deploy`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 argument: command key")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		alias := &config.Alias{
			Root:      ".",
			LocalFile: ".avm.json",
			Global:    false,
		}

		if err := config.Which(alias, key); err != nil {
			return err
		}

		if version, found, source, err := config.ResolveToolWithSource(key); err == nil && found {
			fmt.Println()
			fmt.Printf("Tool %s: %s (%s)\n", key, version, source)
		} else if err != nil {
			return err
		}
		return nil
	},
}
