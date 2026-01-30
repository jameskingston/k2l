package cmd

import (
	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/jameskingston/k2l/internal/match"
	"github.com/spf13/cobra"
)

// completeContextArg completes kubectl contexts
func completeContextArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	contexts, err := kubectl.ListContexts()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if toComplete == "" {
		return contexts, cobra.ShellCompDirectiveNoFileComp
	}

	// Filter by prefix/substring match
	return match.FindAllMatches(toComplete, contexts), cobra.ShellCompDirectiveNoFileComp
}

// completeNamespaceArg completes namespaces for a given context
func completeNamespaceArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Resolve the context from first arg
	context, err := kubectl.ResolveContext(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client := kubectl.NewClient(context, "default")
	namespaces, err := client.ListNamespaces()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if toComplete == "" {
		return namespaces, cobra.ShellCompDirectiveNoFileComp
	}

	return match.FindAllMatches(toComplete, namespaces), cobra.ShellCompDirectiveNoFileComp
}

// completeDeploymentArg completes deployments for a given context and namespace
func completeDeploymentArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Resolve context and namespace
	context, err := kubectl.ResolveContext(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace, err := kubectl.ResolveNamespace(context, args[1])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client := kubectl.NewClient(context, namespace)
	deployments, err := client.ListDeployments()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if toComplete == "" {
		return deployments, cobra.ShellCompDirectiveNoFileComp
	}

	return match.FindAllMatches(toComplete, deployments), cobra.ShellCompDirectiveNoFileComp
}

// completeConfigMapArg completes configmaps for a given context and namespace
func completeConfigMapArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Resolve context and namespace
	context, err := kubectl.ResolveContext(args[0])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	namespace, err := kubectl.ResolveNamespace(context, args[1])
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client := kubectl.NewClient(context, namespace)
	configmaps, err := client.ListConfigMaps()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if toComplete == "" {
		return configmaps, cobra.ShellCompDirectiveNoFileComp
	}

	return match.FindAllMatches(toComplete, configmaps), cobra.ShellCompDirectiveNoFileComp
}

// makeDeploymentCompleter creates a ValidArgsFunction for commands that take context, namespace, deployment
func makeDeploymentCompleter() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completeContextArg(cmd, args, toComplete)
		case 1:
			return completeNamespaceArg(cmd, args, toComplete)
		case 2:
			return completeDeploymentArg(cmd, args, toComplete)
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}

// makeConfigMapCompleter creates a ValidArgsFunction for commands that take context, namespace, configmap
func makeConfigMapCompleter() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		switch len(args) {
		case 0:
			return completeContextArg(cmd, args, toComplete)
		case 1:
			return completeNamespaceArg(cmd, args, toComplete)
		case 2:
			return completeConfigMapArg(cmd, args, toComplete)
		default:
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}
