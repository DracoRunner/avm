package cmd

import (
	"github.com/PrajaNova/avm/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active aliases and environment variables",
	Long:    `Display active aliases and environment overrides currently in scope for this directory, including local and global sources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		alias := &config.Alias{
			Root:      ".",
			LocalFile: ".avm.json",
			Global:    false,
		}

		return config.List(alias)
	},
}
