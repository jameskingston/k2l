package cmd

import (
	"fmt"
	"os"

	"github.com/jameskingston/k2l/internal/history"
	"github.com/spf13/cobra"
)

var againCmd = &cobra.Command{
	Use:     "again",
	Aliases: []string{"."},
	Short:   "Repeat the most recent command",
	Long:    `Repeats the most recently executed k2l command.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAgain()
	},
}

func init() {
	rootCmd.AddCommand(againCmd)
}

func runAgain() {
	hist, err := history.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	entry := hist.GetEntry(0)
	if entry == nil {
		fmt.Println("No command history. Run a command first!")
		return
	}

	// Show what we're repeating
	cmdStr := entry.Command
	for _, arg := range entry.Args {
		cmdStr += " " + arg
	}
	fmt.Printf("Repeating: %s\n\n", cmdStr)

	// Re-execute the command
	executeCommand(entry.Command, entry.Args)
}
