package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
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

const (
	stateTable = iota
	stateZoomRow
	stateSearch
)

type model struct {
	table table.Model[string]
	state int
}

func newModel(search SavedSearch) *model {
	columns := make([]table.ColumnSpec[string], len(search.Columns))
	for i, c := range search.Columns {
		columns[i] = ColumnFromConfig(c)
	}

	filename, exists := search.Options["filename"]
	if !exists {
		panic("filename not found")
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rows := make([]string, 0)
	for scanner.Scan() {
		rows = append(rows, scanner.Text())
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused[string](true),
		table.WithHeight[string](27),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("27")).
		Bold(false)
	t.SetStyles(s)

	return &model{table: t, state: stateTable}
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
			if m.state == stateZoomRow {
				m.state = stateTable
			} else {
				m.state = stateZoomRow
			}
		}

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 5)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.state {

	case stateTable:
		return baseStyle.Render(m.table.View()) + "\n"

	case stateZoomRow:
		rawRow := m.table.SelectedRow()
		var row map[string]interface{}
		json.Unmarshal([]byte(rawRow), &row)
		rendered, err := json.MarshalIndent(row, "", "    ")
		if err != nil {
			return "Invalid"
		}
		return baseStyle.Render(string(rendered))

	default:
		panic("Unknown state")
	}
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
