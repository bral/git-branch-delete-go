package cmd

import (
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
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}
