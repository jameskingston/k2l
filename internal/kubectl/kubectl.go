package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jameskingston/k2l/internal/match"
)

// Client wraps kubectl operations
type Client struct {
	context   string
	namespace string
}

// NewClient creates a new kubectl client
func NewClient(context, namespace string) *Client {
	return &Client{
		context:   context,
		namespace: namespace,
	}
}

// run executes a kubectl command and returns the output
func (c *Client) run(args ...string) ([]byte, error) {
	fullArgs := []string{"--context", c.context, "-n", c.namespace}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command("kubectl", fullArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl error: %s\n%s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// runInteractive runs a kubectl command with terminal attached
func (c *Client) runInteractive(args ...string) error {
	fullArgs := []string{"--context", c.context, "-n", c.namespace}
	fullArgs = append(fullArgs, args...)

	cmd := exec.Command("kubectl", fullArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetDeployment returns the deployment definition in YAML (resolves name internally)
func (c *Client) GetDeployment(name string) (string, error) {
	// First, find the actual deployment name
	deployments, err := c.ListDeployments()
	if err != nil {
		return "", err
	}

	actualName, found := match.FindBestMatch(name, deployments)
	if !found {
		return "", fmt.Errorf("no deployment matching '%s' found", name)
	}

	return c.GetDeploymentByName(actualName)
}

// GetDeploymentByName returns the deployment definition in YAML (exact name)
func (c *Client) GetDeploymentByName(name string) (string, error) {
	output, err := c.run("get", "deployment", name, "-o", "yaml")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// GetConfigMap returns the configmap in YAML (resolves name internally)
func (c *Client) GetConfigMap(name string) (string, error) {
	// First, find the actual configmap name
	configmaps, err := c.ListConfigMaps()
	if err != nil {
		return "", err
	}

	actualName, found := match.FindBestMatch(name, configmaps)
	if !found {
		return "", fmt.Errorf("no configmap matching '%s' found", name)
	}

	return c.GetConfigMapByName(actualName)
}

// GetConfigMapByName returns the configmap in YAML (exact name)
func (c *Client) GetConfigMapByName(name string) (string, error) {
	output, err := c.run("get", "configmap", name, "-o", "yaml")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// Pod represents a kubernetes pod
type Pod struct {
	Name      string
	Status    string
	Ready     string
	Restarts  int
	Age       time.Duration
	AgeString string
}

// GetPods returns pods for a deployment with human-friendly formatting (resolves name internally)
func (c *Client) GetPods(deploymentName string) ([]Pod, error) {
	// Find the actual deployment name first
	deployments, err := c.ListDeployments()
	if err != nil {
		return nil, err
	}

	actualName, found := match.FindBestMatch(deploymentName, deployments)
	if !found {
		return nil, fmt.Errorf("no deployment matching '%s' found", deploymentName)
	}

	return c.GetPodsByDeploymentName(actualName)
}

// GetPodsByDeploymentName returns pods for a deployment (exact name)
func (c *Client) GetPodsByDeploymentName(deploymentName string) ([]Pod, error) {
	// Get pods with label selector matching the deployment
	output, err := c.run("get", "pods", "-l", fmt.Sprintf("app=%s", deploymentName), "-o", "json")
	if err != nil {
		// Try without label selector and filter by name
		output, err = c.run("get", "pods", "-o", "json")
		if err != nil {
			return nil, err
		}
	}

	var podList struct {
		Items []struct {
			Metadata struct {
				Name              string    `json:"name"`
				CreationTimestamp time.Time `json:"creationTimestamp"`
			} `json:"metadata"`
			Status struct {
				Phase             string `json:"phase"`
				ContainerStatuses []struct {
					Ready        bool `json:"ready"`
					RestartCount int  `json:"restartCount"`
				} `json:"containerStatuses"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(output, &podList); err != nil {
		return nil, fmt.Errorf("failed to parse pods: %w", err)
	}

	var pods []Pod
	for _, item := range podList.Items {
		// Filter pods by deployment name
		if !strings.Contains(item.Metadata.Name, deploymentName) {
			continue
		}

		age := time.Since(item.Metadata.CreationTimestamp)
		ready := 0
		total := len(item.Status.ContainerStatuses)
		restarts := 0
		for _, cs := range item.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
			restarts += cs.RestartCount
		}

		pods = append(pods, Pod{
			Name:      item.Metadata.Name,
			Status:    item.Status.Phase,
			Ready:     fmt.Sprintf("%d/%d", ready, total),
			Restarts:  restarts,
			Age:       age,
			AgeString: formatAge(age),
		})
	}

	return pods, nil
}

// ExecShell opens a shell in a pod from the deployment
func (c *Client) ExecShell(deploymentName string, podIndex int) error {
	pods, err := c.GetPods(deploymentName)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found for deployment '%s'", deploymentName)
	}

	if podIndex < 0 || podIndex >= len(pods) {
		podIndex = 0
	}

	return c.ExecShellInPod(pods[podIndex].Name)
}

// ExecShellInPod opens a shell in a specific pod by name
func (c *Client) ExecShellInPod(podName string) error {
	// Try common shells in order of preference
	shells := []string{"/bin/bash", "/bin/sh", "sh"}
	for _, shell := range shells {
		err := c.runInteractive("exec", "-it", podName, "--", shell)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("failed to exec into pod %s", podName)
}

// ListDeployments returns all deployment names in the namespace
func (c *Client) ListDeployments() ([]string, error) {
	output, err := c.run("get", "deployments", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(output))
	return names, nil
}

// ListConfigMaps returns all configmap names in the namespace
func (c *Client) ListConfigMaps() ([]string, error) {
	output, err := c.run("get", "configmaps", "-o", "jsonpath={.items[*].metadata.name}")
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(output))
	return names, nil
}

// ListNamespaces returns all namespace names
func (c *Client) ListNamespaces() ([]string, error) {
	cmd := exec.Command("kubectl", "--context", c.context, "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(output))
	return names, nil
}

// ListContexts returns all available contexts
func ListContexts() ([]string, error) {
	cmd := exec.Command("kubectl", "config", "get-contexts", "-o", "name")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(output))
	return names, nil
}

// ResolveResult contains the result of a resolution
type ResolveResult struct {
	Value     string
	Matches   []string
	Found     bool
	Ambiguous bool
}

// ResolveContext finds the actual context name from a query
func ResolveContext(query string) (string, error) {
	result := ResolveContextWithMatches(query)
	if !result.Found {
		return "", fmt.Errorf("no context matching '%s' found", query)
	}
	return result.Value, nil
}

// ResolveContextWithMatches returns context resolution with ambiguity info
func ResolveContextWithMatches(query string) ResolveResult {
	contexts, err := ListContexts()
	if err != nil {
		return ResolveResult{Found: false}
	}

	result := match.FindMatches(query, contexts)
	return ResolveResult{
		Value:     result.Match,
		Matches:   result.Matches,
		Found:     result.Found,
		Ambiguous: result.Ambiguous,
	}
}

// ResolveNamespace finds the actual namespace name from a query
func ResolveNamespace(context, query string) (string, error) {
	result := ResolveNamespaceWithMatches(context, query)
	if !result.Found {
		return "", fmt.Errorf("no namespace matching '%s' found", query)
	}
	return result.Value, nil
}

// ResolveNamespaceWithMatches returns namespace resolution with ambiguity info
func ResolveNamespaceWithMatches(context, query string) ResolveResult {
	client := NewClient(context, "default")
	namespaces, err := client.ListNamespaces()
	if err != nil {
		return ResolveResult{Found: false}
	}

	result := match.FindMatches(query, namespaces)
	return ResolveResult{
		Value:     result.Match,
		Matches:   result.Matches,
		Found:     result.Found,
		Ambiguous: result.Ambiguous,
	}
}

// ResolveDeploymentWithMatches returns deployment resolution with ambiguity info
func (c *Client) ResolveDeploymentWithMatches(query string) ResolveResult {
	deployments, err := c.ListDeployments()
	if err != nil {
		return ResolveResult{Found: false}
	}

	result := match.FindMatches(query, deployments)
	return ResolveResult{
		Value:     result.Match,
		Matches:   result.Matches,
		Found:     result.Found,
		Ambiguous: result.Ambiguous,
	}
}

// ResolveConfigMapWithMatches returns configmap resolution with ambiguity info
func (c *Client) ResolveConfigMapWithMatches(query string) ResolveResult {
	configmaps, err := c.ListConfigMaps()
	if err != nil {
		return ResolveResult{Found: false}
	}

	result := match.FindMatches(query, configmaps)
	return ResolveResult{
		Value:     result.Match,
		Matches:   result.Matches,
		Found:     result.Found,
		Ambiguous: result.Ambiguous,
	}
}

// formatAge formats a duration in a human-friendly way
func formatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}
