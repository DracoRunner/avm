package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "avm",
	Short: "A lightweight local/global command alias manager",
	Long: `avm - Alias Version Manager

A lightweight local/global command alias manager that works like asdf or nvm.

Configuration:
  - Local config: .avm.json in current directory
  - Global config: ~/.avm.json in user home

If no alias is found, commands pass through to the shell normally.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "avm: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(whichCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(shellInitCmd)
}
