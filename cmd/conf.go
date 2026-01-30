package cmd

import (
	"fmt"
	"os"

	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/spf13/cobra"
)

var confCmd = &cobra.Command{
	Use:   "conf <context> <namespace> <configmap>",
	Short: "Show configmap YAML",
	Long: `Show the configmap in YAML format.

Arguments support substring matching:
  k2l conf prod myapp api-server
  k2l conf prod my api
  k2l conf prod w-f admin  # w-f matches web-frontend`,
	Args:              cobra.MinimumNArgs(3),
	ValidArgsFunction: makeConfigMapCompleter(),
	Run: func(cmd *cobra.Command, args []string) {
		context, namespace, configmap, err := parseArgs(args)
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

		// Resolve configmap (with selector on ambiguity)
		client := kubectl.NewClient(actualContext, actualNamespace)
		actualConfigMap, err := resolveConfigMap(client, configmap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Save to history
		saveToHistory("conf", args)

		// Get configmap
		yaml, err := client.GetConfigMapByName(actualConfigMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(yaml)
	},
}
