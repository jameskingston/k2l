package cmd

import (
	"fmt"
	"os"

	"github.com/jameskingston/k2l/internal/kubectl"
	"github.com/jameskingston/k2l/internal/ui"
)

// resolveContext resolves a context query, showing a selector if ambiguous
func resolveContext(query string) (string, error) {
	result := kubectl.ResolveContextWithMatches(query)

	if !result.Found {
		return "", fmt.Errorf("no context matching '%s' found", query)
	}

	if result.Ambiguous {
		fmt.Fprintf(os.Stderr, "Multiple contexts match '%s':\n", query)
		selected, err := ui.Select("Select context:", result.Matches)
		if err != nil {
			return "", err
		}
		if selected == nil {
			return "", fmt.Errorf("cancelled")
		}
		return selected.Value, nil
	}

	return result.Value, nil
}

// resolveNamespace resolves a namespace query, showing a selector if ambiguous
func resolveNamespace(context, query string) (string, error) {
	result := kubectl.ResolveNamespaceWithMatches(context, query)

	if !result.Found {
		return "", fmt.Errorf("no namespace matching '%s' found", query)
	}

	if result.Ambiguous {
		fmt.Fprintf(os.Stderr, "Multiple namespaces match '%s':\n", query)
		selected, err := ui.Select("Select namespace:", result.Matches)
		if err != nil {
			return "", err
		}
		if selected == nil {
			return "", fmt.Errorf("cancelled")
		}
		return selected.Value, nil
	}

	return result.Value, nil
}

// resolveDeployment resolves a deployment query, showing a selector if ambiguous
func resolveDeployment(client *kubectl.Client, query string) (string, error) {
	result := client.ResolveDeploymentWithMatches(query)

	if !result.Found {
		return "", fmt.Errorf("no deployment matching '%s' found", query)
	}

	if result.Ambiguous {
		fmt.Fprintf(os.Stderr, "Multiple deployments match '%s':\n", query)
		selected, err := ui.Select("Select deployment:", result.Matches)
		if err != nil {
			return "", err
		}
		if selected == nil {
			return "", fmt.Errorf("cancelled")
		}
		return selected.Value, nil
	}

	return result.Value, nil
}

// resolveConfigMap resolves a configmap query, showing a selector if ambiguous
func resolveConfigMap(client *kubectl.Client, query string) (string, error) {
	result := client.ResolveConfigMapWithMatches(query)

	if !result.Found {
		return "", fmt.Errorf("no configmap matching '%s' found", query)
	}

	if result.Ambiguous {
		fmt.Fprintf(os.Stderr, "Multiple configmaps match '%s':\n", query)
		selected, err := ui.Select("Select configmap:", result.Matches)
		if err != nil {
			return "", err
		}
		if selected == nil {
			return "", fmt.Errorf("cancelled")
		}
		return selected.Value, nil
	}

	return result.Value, nil
}
