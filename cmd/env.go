package cmd

import (
	"avm/internal/config"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

var envFormat string

var envCmd = &cobra.Command{
	Use:    "env",
	Short:  "Resolve environment variables for current directory",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		env, err := config.ResolveEnv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "avm env error: %v\n", err)
			os.Exit(1)
		}

		if envFormat != "export" {
			fmt.Fprintln(os.Stderr, "unknown env format")
			os.Exit(1)
		}

		var keys []string
		for key := range env {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			fmt.Fprintf(os.Stdout, "export %s=%s\n", key, shellQuote(env[key]))
		}
	},
}

func init() {
	envCmd.Flags().StringVarP(&envFormat, "format", "f", "export", "Output format")
	rootCmd.AddCommand(envCmd)
}

