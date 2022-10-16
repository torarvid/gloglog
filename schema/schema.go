package schema

import (
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torarvid/gloglog/config"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type Attribute struct {
	Name      string
	Width     int
	Selectors []string
	Type      string
	Format    *string
}

func (a Attribute) FilterValue() string { return "" }

type Model struct {
	Attributes []Attribute
	list       list.Model
	choice     *Attribute
}

func FromLogView(lv config.LogView, width, height int) Model {
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
	items := make([]list.Item, len(attrs))
	for i, attr := range attrs {
		items[i] = attr
	}

	log.Println("Create schema")
	l := list.New(items, itemDelegate{}, width, height)
	l.Title = "Schema attributes"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	return Model{Attributes: attrs, list: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	log.Println("Schema Update")
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		log.Println("Schema: WindowSizeMsg")
		m.list.SetSize(msg.Width-5, msg.Height-5)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			attr, ok := m.list.SelectedItem().(Attribute)
			if ok {
				m.choice = &attr
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.choice != nil {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice.Name))
	}
	return "\n" + m.list.View()
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	attr, ok := listItem.(Attribute)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, attr.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}
