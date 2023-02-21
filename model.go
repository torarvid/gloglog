package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tidwall/gjson"
	"github.com/torarvid/gloglog/config"
	"github.com/torarvid/gloglog/schema"
	"github.com/torarvid/gloglog/search"
	"github.com/torarvid/gloglog/table"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240"))

const (
	stateTable = iota
	stateZoomRow
	stateSchema
	stateSearch
)

var stateOverlays []int = []int{stateZoomRow, stateSchema, stateSearch}

type model struct {
	table        table.Model[string]
	schema       schema.Model
	rows         []string
	filteredRows []string
	view         config.LogView
	filters      []RowFilter
	search       search.Model
	state        int
	termWidth    int
	termHeight   int
}

func newModel(logView config.LogView) *model {
	modelInitTime := time.Now()
	filename, exists := logView.Options["filename"]
	if !exists {
		panic("filename not found")
	}
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	fmt.Fprintf(os.Stderr, "Scanning file '%s'", filename)
	rows := make([]string, 0)
	for scanner.Scan() {
		rows = append(rows, scanner.Text())
	}
	scanTime := time.Since(modelInitTime)
	fmt.Fprintf(os.Stderr, " done in %d ms. %d rows found.\n", scanTime.Milliseconds(), len(rows))

	t := table.New(
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

	m := &model{
		table:   t,
		state:   stateTable,
		rows:    rows,
		view:    logView,
		filters: make([]RowFilter, 0),
		schema:  schema.FromLogView(logView, 1, 1),
		search:  search.FromLogView(logView, 40, 15),
	}
	m.updateColumns(logView.Attrs)
	m.table.SetRows(m.rows)
	return m
}

type RowFilter func(string) bool

func (rf RowFilter) Filter(s string) bool {
	return rf(s)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	var cmd tea.Cmd
	switch m.state {
	case stateTable:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case " ":
				m.state = stateZoomRow
			case "s":
				m.state = stateSchema
			case "/":
				m.state = stateSearch
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	case stateZoomRow:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case " ", "esc":
				m.state = stateTable
			}
		}
	case stateSchema:
		switch msg := msg.(type) {
		case schema.Close:
			m.state = stateTable
			return m, nil
		case schema.UpdatedSchemaMsg:
			m.updateColumns(msg.Attributes)
		}
		m.schema, cmd = m.schema.Update(msg)
		cmds = append(cmds, cmd)
	case stateSearch:
		switch msg := msg.(type) {
		case search.Close:
			m.state = stateTable
			return m, nil
		case search.UpdatedFiltersMsg:
			m.SetFilters(msg.Filters)
		}
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
		m.schema, cmd = m.schema.Update(msg)
		cmds = append(cmds, cmd)
		m.termWidth, m.termHeight = msg.Width, msg.Height
	}
	return m, tea.Batch(cmds...)
}

func (m *model) updateColumns(attrs []config.Attribute) {
	columns := make([]table.ColumnSpec[string], len(attrs))
	for i, c := range attrs {
		columns[i] = ColumnFromConfig(c)
	}
	m.table.SetColumns(columns)
}

func (m model) View() string {
	if !firstDraw {
		timeToFirstDraw := time.Since(appStartTime)
		log.Printf("Time to first draw: %s\n", timeToFirstDraw)
		firstDraw = true
	}

	switch m.state {

	case stateTable:
		return baseStyle.Render(m.table.View())

	case stateZoomRow:
		rawRow := m.table.SelectedRow()
		var row map[string]interface{}
		err := json.Unmarshal([]byte(rawRow), &row)
		if err != nil {
			return "Invalid"
		}
		rendered, err := json.MarshalIndent(row, "", "    ")
		if err != nil {
			return "Invalid"
		}
		return baseStyle.Width(m.termWidth - 2).Render(string(rendered))

	case stateSchema:
		return baseStyle.Render(m.schema.View())

	case stateSearch:
		return baseStyle.Render(m.search.View())

	default:
		panic("Unknown state")
	}
}

func (m *model) SetFilters(filters []config.Filter) {
	rowFilters := make([]RowFilter, len(filters))
	for i, filter := range filters {
		filter := filter
		rowFilter := func(row string) bool {
			return strings.Contains(row, filter.Term)
		}
		rowFilters[i] = rowFilter
	}
	m.filters = rowFilters
	m.updateFilteredRows()
	m.table.SetRows(m.filteredRows)
}

func (m *model) updateFilteredRows() {
	m.filteredRows = make([]string, 0, len(m.rows)/10)
	for _, row := range m.rows {
		include := true
		for _, filter := range m.filters {
			if !filter(row) {
				include = false
				break
			}
		}
		if include {
			m.filteredRows = append(m.filteredRows, row)
		}
	}
}

type Column struct {
	title       string
	width       int
	valueGetter func(string) string
}

func ColumnFromConfig(c config.Attribute) *Column {
	return &Column{
		title:       c.Name,
		width:       c.Width,
		valueGetter: valueFromSelector(c.Selectors, c.Type, c.Format),
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

func valueFromSelector(selectors []string, typ string, format *string) func(string) string {
	for _, selector := range selectors {
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
	return nil
}
