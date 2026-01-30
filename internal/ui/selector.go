package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Bold(true)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

// Item represents a selectable item
type Item struct {
	title string
	desc  string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

// SelectResult contains the result of a selection
type SelectResult struct {
	Index int
	Value string
}

// Model is the bubbletea model for selection
type Model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(Item); ok {
				m.choice = item.title
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

// Select presents a list of items and returns the selected one
func Select(title string, items []string) (*SelectResult, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("no items to select from")
	}

	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = Item{title: item}
	}

	const defaultWidth = 60
	const listHeight = 14

	l := list.New(listItems, itemDelegate{}, defaultWidth, listHeight)
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := Model{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(Model)
	if fm.quitting || fm.choice == "" {
		return nil, nil
	}

	// Find the index
	for i, item := range items {
		if item == fm.choice {
			return &SelectResult{Index: i, Value: fm.choice}, nil
		}
	}

	return &SelectResult{Value: fm.choice}, nil
}

// SelectPod presents a list of pods and returns the selected one
func SelectPod(pods []string) (*SelectResult, error) {
	return Select("Select a pod:", pods)
}

// itemDelegate handles rendering of list items
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(Item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// SimpleSelect provides a simpler selection without bubbletea for quick selections
func SimpleSelect(title string, items []string, defaultIndex int) (*SelectResult, error) {
	return Select(title, items)
}
