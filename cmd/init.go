package cmd

import (
	"fmt"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Create a local .avm.json file",
	Long:    `Creates a .avm.json file in the current directory for local aliases.`,
	Example: `  avm init`,
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := &config.Alias{
			Root:      ".",
			LocalFile: ".avm.json",
			Global:    false,
		}

		if err := config.Init(alias); err != nil {
			return err
		}

		fmt.Println("✓ Created .avm.json in current directory")
		return nil
	},
}
