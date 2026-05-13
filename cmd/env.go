package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/PrajaNova/avm/internal/tooling"
	"github.com/spf13/cobra"
)

var envFormat string

var envCmd = &cobra.Command{
	Use:    "env",
	Short:  "Resolve environment variables for current directory",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if envFormat != "export" {
			return fmt.Errorf("unknown env format: %s", envFormat)
		}

		env, err := config.ResolveEnv()
		if err != nil {
			return fmt.Errorf("avm env error: %w", err)
		}

		toolEnv, err := tooling.ResolveToolEnv(".")
		if err != nil {
			return err
		}
		for key, value := range toolEnv {
			env[key] = value
		}

		var keys []string
		for key := range env {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Fprintf(os.Stdout, "export %s=%s\n", key, shellQuote(env[key]))
		}
		return nil
	},
}

func init() {
	envCmd.Flags().StringVarP(&envFormat, "format", "f", "export", "Output format")
	rootCmd.AddCommand(envCmd)
}
