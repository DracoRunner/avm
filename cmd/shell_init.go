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
    init|add|list|ls|remove|rm|which|version|help|shell-init|plugin|completion|--help|-h|--version|-v)
      command avm-bin "$@"
      return $?
      ;;
  esac

  # Try to resolve the alias
  local resolved
  if resolved=$(command avm-bin resolve "$@" 2>/dev/null); then
    eval "$resolved"
    return $?
  fi

  # No alias found, show interactive suggestions
  # We use a temporary file to capture the selected command if exit code is 10
  local _avm_out_file
  _avm_out_file=$(mktemp)

  local _avm_ret_code

  # We run avm-bin directly so it stays connected to the terminal (TTY).
  # We pass a temp file path via environment variable for it to write the result.
  AVM_RESULT_FILE="$_avm_out_file" command avm-bin "$@"
  _avm_ret_code=$?

  if [ $_avm_ret_code -eq 10 ]; then
    # User picked a suggestion (which is just the command key)
    local suggestion
    suggestion=$(cat "$_avm_out_file")
    rm -f "$_avm_out_file"
    if [ -n "$suggestion" ]; then
      # Re-run avm with the picked suggestion
      shift
      avm "$suggestion" "$@"
      return $?
    fi
  elif [ $_avm_ret_code -eq 0 ]; then
    # User chose "None" or hit ESC, or no suggestions were found
    # Passthrough: run the original command
    rm -f "$_avm_out_file"
    # Execute the original command, properly quoting arguments
    "$@"
    return $?
  else
    # Some other error, return exit code
    rm -f "$_avm_out_file"
    return $_avm_ret_code
  fi
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
