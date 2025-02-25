package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	completionCmd := newCompletionCmd()
	rootCmd.AddCommand(completionCmd)
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(git-branch-delete completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ git-branch-delete completion bash > /etc/bash_completion.d/git-branch-delete
  # macOS:
  $ git-branch-delete completion bash > $(brew --prefix)/etc/bash_completion.d/git-branch-delete

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ git-branch-delete completion zsh > "${fpath[1]}/_git-branch-delete"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ git-branch-delete completion fish | source

  # To load completions for each session, execute once:
  $ git-branch-delete completion fish > ~/.config/fish/completions/git-branch-delete.fish

PowerShell:
  PS> git-branch-delete completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> git-branch-delete completion powershell > git-branch-delete.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCompletion(cmd, args); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		if err := cmd.Root().GenBashCompletion(os.Stdout); err != nil {
			return fmt.Errorf("failed to generate bash completion: %w", err)
		}
	case "zsh":
		if err := cmd.Root().GenZshCompletion(os.Stdout); err != nil {
			return fmt.Errorf("failed to generate zsh completion: %w", err)
		}
	case "fish":
		if err := cmd.Root().GenFishCompletion(os.Stdout, true); err != nil {
			return fmt.Errorf("failed to generate fish completion: %w", err)
		}
	case "powershell":
		if err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout); err != nil {
			return fmt.Errorf("failed to generate powershell completion: %w", err)
		}
	default:
		return fmt.Errorf("invalid shell type %q", args[0])
	}
	return nil
}
