package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
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
	rows := logView.GetRows()
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
	view := config.TheConfig.GetActiveView()
	view.Attrs = attrs
	config.TheConfig.Save()
}

func (m model) View() string {
	if !firstDraw {
		timeToFirstDraw := time.Since(appStartTime)
		slog.Info("First draw happened", "delay", timeToFirstDraw)
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
		valueGetter: valueGetterFromSelectors(c.Selectors, c.Type, c.Format),
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

// valueGetterFromSelectors returns a function that can be used to get column values.
//
// For now it is useful for getting values from JSON data (including nested JSON data).
//
// Example:
//
//	valueGetterFromSelectors([]string{"json(data.barcode)"}, "", nil)
//
// The call above will return a function that can be called for each log file entry to retrieve
// barcode information. It uses gjson to extract the value from JSON looking like
// {"data": {"barcode": "1234", "whatever": 111}, "other_data": []}.
//
// If the extracted data is of a formattable type (e.g. time), the typ+format parameters can be
// used to format the value. For now, only time is supported.
//
// In cases where you have nested data (say {"data": "{\"barcode\": \"1234\"}"}), you can use the
// | character to pipe the result of one parsed json into another. For example:
//
//	valueGetterFromSelectors([]string{"json(data)|json(barcode)"}, "", nil)
//
// In cases where your log source has imperfectly structured data, you can use the fact that the
// 'selectors' parameter is a slice. This means that you can provide multiple selectors and the
// first one to yield a non-empty result is returned.
func valueGetterFromSelectors(selectors []string, typ string, format *string) func(string) string {
	getters := make([]func(string) string, len(selectors))
	for i, selector := range selectors {
		if selector == "." {
			getters[i] = identity
			continue
		}
		partialSelectors := strings.Split(selector, "|")
		getters[i] = func(s string) string {
			for _, sel := range partialSelectors {
				sel = strings.TrimSpace(sel)
				if strings.HasPrefix(sel, "json(") {
					// set jsonPath to whatever is inside the parentheses of
					// 'json(...)'
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
	return func(input string) (out string) {
		for _, getter := range getters {
			out = getter(input)
			if out != "" {
				break
			}
		}
		return out
	}
}
