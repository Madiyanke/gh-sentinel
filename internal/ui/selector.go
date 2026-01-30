package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the TUI
var (
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	docStyle = lipgloss.NewStyle().
		Margin(1, 2)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("214")).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("86"))

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true).
		Underline(true)

	highlightStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("226")).
		Bold(true)
)

// WorkflowItem represents a workflow run in the list
type WorkflowItem struct {
	ID          int64
	TitleText   string // Title text for display
	DescText    string // Description text
	Status      string
	Conclusion  string
	Path        string
	Icon        string
}

func (i WorkflowItem) FilterValue() string {
	return i.TitleText
}

func (i WorkflowItem) Title() string {
	return fmt.Sprintf("%s  %s", i.Icon, i.TitleText)
}

func (i WorkflowItem) Description() string {
	return i.DescText
}

// WorkflowSelectorModel is the model for workflow selection
type WorkflowSelectorModel struct {
	list     list.Model
	selected *WorkflowItem
	quitting bool
}

func (m WorkflowSelectorModel) Init() tea.Cmd {
	return nil
}

func (m WorkflowSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(WorkflowItem); ok {
				m.selected = &item
				m.quitting = true
				return m, tea.Quit
			}
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m WorkflowSelectorModel) View() string {
	if m.quitting {
		if m.selected != nil {
			return successStyle.Render(fmt.Sprintf("✓ Selected: %s", m.selected.TitleText))
		}
		return dimStyle.Render("Operation cancelled")
	}
	return docStyle.Render(m.list.View())
}

// NewWorkflowSelector creates a new workflow selector
func NewWorkflowSelector(items []WorkflowItem) *WorkflowSelectorModel {
	// Convert to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	// Create custom delegate with better styling
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("205")).
		BorderForeground(lipgloss.Color("205"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240"))

	l := list.New(listItems, delegate, 0, 0)
	l.Title = "🛡️  Sentinel CI - Workflow Runs"
	l.Styles.Title = titleStyle

	return &WorkflowSelectorModel{
		list: l,
	}
}

// GetSelected returns the selected item after the program exits
func (m *WorkflowSelectorModel) GetSelected() *WorkflowItem {
	return m.selected
}

// ShowWorkflowSelector displays the workflow selector and returns the selected item
func ShowWorkflowSelector(items []WorkflowItem) (*WorkflowItem, error) {
	model := NewWorkflowSelector(items)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if m, ok := finalModel.(WorkflowSelectorModel); ok {
		return m.GetSelected(), nil
	}

	return nil, nil
}

// FormatSuccess returns a success message with styling
func FormatSuccess(msg string) string {
	return successStyle.Render("✓ " + msg)
}

// FormatError returns an error message with styling
func FormatError(msg string) string {
	return errorStyle.Render("✗ " + msg)
}

// FormatWarning returns a warning message with styling
func FormatWarning(msg string) string {
	return warningStyle.Render("⚠ " + msg)
}

// FormatInfo returns an info message with styling
func FormatInfo(msg string) string {
	return infoStyle.Render("ℹ " + msg)
}

// FormatHeader returns a header with styling
func FormatHeader(msg string) string {
	return headerStyle.Render(msg)
}

// FormatHighlight returns highlighted text
func FormatHighlight(msg string) string {
	return highlightStyle.Render(msg)
}

// FormatDim returns dimmed text
func FormatDim(msg string) string {
	return dimStyle.Render(msg)
}

// PrintBanner displays the application banner
func PrintBanner() {
	banner := `
🛡️  ╔═══════════════════════════════════════╗
   ║     SENTINEL CI - DevOps Guardian    ║
   ║   AI-Powered CI/CD Pipeline Repair   ║
   ╚═══════════════════════════════════════╝
`
	fmt.Println(infoStyle.Render(banner))
}