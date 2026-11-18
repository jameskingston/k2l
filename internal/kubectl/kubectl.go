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

func (c *Client) fullArgs(args []string) []string {
	return append([]string{"--context", c.context, "-n", c.namespace}, args...)
}

// run executes a kubectl command and returns the output
func (c *Client) run(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", c.fullArgs(args)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl error: %s\n%s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// runShown prints the command to stderr, then executes it like run.
// Stderr keeps stdout clean for piping (e.g. YAML output).
func (c *Client) runShown(args ...string) ([]byte, error) {
	c.printCommand(args)
	return c.run(args...)
}

// printCommand prints the full kubectl command in a copy-pasteable form
func (c *Client) printCommand(args []string) {
	fmt.Fprintf(os.Stderr, "$ %s\n", CommandString(c.fullArgs(args)...))
}

// CommandString renders a kubectl invocation as a shell-quoted command line
func CommandString(args ...string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, "kubectl")
	for _, a := range args {
		parts = append(parts, shellQuote(a))
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	safe := true
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_@%+=:,./-", r)) {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// runInteractive prints and runs a kubectl command with terminal attached
func (c *Client) runInteractive(args ...string) error {
	c.printCommand(args)
	cmd := exec.Command("kubectl", c.fullArgs(args)...)
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
	output, err := c.runShown("get", "deployment", name, "-o", "yaml")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// Container is a container spec within a deployment's pod template
type Container struct {
	Name  string
	Image string
}

// GetDeploymentContainers returns the (non-init) containers of a deployment (exact name)
func (c *Client) GetDeploymentContainers(name string) ([]Container, error) {
	output, err := c.run("get", "deployment", name, "-o", "json")
	if err != nil {
		return nil, err
	}

	var dep struct {
		Spec struct {
			Template struct {
				Spec struct {
					Containers []struct {
						Name  string `json:"name"`
						Image string `json:"image"`
					} `json:"containers"`
				} `json:"spec"`
			} `json:"template"`
		} `json:"spec"`
	}
	if err := json.Unmarshal(output, &dep); err != nil {
		return nil, fmt.Errorf("failed to parse deployment: %w", err)
	}

	var containers []Container
	for _, ct := range dep.Spec.Template.Spec.Containers {
		containers = append(containers, Container{Name: ct.Name, Image: ct.Image})
	}
	return containers, nil
}

// SetImage sets the image of a container in a deployment (exact names)
func (c *Client) SetImage(deployment, container, image string) error {
	output, err := c.runShown("set", "image", "deployment/"+deployment, container+"="+image)
	if err != nil {
		return err
	}
	fmt.Print(string(output))
	return nil
}

// ReplaceImageTag swaps the tag (and drops any digest) of an image reference,
// e.g. "registry:5000/repo/app:v1" -> "registry:5000/repo/app:<tag>"
func ReplaceImageTag(image, tag string) string {
	if i := strings.Index(image, "@"); i >= 0 {
		image = image[:i]
	}
	if i := strings.LastIndex(image, ":"); i > strings.LastIndex(image, "/") {
		image = image[:i]
	}
	return image + ":" + tag
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
	output, err := c.runShown("get", "configmap", name, "-o", "yaml")
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
	return c.getPods(deploymentName, c.run)
}

// GetPodsByDeploymentNameShown is GetPodsByDeploymentName, printing the kubectl commands it runs
func (c *Client) GetPodsByDeploymentNameShown(deploymentName string) ([]Pod, error) {
	return c.getPods(deploymentName, c.runShown)
}

func (c *Client) getPods(deploymentName string, run func(args ...string) ([]byte, error)) ([]Pod, error) {
	// Get pods with label selector matching the deployment
	output, err := run("get", "pods", "-l", fmt.Sprintf("app=%s", deploymentName), "-o", "json")
	if err != nil {
		// Try without label selector and filter by name
		output, err = run("get", "pods", "-o", "json")
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
