package cmd

import (
	"fmt"
	"os"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
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
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			query := args[0]

			// Get subcommand suggestions from Cobra
			subCommandSuggestions := cmd.SuggestionsFor(query)

			// Get alias suggestions
			aliasSuggestions := config.SuggestAliases(query)

			allSuggestions := append(subCommandSuggestions, aliasSuggestions...)

			if len(allSuggestions) > 0 {
				fmt.Fprintf(os.Stderr, "avm: unknown command or alias \"%s\"\n", color.RedString(query))

				options := append(allSuggestions, "None (run as-is)")
				prompt := promptui.Select{
					Label: "Did you mean one of these?",
					Items: options,
					Templates: &promptui.SelectTemplates{
						Active:   `▸ {{ . | cyan }}`,
						Inactive: `  {{ . }}`,
						Selected: `✔ Selected {{ . | green }}`,
					},
					HideSelected: true,
					Stdin:        os.Stdin,
					Stdout:       os.Stderr,
				}

				_, result, err := prompt.Run()

				if err != nil {
					// User hit ESC or interrupted
					return nil
				}

				if result == "None (run as-is)" {
					return nil
				}

				// If AVM_RESULT_FILE is set, write the result there for the shell function
				resultFile := os.Getenv("AVM_RESULT_FILE")
				if resultFile != "" {
					if err := os.WriteFile(resultFile, []byte(result), 0644); err != nil {
						return err
					}
				} else {
					// Fallback to stdout
					fmt.Println(result)
				}

				return exitCode(10)
			} else {
				// No suggestions, just exit 0 to let shell handle passthrough
				return nil
			}
		} else {
			return cmd.Help()
		}
		return nil
	},
}

type exitCodeError struct {
	code int
}

func (e exitCodeError) Error() string {
	return ""
}

func exitCode(code int) error {
	return exitCodeError{code: code}
}

func Execute() {
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		if exitErr, ok := err.(exitCodeError); ok {
			os.Exit(exitErr.code)
		}
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
	rootCmd.AddCommand(pluginCmd)
}
