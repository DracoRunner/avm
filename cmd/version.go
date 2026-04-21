package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the CLI version",
	Long:  `Display the current version of the avm CLI tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("avm version %s\n", version)
	},
}
