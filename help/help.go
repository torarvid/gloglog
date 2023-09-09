package help

import (
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	focused bool
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	slog.Info("help update: ", "msg", msg)
	return m, nil
}

func (m Model) View() string {
	return "Help text here"
}

func (m *Model) Blur() {
	slog.Info("help blur: ")
	m.focused = false
}

func (m *Model) Focus() {
	slog.Info("help focus: ")
	m.focused = true
}

func (m Model) Focused() bool {
	slog.Info("help focused: ", "focused", m.focused)
	return m.focused
}
