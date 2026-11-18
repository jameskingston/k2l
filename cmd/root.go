package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jameskingston/k2l/internal/history"
	"github.com/jameskingston/k2l/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "k2l [command] [context] [namespace] [resource]",
	Short: "A kubectl wrapper for common operations",
	Long: `k2l is a simplified kubectl wrapper that provides quick access to common operations.

Commands:
  dep   - Show deployment YAML
  conf  - Show configmap YAML
  shell - Open shell in a pod
  pods  - List pods for a deployment
  update-image - Update the image tag for a deployment

Examples:
  k2l dep prod myapp api-server    # Show deployment YAML
  k2l conf prod myapp api-server   # Show configmap YAML
  k2l shell prod myapp api-server  # Open shell in deployment pod
  k2l pods prod myapp api-server   # List pods
  k2l update-image dev myapp api-server branch_1  # Set image tag to branch_1

The kubectl command that performs the final action is printed to stderr.

Identifiers support substring matching:
  k2l conf prod my api             # Matches myapp, api-server
  k2l conf prod w-f admin          # w-f matches web-frontend

Run without arguments to select from history.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Set the Run function here to avoid initialization cycle
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		// If no args, show history selector
		if len(args) == 0 {
			runHistorySelector()
			return
		}
	}

	// Add subcommands
	rootCmd.AddCommand(depCmd)
	rootCmd.AddCommand(confCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(podsCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(historyCmd)
}

func runHistorySelector() {
	hist, err := history.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading history: %v\n", err)
		os.Exit(1)
	}

	entries := hist.GetDisplayStrings()
	if len(entries) == 0 {
		fmt.Println("No command history. Run a command first!")
		fmt.Println("\nUsage: k2l <command> <context> <namespace> <resource>")
		fmt.Println("\nCommands: dep, conf, shell, pods, update-image")
		return
	}

	result, err := ui.Select("Select a previous command:", entries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if result == nil {
		// User cancelled
		return
	}

	// Parse the selected command and execute it
	entry := hist.GetEntry(result.Index)
	if entry == nil {
		return
	}

	// Re-execute the command
	executeCommand(entry.Command, entry.Args)
}

func executeCommand(command string, args []string) {
	// Build new args for rootCmd
	newArgs := append([]string{command}, args...)

	// Reset command for re-execution
	rootCmd.SetArgs(newArgs)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// saveToHistory saves the command to history
func saveToHistory(command string, args []string) {
	hist, err := history.New()
	if err != nil {
		return // Silently fail on history save errors
	}
	hist.Add(command, args)
}

// parseArgs parses positional arguments for context, namespace, and resource
func parseArgs(args []string) (context, namespace, resource string, err error) {
	if len(args) < 3 {
		return "", "", "", fmt.Errorf("requires context, namespace, and resource arguments")
	}

	return args[0], args[1], strings.Join(args[2:], "-"), nil
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage command history",
	Long:  "View and manage k2l command history",
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear command history",
	Run: func(cmd *cobra.Command, args []string) {
		hist, err := history.New()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := hist.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing history: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("History cleared")
	},
}

func init() {
	historyCmd.AddCommand(historyClearCmd)
}
