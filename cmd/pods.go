package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/spf13/cobra"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	runningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	pendingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	failedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	ageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

var podsCmd = &cobra.Command{
	Use:   "pods <context> <namespace> <deployment>",
	Short: "List pods for a deployment",
	Long: `List all pods for the specified deployment with human-friendly formatting.

Shows pod name, status, ready state, restarts, and age.

Arguments support substring matching:
  k2l pods prod myapp api-server
  k2l pods prod my api`,
	Args:              cobra.MinimumNArgs(3),
	ValidArgsFunction: makeDeploymentCompleter(),
	Run: func(cmd *cobra.Command, args []string) {
		context, namespace, deployment, err := parseArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Resolve context (with selector on ambiguity)
		actualContext, err := resolveContext(context)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Resolve namespace (with selector on ambiguity)
		actualNamespace, err := resolveNamespace(actualContext, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Resolve deployment (with selector on ambiguity)
		client := kubectl.NewClient(actualContext, actualNamespace)
		actualDeployment, err := resolveDeployment(client, deployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Save to history
		saveToHistory("pods", args)

		// Get pods
		pods, err := client.GetPodsByDeploymentNameShown(actualDeployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(pods) == 0 {
			fmt.Printf("No pods found for deployment '%s'\n", deployment)
			return
		}

		// Calculate column widths (based on actual text, not styled)
		nameW, statusW, readyW, restartsW, ageW := len("NAME"), len("STATUS"), len("READY"), len("RESTARTS"), len("AGE")
		for _, pod := range pods {
			if len(pod.Name) > nameW {
				nameW = len(pod.Name)
			}
			if len(pod.Status) > statusW {
				statusW = len(pod.Status)
			}
			if len(pod.Ready) > readyW {
				readyW = len(pod.Ready)
			}
			restartStr := fmt.Sprintf("%d", pod.Restarts)
			if len(restartStr) > restartsW {
				restartsW = len(restartStr)
			}
			if len(pod.AgeString) > ageW {
				ageW = len(pod.AgeString)
			}
		}

		// Print header with padding
		fmt.Printf("%s  %s  %s  %s  %s\n",
			headerStyle.Render(padRight("NAME", nameW)),
			headerStyle.Render(padRight("STATUS", statusW)),
			headerStyle.Render(padRight("READY", readyW)),
			headerStyle.Render(padRight("RESTARTS", restartsW)),
			headerStyle.Render("AGE"),
		)

		// Print rows with padding, then style
		for _, pod := range pods {
			restartStr := fmt.Sprintf("%d", pod.Restarts)
			fmt.Printf("%s  %s  %s  %s  %s\n",
				padRight(pod.Name, nameW),
				styleStatus(padRight(pod.Status, statusW)),
				padRight(pod.Ready, readyW),
				padRight(restartStr, restartsW),
				ageStyle.Render(pod.AgeString),
			)
		}
	},
}

func styleStatus(status string) string {
	switch {
	case strings.HasPrefix(status, "Running"):
		return runningStyle.Render(status)
	case strings.HasPrefix(status, "Pending"), strings.HasPrefix(status, "ContainerCreating"):
		return pendingStyle.Render(status)
	case strings.HasPrefix(status, "Failed"), strings.HasPrefix(status, "Error"), strings.HasPrefix(status, "CrashLoopBackOff"):
		return failedStyle.Render(status)
	default:
		return status
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
