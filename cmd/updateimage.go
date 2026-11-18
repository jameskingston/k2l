package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/jameskingston/k2l/internal/ui"
	"github.com/spf13/cobra"
)

// Image tags may contain letters, digits, '_', '.', '-', must not start with '.' or '-', max 128 chars
var validTag = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)

var updateImageCmd = &cobra.Command{
	Use:   "update-image <context> <namespace> <deployment> <tag>",
	Short: "Update the image tag for a deployment",
	Long: `Replace the image tag of a deployment's container, keeping the same repository.

For example, an image of
  277635488776.dkr.ecr.us-east-1.amazonaws.com/live-portal/backend:v1.19.6
becomes
  277635488776.dkr.ecr.us-east-1.amazonaws.com/live-portal/backend:branch_1

If the deployment has multiple containers, you'll be prompted to select one.

Arguments support substring matching:
  k2l update-image pt-eks-dev live-portal-dev live-portal-backend branch_1
  k2l update-image dev live-portal backend branch_1`,
	Args:              cobra.MinimumNArgs(4),
	ValidArgsFunction: makeDeploymentCompleter(),
	Run: func(cmd *cobra.Command, args []string) {
		tag := args[len(args)-1]
		context, namespace, deployment, err := parseArgs(args[:len(args)-1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if !validTag.MatchString(tag) {
			fmt.Fprintf(os.Stderr, "Error: '%s' is not a valid image tag (allowed: letters, digits, '_', '.', '-'; max 128 chars)\n", tag)
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

		containers, err := client.GetDeploymentContainers(actualDeployment)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(containers) == 0 {
			fmt.Fprintf(os.Stderr, "No containers found in deployment '%s'\n", actualDeployment)
			os.Exit(1)
		}

		container := containers[0]

		// If multiple containers, let user select
		if len(containers) > 1 {
			names := make([]string, len(containers))
			for i, c := range containers {
				names[i] = fmt.Sprintf("%s (%s)", c.Name, c.Image)
			}

			result, err := ui.Select("Select container:", names)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if result == nil {
				return // User cancelled
			}

			container = containers[result.Index]
		}

		newImage := kubectl.ReplaceImageTag(container.Image, tag)
		if newImage == container.Image {
			fmt.Printf("Container '%s' is already using %s\n", container.Name, newImage)
			return
		}

		// Save to history
		saveToHistory("update-image", args)

		fmt.Printf("Container: %s\n  from: %s\n    to: %s\n", container.Name, container.Image, newImage)

		if err := client.SetImage(actualDeployment, container.Name, newImage); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateImageCmd)
}
