package cmd

import (
	"fmt"
	"os"

	"avm/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a local .avm.json file",
	Long:  `Creates a .avm.json file in the current directory for local aliases.`,
	Example: `  avm init`,
	Run: func(cmd *cobra.Command, args []string) {
		alias := &config.Alias{
			Root:      ".",
			LocalFile: ".avm.json",
			Global:    false,
		}

		if err := config.Init(alias); err != nil {
			fmt.Fprintf(os.Stderr, "avm: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Created .avm.json in current directory")
	},
}
