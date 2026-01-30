package cmd

import (
	"fmt"
	"os"

	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/spf13/cobra"
)

var depCmd = &cobra.Command{
	Use:   "dep <context> <namespace> <deployment>",
	Short: "Show deployment YAML",
	Long: `Show the deployment definition in YAML format.

Arguments support substring matching:
  k2l dep prod myapp api-server
  k2l dep prod my api
  k2l dep prod w-f admin  # w-f matches web-frontend`,
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
		saveToHistory("dep", args)

		// Get deployment
		yaml, err := client.GetDeploymentByName(actualDeployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(yaml)
	},
}
