package cmd

import (
	"fmt"
	"os"

	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/jameskingston/k2l/internal/ui"
	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell <context> <namespace> <deployment>",
	Short: "Open shell in a deployment pod",
	Long: `Open an interactive shell in a pod from the specified deployment.

If multiple pods exist, you'll be prompted to select one.

Arguments support substring matching:
  k2l shell prod myapp api-server
  k2l shell prod my api`,
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
		saveToHistory("shell", args)

		// Get pods
		pods, err := client.GetPodsByDeploymentName(actualDeployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(pods) == 0 {
			fmt.Fprintf(os.Stderr, "No pods found for deployment '%s'\n", deployment)
			os.Exit(1)
		}

		podIndex := 0

		// If multiple pods, let user select
		if len(pods) > 1 {
			podNames := make([]string, len(pods))
			for i, p := range pods {
				podNames[i] = fmt.Sprintf("%s (%s, %s)", p.Name, p.Status, p.AgeString)
			}

			result, err := ui.SelectPod(podNames)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if result == nil {
				return // User cancelled
			}

			podIndex = result.Index
		} else {
			fmt.Printf("Connecting to %s...\n", pods[0].Name)
		}

		// Exec into pod
		if err := client.ExecShellInPod(pods[podIndex].Name); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}
