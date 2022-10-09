package help

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	focused bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Println("help update: ", msg)
	return m, nil
}

func (m Model) View() string {
	return "Help text here"
}

func (m *Model) Blur() {
	log.Println("help blur: ")
	m.focused = false
}

func (m *Model) Focus() {
	log.Println("help focus: ")
	m.focused = true
}

func (m Model) Focused() bool {
	log.Println("help focused: ", m.focused)
	return m.focused
}
