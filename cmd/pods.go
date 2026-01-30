package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

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
		pods, err := client.GetPodsByDeploymentName(actualDeployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(pods) == 0 {
			fmt.Printf("No pods found for deployment '%s'\n", deployment)
			return
		}

		// Print with nice formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Header
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			headerStyle.Render("NAME"),
			headerStyle.Render("STATUS"),
			headerStyle.Render("READY"),
			headerStyle.Render("RESTARTS"),
			headerStyle.Render("AGE"),
		)

		for _, pod := range pods {
			statusStr := styleStatus(pod.Status)
			ageStr := ageStyle.Render(pod.AgeString)

			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n",
				pod.Name,
				statusStr,
				pod.Ready,
				pod.Restarts,
				ageStr,
			)
		}

		w.Flush()
	},
}

func styleStatus(status string) string {
	switch status {
	case "Running":
		return runningStyle.Render(status)
	case "Pending", "ContainerCreating":
		return pendingStyle.Render(status)
	case "Failed", "Error", "CrashLoopBackOff":
		return failedStyle.Render(status)
	default:
		return status
	}
}
