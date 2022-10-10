package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
	"github.com/tidwall/gjson"
	"github.com/torarvid/gloglog/table"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model[string]
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
		}

	case tea.WindowSizeMsg:
		m.table.SetWidth(msg.Width - 2)
		m.table.SetHeight(msg.Height - 5)
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

type Column struct {
	title       string
	width       int
	valueGetter func(string) string
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

func timestampGetter(s string) string {
	var js map[string]any
	err := json.Unmarshal([]byte(s), &js)
	ts := ""
	if err == nil {
		ts = js["timestamp"].(string)
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "Invalid"
	}
	return t.Format("15:04:05.000")
}

func identity(s string) string {
	return s
}

func getValueFromPath(path, typ string, format *string) func(string) string {
	if path == "." {
		return identity
	}
	return func(s string) string {
		val := gjson.Get(s, path)
		str := val.String()
		switch typ {
		case "time":
			t, err := time.Parse(time.RFC3339, str)
			if err != nil {
				return "Invalid"
			}
			if format == nil {
				return t.Format(time.StampMilli)
			}
			return t.Format(*format)
		default:
			return str
		}
	}
}

func loadConfig() Config {
	configString, err := ioutil.ReadFile("foo.toml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = toml.Unmarshal([]byte(configString), &config)
	if err != nil {
		panic(err)
	}
	return config
}

func main() {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)

	config := loadConfig()
	search := config.SavedSearches[0]
	columns := make([]table.ColumnSpec[string], len(search.Columns))
	for i, c := range search.Columns {
		columns[i] = &Column{
			title:       c.Name,
			width:       c.Width,
			valueGetter: getValueFromPath(c.Path, c.Type, c.Format),
		}
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

	m := model{t}
	if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
