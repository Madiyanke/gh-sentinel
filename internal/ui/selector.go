package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("#7D56F4")).Bold(true)
	docStyle   = lipgloss.NewStyle().Margin(1, 2)
)

type Item struct {
	TitleStr, DescStr string
	ID                int64
}

func (i Item) Title() string       { return i.TitleStr }
func (i Item) Description() string { return i.DescStr }
func (i Item) FilterValue() string { return i.TitleStr }

type model struct {
	list         list.Model
	SelectedItem *Item
	Quitting     bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				m.SelectedItem = &i
			}
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

func (m model) View() string {
	if m.Quitting { return "" }
	return docStyle.Render(m.list.View())
}

func StartSelector(items []list.Item) *Item {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "SENTINEL-CI : Échecs détectés"
	l.Styles.Title = titleStyle

	m := model{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, _ := p.Run()
	
	return finalModel.(model).SelectedItem
}