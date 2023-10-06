package search

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torarvid/gloglog/config"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).MarginTop(1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().
				HelpStyle.PaddingLeft(4).
				PaddingBottom(1).
				PaddingRight(4)
	detailStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			PaddingLeft(2).
			PaddingRight(4)
)

type KeyMap struct {
	SelectNextField key.Binding
	SelectPrevField key.Binding
	EditFilter      key.Binding
	NewFilter       key.Binding
	Exit            key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SelectNextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("next field", "Tab"),
		),
		SelectPrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("previous field", "Shift+Tab"),
		),
		EditFilter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("edit filter", "Enter"),
		),
		NewFilter: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("new filter", "Ctrl+N"),
		),
		Exit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("exit", "Esc"),
		),
	}
}

type Filter struct {
	Term     string
	Operator config.FilterOp
	Attr     *config.Attribute
	inputs   []textinput.Model
}

func (f Filter) FilterValue() string { return "" }

func (a Filter) View() string {
	labels := []string{"Term", "Operator", "Attr"}
	parts := make([]string, len(labels))
	for i, label := range labels {
		parts[i] = label + "\n" + a.inputs[i].View()
	}

	return strings.Join(parts, "\n\n")
}

type Model struct {
	Filters  []Filter
	list     list.Model
	selected *int
	keyMap   KeyMap
	logView  config.LogView
}

func newFilter(
	term string,
	op config.FilterOp,
	attr *config.Attribute,
	attrPlaceholder string,
) Filter {
	termInput := textinput.New()
	termInput.Placeholder = "Filter term"
	termInput.SetValue(term)
	termInput.CharLimit = 300
	termInput.Focus()

	opInput := textinput.New()
	opInput.Placeholder = "contains"
	opInput.SetValue(string(op))
	opInput.Blur()

	attrInput := textinput.New()
	if attrPlaceholder == "" {
		attrPlaceholder = "<name of attribute>"
	}
	attrInput.Placeholder = attrPlaceholder
	if attr != nil {
		attrInput.SetValue(attr.Name)
	}
	attrInput.Blur()

	return Filter{
		Term:     term,
		Operator: op,
		Attr:     attr,
		inputs:   []textinput.Model{termInput, opInput, attrInput},
	}
}

func FromLogView(lv config.LogView, width, height int) Model {
	filters := make([]Filter, len(lv.Filters))
	attrPlaceholder := ""
	if len(lv.Attrs) > 0 {
		attrPlaceholder = lv.Attrs[0].Name
	}
	for i, filter := range lv.Filters {
		filters[i] = newFilter(filter.Term, filter.Operator, filter.Attr, attrPlaceholder)
	}
	items := listItemsFromFilters(filters)
	keyMap := DefaultKeyMap()
	keys := []key.Binding{keyMap.SelectNextField, keyMap.SelectPrevField}

	l := list.New(items, itemDelegate{keys}, width, height)
	l.Title = "Search filters"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.KeyMap.NextPage.SetEnabled(false)
	l.KeyMap.PrevPage.SetEnabled(false)
	l.KeyMap.GoToStart.SetEnabled(false)
	l.KeyMap.GoToEnd.SetEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.DisableQuitKeybindings()
	return Model{Filters: filters, list: l, keyMap: keyMap, logView: lv}
}

func listItemsFromFilters(filters []Filter) []list.Item {
	items := make([]list.Item, len(filters))
	for i, filter := range filters {
		filter := filter
		items[i] = &filter
	}
	return items
}

func (m Model) Init() tea.Cmd {
	return nil
}

type Close struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	slog.Info("search update", "msg", msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-5, msg.Height-5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.SelectNextField):
			m.focusNextInput()
		case key.Matches(msg, m.keyMap.SelectPrevField):
			m.focusPrevInput()
		case key.Matches(msg, m.keyMap.Exit):
			if m.selected == nil {
				return m, func() tea.Msg { return Close{} }
			}
			m.deselect()
		case key.Matches(msg, m.keyMap.EditFilter):
			if m.selected == nil {
				m.selectFilter(m.list.Index())
				return m, nil
			} else {
				i := *m.selected
				m.Filters[i].Term = m.Filters[i].inputs[0].Value()
				if op, err := config.ParseFilterOp(m.Filters[i].inputs[1].Value()); err == nil {
					m.Filters[i].Operator = op
				}
				if attr, err := m.logView.GetAttributeWithName(m.Filters[i].inputs[2].Value()); err == nil {
					m.Filters[i].Attr = attr
				}
				m.deselect()
				m.list.SetItems(listItemsFromFilters(m.Filters))
				return m, m.UpdateFilters()
			}
		case key.Matches(msg, m.keyMap.NewFilter):
			m.Filters = append(m.Filters, newFilter("", config.Contains, nil, ""))
			m.list.SetItems(listItemsFromFilters(m.Filters))
			m.selectFilter(len(m.Filters) - 1)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	slog.Info("sucmd", "cmd", cmd)
	cmds = append(cmds, cmd)
	if m.selected != nil {
		inputs := m.Filters[*m.selected].inputs
		for i, input := range inputs {
			if !input.Focused() {
				continue
			}
			inputs[i], cmd = input.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) focusNextInput() {
	inputs := m.Filters[*m.selected].inputs
	for i := range inputs {
		if inputs[i].Focused() {
			inputs[i].Blur()
			inputs[(i+1)%len(inputs)].Focus()
			return
		}
	}
}

func (m *Model) focusPrevInput() {
	inputs := m.Filters[*m.selected].inputs
	for i, input := range inputs {
		if input.Focused() {
			inputs[i].Blur()
			inputs[(i-1+len(inputs))%len(inputs)].Focus()
			return
		}
	}
}

type UpdatedFiltersMsg struct {
	Filters []config.Filter
}

func (m Model) UpdateFilters() tea.Cmd {
	return func() tea.Msg {
		cfgFilters := make([]config.Filter, len(m.Filters))
		for i, filter := range m.Filters {
			cfgFilters[i] = config.Filter{
				Term:     filter.Term,
				Operator: filter.Operator,
				Attr:     filter.Attr,
			}
		}
		return UpdatedFiltersMsg{Filters: cfgFilters}
	}
}

func (m Model) View() string {
	filterList := m.list.View()
	selectionView := ""
	if m.selected != nil {
		selectionView = detailStyle.Height(m.list.Height()).Render(m.Filters[*m.selected].View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, filterList, selectionView)
}

func (m *Model) selectFilter(i int) {
	m.selected = &i
	m.list.KeyMap.CursorDown.SetEnabled(false)
	m.list.KeyMap.CursorUp.SetEnabled(false)
	m.keyMap.SelectNextField.SetEnabled(true)
	m.keyMap.SelectPrevField.SetEnabled(true)
}

func (m *Model) deselect() {
	m.selected = nil
	m.list.KeyMap.CursorDown.SetEnabled(true)
	m.list.KeyMap.CursorUp.SetEnabled(true)
	m.keyMap.SelectNextField.SetEnabled(false)
	m.keyMap.SelectPrevField.SetEnabled(false)
}

type itemDelegate struct {
	keys []key.Binding
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	filter, ok := listItem.(*Filter)
	if !ok {
		return
	}

	name := "[anything]"
	if filter.Attr != nil {
		name = filter.Attr.Name
	}
	str := fmt.Sprintf("%d. %s", index+1, name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}
func (d itemDelegate) ShortHelp() []key.Binding  { return d.keys }
func (d itemDelegate) FullHelp() [][]key.Binding { return nil }
