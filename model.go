package main

import (
	"encoding/json"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidwall/gjson"
	"github.com/torarvid/gloglog/table"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table   table.Model[string]
	zoomRow bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case " ":
			m.zoomRow = !m.zoomRow
		}

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 5)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.zoomRow {
		rawRow := m.table.SelectedRow()
		var row map[string]interface{}
		json.Unmarshal([]byte(rawRow), &row)
		rendered, err := json.MarshalIndent(row, "", "    ")
		if err != nil {
			return "Invalid"
		}
		return baseStyle.Render(string(rendered))
	}
	return baseStyle.Render(m.table.View()) + "\n"
}

type Column struct {
	title       string
	width       int
	valueGetter func(string) string
}

func ColumnFromConfig(c SearchColumn) *Column {
	return &Column{
		title:       c.Name,
		width:       c.Width,
		valueGetter: valueFromSelector(c.Selector, c.Type, c.Format),
	}
}

func (c *Column) Title() string {
	return c.title
}

func (c *Column) Width() int {
	return c.width
}

func (c *Column) SetWidth(width int) {
	c.width = width
}

func (c *Column) GetValue(s string) string {
	return c.valueGetter(s)
}

func identity(s string) string {
	return s
}

func valueFromSelector(selector, typ string, format *string) func(string) string {
	if selector == "." {
		return identity
	}
	partialSelectors := strings.Split(selector, "|")
	return func(s string) string {
		for _, sel := range partialSelectors {
			sel = strings.TrimSpace(sel)
			if strings.HasPrefix(sel, "json(") {
				jsonPath := sel[5 : len(sel)-1]
				if jsonPath == "." {
					continue
				}
				s = gjson.Get(s, jsonPath).String()
			}
		}
		switch typ {
		case "time":
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return "Invalid"
			}
			if format == nil {
				return t.Format(time.StampMilli)
			}
			return t.Format(*format)
		default:
			return s
		}
	}
}
