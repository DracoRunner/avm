package cmd

import (
	"avm/internal/config"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var placeholderRegex = regexp.MustCompile(`\$[0-9]+`)

var resolveCmd = &cobra.Command{
	Use:    "resolve <key> [args...]",
	Short:  "Resolve an alias, expand placeholders, and print only the command",
	Hidden: true,
	Args:   cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			os.Exit(1)
		}
		key := args[0]
		restArgs := args[1:]

		val, found, _, err := config.ResolveWithSource(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "avm resolve error: %v\n", err)
			os.Exit(1)
		}

		if !found {
			os.Exit(1)
		}

		cmdline := expandPlaceholders(val, restArgs)
		fmt.Print(cmdline)
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
