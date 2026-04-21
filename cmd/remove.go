package cmd

import (
	"fmt"
	"os"

	"avm/internal/config"
	"github.com/spf13/cobra"
)

var removeGlobal bool

var removeCmd = &cobra.Command{
	Use:     "remove [key]",
	Aliases: []string{"rm"},
	Short:   "Remove an alias",
	Long:    `Remove an alias from either local or global configuration.`,
	Example: `  avm remove oldalias
  avm remove -g oldalias`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 argument: key")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		var alias *config.Alias
		if removeGlobal {
			alias = &config.Alias{
				Root:      os.Getenv("HOME"),
				LocalFile: ".avm.json",
				Global:    true,
			}
		} else {
			alias = &config.Alias{
				Root:      ".",
				LocalFile: ".avm.json",
				Global:    false,
			}
		}

		if err := config.Remove(alias, key); err != nil {
			fmt.Fprintf(os.Stderr, "avm: %v\n", err)
			os.Exit(1)
		}

		if removeGlobal {
			fmt.Printf("✓ Removed global alias '%s'\n", key)
		} else {
			fmt.Printf("✓ Removed local alias '%s'\n", key)
		}
	},
}

func init() {
	removeCmd.Flags().BoolVarP(&removeGlobal, "global", "g", false, "Use global configuration")
}
