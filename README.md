# k2l

A simplified kubectl wrapper for common operations.

## Installation

```bash
go install github.com/jameskingston/k2l@latest
```

Or build from source:

```bash
git clone https://github.com/jameskingston/k2l
cd k2l
go build -o k2l .
mv k2l /usr/local/bin/
```

## Usage

```
k2l <command> <context> <namespace> <resource>
```

### Commands

| Command | Description |
|---------|-------------|
| `dep`   | Show deployment YAML |
| `conf`  | Show configmap YAML |
| `shell` | Open shell in a pod |
| `pods`  | List pods for a deployment |

### Examples

```bash
# Show deployment YAML
k2l dep prod myapp api-server

# Show configmap YAML  
k2l conf prod myapp api-server

# Open shell in a pod from the deployment
k2l shell prod myapp api-server

# List pods with human-friendly output
k2l pods prod myapp api-server
```

### Substring Matching

All identifiers support substring matching:

```bash
# These are equivalent if they match uniquely:
k2l conf prod myapp api-server
k2l conf prod my api
```

### Dash-Part Matching

Identifiers with dashes support partial matching of each part:

```bash
# "w-f" matches both "web-frontend" and "web-firewall" - a selector will appear
k2l conf prod myapp w-f

# "api-s" matches "api-server"
k2l conf prod default api-s
```

When multiple items match your query, an interactive selector will appear to let you choose.

### History

Run `k2l` without arguments to select from your command history:

```bash
k2l
# Use arrow keys to select a previous command
```

Repeat the most recent command:

```bash
k2l again
# or, using the short alias:
k2l .
```

Clear history:

```bash
k2l history clear
```

## Shell Completion

### Zsh

Add to your `.zshrc`:

```bash
source <(k2l completion zsh)
```

Or install the completion file:

```bash
k2l completion zsh > "${fpath[1]}/_k2l"
```

For the bundled completion with dynamic kubectl lookups:

```bash
cp completions/_k2l "${fpath[1]}/_k2l"
```

### Bash

```bash
source <(k2l completion bash)

# Or for persistent completion:
k2l completion bash > /etc/bash_completion.d/k2l
```

### Fish

```bash
k2l completion fish | source

# Or for persistent completion:
k2l completion fish > ~/.config/fish/completions/k2l.fish
```

## Configuration

History is stored in `~/.config/k2l/history.json`.

## License

MIT
