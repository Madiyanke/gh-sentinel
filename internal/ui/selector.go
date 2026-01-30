package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Item struct {
	TitleStr, DescStr string
	ID                int64
	Path              string
}

func (i Item) Title() string       { return i.TitleStr }
func (i Item) Description() string { return i.DescStr }
func (i Item) FilterValue() string { return i.TitleStr }

type model struct {
	list         list.Model
	SelectedItem *Item
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(Item); ok {
				m.SelectedItem = &i
			}
			return m, tea.Quit
		case "q", "ctrl+c":
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
	return docStyle.Render(m.list.View())
}

func StartSelector(items []list.Item) *Item {
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = " SENTINEL-CI: I detected failures "
	p := tea.NewProgram(model{list: l}, tea.WithAltScreen())
	finalModel, _ := p.Run()
	if finalModel == nil { return nil }
	return finalModel.(model).SelectedItem
}