package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/spf13/cobra"
)

var placeholderRegex = regexp.MustCompile(`\$[0-9]+`)

var resolveCmd = &cobra.Command{
	Use:    "resolve <key> [args...]",
	Short:  "Resolve an alias, expand placeholders, and print only the command",
	Hidden: true,
	Args:   cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return exitCode(1)
		}
		key := args[0]
		restArgs := args[1:]

		val, found, _, err := config.ResolveWithSource(key)
		if err != nil {
			return fmt.Errorf("avm resolve error: %w", err)
		}

		if !found {
			return exitCode(1)
		}

		cmdline := expandPlaceholders(val, restArgs)
		fmt.Print(cmdline)
		return nil
	},
}

// expandPlaceholders replaces $1..$N with args, leaves others empty.
func expandPlaceholders(template string, args []string) string {
	if !strings.ContainsRune(template, '$') {
		if len(args) == 0 {
			return template
		}
		return template + " " + shellQuoteAll(args)
	}
	out := template
	for i, a := range args {
		out = strings.ReplaceAll(out, fmt.Sprintf("$%d", i+1), shellQuote(a))
	}
	return placeholderRegex.ReplaceAllString(out, "")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellQuoteAll(args []string) string {
	var quoted []string
	for _, a := range args {
		quoted = append(quoted, shellQuote(a))
	}
	return strings.Join(quoted, " ")
}

func init() {
	rootCmd.AddCommand(resolveCmd)
}
