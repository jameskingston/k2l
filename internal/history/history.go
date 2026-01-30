package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Entry represents a single history entry
type Entry struct {
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// History manages command history
type History struct {
	path    string
	entries map[string]*Entry
}

// New creates a new History instance
func New() (*History, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "k2l")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	h := &History{
		path:    filepath.Join(configDir, "history.json"),
		entries: make(map[string]*Entry),
	}

	if err := h.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return h, nil
}

// load reads history from disk
func (h *History) load() error {
	file, err := os.Open(h.path)
	if err != nil {
		return err
	}
	defer file.Close()

	var entries []*Entry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return err
	}

	for _, e := range entries {
		key := h.makeKey(e.Command, e.Args)
		h.entries[key] = e
	}

	return nil
}

// save writes history to disk
func (h *History) save() error {
	entries := h.GetSortedEntries()

	file, err := os.Create(h.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// makeKey creates a unique key for a command
func (h *History) makeKey(command string, args []string) string {
	key := command
	for _, arg := range args {
		key += ":" + arg
	}
	return key
}

// Add adds or updates a history entry
func (h *History) Add(command string, args []string) error {
	key := h.makeKey(command, args)

	if entry, exists := h.entries[key]; exists {
		entry.Count++
		entry.Timestamp = time.Now()
	} else {
		h.entries[key] = &Entry{
			Command:   command,
			Args:      args,
			Timestamp: time.Now(),
			Count:     1,
		}
	}

	return h.save()
}

// GetSortedEntries returns entries sorted by most recent first
func (h *History) GetSortedEntries() []*Entry {
	entries := make([]*Entry, 0, len(h.entries))
	for _, e := range h.entries {
		entries = append(entries, e)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})

	return entries
}

// GetDisplayStrings returns formatted strings for display
func (h *History) GetDisplayStrings() []string {
	entries := h.GetSortedEntries()
	strs := make([]string, len(entries))
	for i, e := range entries {
		str := e.Command
		for _, arg := range e.Args {
			str += " " + arg
		}
		strs[i] = str
	}
	return strs
}

// GetEntry returns an entry by index (sorted by most recent)
func (h *History) GetEntry(index int) *Entry {
	entries := h.GetSortedEntries()
	if index < 0 || index >= len(entries) {
		return nil
	}
	return entries[index]
}

// Clear clears all history
func (h *History) Clear() error {
	h.entries = make(map[string]*Entry)
	return h.save()
}

// GetHistoryPath returns the path to the history file
func GetHistoryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "k2l", "history.json"), nil
}

// LoadContextsFromKubeconfig reads available contexts from kubeconfig
func LoadContextsFromKubeconfig() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")
	file, err := os.Open(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var contexts []string
	scanner := bufio.NewScanner(file)
	inContexts := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "contexts:" {
			inContexts = true
			continue
		}
		if inContexts {
			if len(line) > 0 && line[0] != ' ' && line[0] != '-' {
				break
			}
			if len(line) > 8 && line[:8] == "- name: " {
				contexts = append(contexts, line[8:])
			}
		}
	}

	return contexts, scanner.Err()
}
