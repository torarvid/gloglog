package schema

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torarvid/gloglog/config"
)

type Attribute struct {
	Name      string
	Width     int
	Selectors []string
	Type      string
	Format    *string
}

type Model struct {
	Attributes []Attribute
}

func FromLogView(lv config.LogView) Model {
	attrs := make([]Attribute, len(lv.Attrs))
	for i, attr := range lv.Attrs {
		attrs[i] = Attribute{
			Name:      attr.Name,
			Width:     attr.Width,
			Selectors: []string{attr.Selector},
			Type:      attr.Type,
			Format:    attr.Format,
		}
	}
	return Model{Attributes: attrs}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	output := ""
	for _, attr := range m.Attributes {
		output += attr.Name
	}
	return output
}
