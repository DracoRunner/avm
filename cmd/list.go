package cmd

import (
	"fmt"
	"os"

	"avm/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active aliases and environment variables",
	Long:    `Display active aliases and environment overrides currently in scope for this directory, including local and global sources.`,
	Run: func(cmd *cobra.Command, args []string) {
		alias := &config.Alias{
			Root:      ".",
			LocalFile: ".avm.json",
			Global:    false,
		}

		if err := config.List(alias); err != nil {
			fmt.Fprintf(os.Stderr, "avm: %v\n", err)
			os.Exit(1)
		}
	},
}
