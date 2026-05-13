package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/PrajaNova/avm/internal/config"
	"github.com/spf13/cobra"
)

var globalAlias bool

var addCmd = &cobra.Command{
	Use:   "add [key] [value]",
	Short: "Add or update an alias",
	Long:  `Add or update an alias in either local or global configuration.`,
	Example: `  avm add start "npm run dev"
  avm add deploy "sh ./scripts/deploy.sh"
  avm add -g cleanup "docker system prune -a"`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires at least 2 arguments: key and value")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := strings.Join(args[1:], " ")

		var alias *config.Alias
		if globalAlias {
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

		if err := config.Add(alias, key, value); err != nil {
			return err
		}

		if globalAlias {
			fmt.Printf("✓ Added global alias '%s' → %s\n", key, value)
		} else {
			fmt.Printf("✓ Added local alias '%s' → %s\n", key, value)
		}
		return nil
	},
}

func init() {
	addCmd.Flags().BoolVarP(&globalAlias, "global", "g", false, "Use global configuration")
}
