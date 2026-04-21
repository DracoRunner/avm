package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const shellInitScript = `# avm shell function
# Enables directory-aware alias resolution
avm() {
  local key="$1"

  # No arguments - show help
  if [ -z "$key" ]; then
    command avm-bin "$@"
    return $?
  fi

  # Check if it's an avm subcommand
  case "$key" in
    init|add|list|ls|remove|rm|which|version|help|shell-init|completion|--help|-h|--version|-v)
      command avm-bin "$@"
      return $?
      ;;
  esac

  # Try to resolve the alias
  local resolved
  resolved=$(command avm-bin which "$key" 2>/dev/null | grep "^Command:" | sed 's/^Command: //')

  if [ -n "$resolved" ]; then
    shift
    # Check if command contains placeholders ($1, $2, etc.)
    if echo "$resolved" | grep -qE '\$[0-9]+'; then
      # Substitute placeholders with provided arguments
      local cmd="$resolved"
      local i=1
      for arg in "$@"; do
        cmd=$(echo "$cmd" | sed "s|\\\$$i|$arg|g")
        i=$((i + 1))
      done
      # Remove any remaining unsubstituted placeholders
      cmd=$(echo "$cmd" | sed 's/\$[0-9]\+//g')
      eval "$cmd"
      return $?
    else
      # No placeholders - append args at the end (original behavior)
      eval "$resolved $@"
      return $?
    fi
  fi

  # No alias found, show usage
  command avm-bin "$@"
  return $?
}
`

var shellInitCmd = &cobra.Command{
	Use:   "shell-init",
	Short: "Output shell function for bash/zsh integration",
	Long: `Output a shell function that enables directory-aware alias resolution.

Add this to your ~/.zshrc or ~/.bashrc:

  eval "$(avm shell-init)"

Then reload your shell:

  source ~/.zshrc  # or source ~/.bashrc

Note: The binary is installed as 'avm-bin' and the shell function 'avm' wraps it.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(shellInitScript)
	},
}
