package cmd

import (
	"avm/internal/config"
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
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
					os.Exit(0)
				}

				if result == "None (run as-is)" {
					os.Exit(0)
				}

				// If AVM_RESULT_FILE is set, write the result there for the shell function
				resultFile := os.Getenv("AVM_RESULT_FILE")
				if resultFile != "" {
					if err := os.WriteFile(resultFile, []byte(result), 0644); err != nil {
						os.Exit(1)
					}
				} else {
					// Fallback to stdout
					fmt.Println(result)
				}

				os.Exit(10)
			} else {
				// No suggestions, just exit 0 to let shell handle passthrough
				os.Exit(0)
			}
		} else {
			cmd.Help()
		}
	},
}

func Execute() {
	rootCmd.SilenceErrors = true
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
	rootCmd.AddCommand(pluginCmd)
}
