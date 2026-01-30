package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for k2l.

To load completions:

Bash:
  $ source <(k2l completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ k2l completion bash > /etc/bash_completion.d/k2l
  # macOS:
  $ k2l completion bash > $(brew --prefix)/etc/bash_completion.d/k2l

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ k2l completion zsh > "${fpath[1]}/_k2l"

  # Or add to your .zshrc:
  $ source <(k2l completion zsh)

Fish:
  $ k2l completion fish | source
  # To load completions for each session, execute once:
  $ k2l completion fish > ~/.config/fish/completions/k2l.fish

PowerShell:
  PS> k2l completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> k2l completion powershell > k2l.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
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

