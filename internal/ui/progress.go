package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel displays a progress indicator with status
type ProgressModel struct {
	spinner  spinner.Model
	progress progress.Model
	status   string
	percent  float64
	done     bool
	err      error
}

type tickMsg time.Time
type progressMsg float64
type statusMsg string
type doneMsg struct{ err error }

func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case tickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, tea.Batch(cmd, tickCmd())
		}

	case progressMsg:
		m.percent = float64(msg)
		if m.percent >= 1.0 {
			m.done = true
			return m, tea.Quit
		}

	case statusMsg:
		m.status = string(msg)

	case doneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m ProgressModel) View() string {
	if m.done {
		if m.err != nil {
			return errorStyle.Render(fmt.Sprintf("✗ %s: %v\n", m.status, m.err))
		}
		return successStyle.Render(fmt.Sprintf("✓ %s\n", m.status))
	}

	return fmt.Sprintf("\n%s %s\n\n", m.spinner.View(), infoStyle.Render(m.status))
}

// NewProgressModel creates a new progress model
func NewProgressModel(status string) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	p := progress.New(progress.WithDefaultGradient())

	return ProgressModel{
		spinner: s,
		progress: p,
		status:  status,
	}
}

// ConfirmationModel handles yes/no confirmations
type ConfirmationModel struct {
	prompt   string
	details  string
	confirmed bool
	cancelled bool
}

func (m ConfirmationModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.confirmed = true
			return m, tea.Quit
		case "n", "N", "q", "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ConfirmationModel) View() string {
	var b strings.Builder

	b.WriteString(warningStyle.Render("⚠  " + m.prompt) + "\n\n")
	
	if m.details != "" {
		b.WriteString(dimStyle.Render(m.details) + "\n\n")
	}

	b.WriteString(infoStyle.Render("Press [y] to confirm, [n] to cancel") + "\n")

	return b.String()
}

// NewConfirmationModel creates a new confirmation dialog
func NewConfirmationModel(prompt, details string) ConfirmationModel {
	return ConfirmationModel{
		prompt:  prompt,
		details: details,
	}
}

// ShowConfirmation displays a confirmation dialog and returns the result
func ShowConfirmation(prompt, details string) (bool, error) {
	model := NewConfirmationModel(prompt, details)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	if m, ok := finalModel.(ConfirmationModel); ok {
		return m.confirmed, nil
	}

	return false, nil
}

// DiffViewerModel displays a diff comparison
type DiffViewerModel struct {
	title    string
	diff     string
	viewport int
	quitting bool
}

func (m DiffViewerModel) Init() tea.Cmd {
	return nil
}

func (m DiffViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.viewport = msg.Height - 4
	}
	return m, nil
}

func (m DiffViewerModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(m.title) + "\n\n")

	// Color diff lines
	lines := strings.Split(m.diff, "\n")
	displayLines := lines
	if len(lines) > m.viewport && m.viewport > 0 {
		displayLines = lines[:m.viewport]
	}

	for _, line := range displayLines {
		if strings.HasPrefix(line, "+") {
			b.WriteString(successStyle.Render(line) + "\n")
		} else if strings.HasPrefix(line, "-") {
			b.WriteString(errorStyle.Render(line) + "\n")
		} else if strings.HasPrefix(line, "===") {
			b.WriteString(headerStyle.Render(line) + "\n")
		} else {
			b.WriteString(dimStyle.Render(line) + "\n")
		}
	}

	if len(lines) > m.viewport && m.viewport > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("\n... (%d more lines)", len(lines)-m.viewport)) + "\n")
	}

	b.WriteString("\n" + infoStyle.Render("Press any key to continue") + "\n")

	return b.String()
}

// NewDiffViewerModel creates a new diff viewer
func NewDiffViewerModel(title, diff string) DiffViewerModel {
	return DiffViewerModel{
		title:    title,
		diff:     diff,
		viewport: 20,
	}
}

// ShowDiff displays a diff viewer
func ShowDiff(title, diff string) error {
	model := NewDiffViewerModel(title, diff)
	p := tea.NewProgram(model)
	_, err := p.Run()
	return err
}
