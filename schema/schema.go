package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
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
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1).PaddingRight(4)
	detailStyle       = lipgloss.NewStyle().
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				PaddingLeft(2).
				PaddingRight(4)
)

type Attribute struct {
	Name      string
	Width     int
	Selectors []string
	Type      string
	Format    *string
	inputs    []textinput.Model
}

func (a Attribute) FilterValue() string { return "" }

func (a Attribute) View() string {
	labels := []string{"Name", "Width", "Selectors", "Type", "Format"}
	parts := make([]string, len(labels))
	for i, label := range labels {
		parts[i] = label + "\n" + a.inputs[i].View()
	}

	return strings.Join(parts, "\n\n")
}

type KeyMapMain struct {
	EnterDetail key.Binding
	Exit        key.Binding
	NewField    key.Binding
	DeleteField key.Binding
}

type KeyMapDetail struct {
	SelectNextField key.Binding
	SelectPrevField key.Binding
}

type Model struct {
	Attributes   []Attribute
	list         list.Model
	selected     *int
	keyMap       KeyMapMain
	detailKeyMap KeyMapDetail
}

func createAttribute(attr config.Attribute) Attribute {
	nameInput := textinput.New()
	nameInput.Placeholder = "Title/Name/Barcode/Host/..."
	nameInput.SetValue(attr.Name)
	nameInput.CharLimit = 30
	nameInput.Focus()

	widthInput := textinput.New()
	widthInput.Placeholder = "15"
	widthInput.SetValue(strconv.Itoa(attr.Width))
	widthInput.Blur()

	selectorsInput := textinput.New()
	selectorsInput.Placeholder = "error.code / server.host.ip / ..."
	selectorsInput.SetValue("[\"" + strings.Join(attr.Selectors, "\", \"") + "\"]")
	selectorsInput.Blur()

	typeInput := textinput.New()
	typeInput.Placeholder = "string / int / time / ..."
	typeInput.SetValue(attr.Type)
	typeInput.Blur()

	formatInput := textinput.New()
	formatInput.Placeholder = "2006-01-02 15:04:05.000"
	if attr.Format != nil {
		formatInput.SetValue(*attr.Format)
	}
	formatInput.Blur()

	return Attribute{
		Name:      attr.Name,
		Width:     attr.Width,
		Selectors: attr.Selectors,
		Type:      attr.Type,
		Format:    attr.Format,
		inputs:    []textinput.Model{nameInput, widthInput, selectorsInput, typeInput, formatInput},
	}
}

func FromLogView(lv config.LogView, width, height int) Model {
	attrs := make([]Attribute, len(lv.Attrs))
	for i, attr := range lv.Attrs {
		attrs[i] = createAttribute(attr)
	}
	items := listItemsFromAttributes(attrs)

	log.Println("Create schema")
	keyMap := MainKeyMap()
	keys := []key.Binding{keyMap.EnterDetail, keyMap.Exit, keyMap.NewField, keyMap.DeleteField}
	l := list.New(items, itemDelegate{keys}, width, height)
	l.Title = "Schema attributes"
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
	return Model{Attributes: attrs, list: l, keyMap: keyMap, detailKeyMap: DetailKeyMap()}

}

func listItemsFromAttributes(attrs []Attribute) []list.Item {
	items := make([]list.Item, len(attrs))
	for i, attr := range attrs {
		attr := attr
		items[i] = &attr
	}
	return items
}

func MainKeyMap() KeyMapMain {
	return KeyMapMain{
		EnterDetail: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter detail", "Enter"),
		),
		Exit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("exit", "Esc"),
		),
		NewField: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("new field", "n"),
		),
		DeleteField: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("delete field", "d"),
		),
	}
}
func DetailKeyMap() KeyMapDetail {
	return KeyMapDetail{
		SelectNextField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("next field", "Tab"),
		),
		SelectPrevField: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("previous field", "Shift+Tab"),
		),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

type Close struct{}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width-5, msg.Height-5)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.EnterDetail):
			if m.selected == nil {
				m.selectAttr(m.list.Index())
				return m, nil
			}
			i := *m.selected
			m.Attributes[i].Name = m.Attributes[i].inputs[0].Value()
			if w, err := strconv.Atoi(m.Attributes[i].inputs[1].Value()); err == nil {
				m.Attributes[i].Width = w
			}
			var selectors []string
			if err := json.Unmarshal([]byte(m.Attributes[i].inputs[2].Value()), &selectors); err == nil {
				m.Attributes[i].Selectors = selectors
			}
			m.Attributes[i].Type = m.Attributes[i].inputs[3].Value()
			if m.Attributes[i].inputs[4].Value() != "" {
				format := m.Attributes[i].inputs[4].Value()
				m.Attributes[i].Format = &format
			} else {
				m.Attributes[i].Format = nil
			}
			m.deselect()
			m.list.SetItems(listItemsFromAttributes(m.Attributes))
			return m, m.UpdateSchema()
		case key.Matches(msg, m.keyMap.Exit):
			if m.selected == nil {
				return m, func() tea.Msg { return Close{} }
			}
			m.deselect()
		case key.Matches(msg, m.keyMap.NewField):
			if m.selected != nil {
				break
			}
			attr := createAttribute(config.Attribute{})
			m.Attributes = append(m.Attributes, attr)
			listItems := m.list.Items()
			listItems = append(listItems, &attr)
			m.list.SetItems(listItems)
			m.selectAttr(len(m.Attributes) - 1)
			return m, nil
		case key.Matches(msg, m.keyMap.DeleteField):
			if m.selected != nil {
				break
			}
			if len(m.Attributes) == 0 {
				return m, nil
			}
			m.Attributes = append(
				m.Attributes[:m.list.Index()], m.Attributes[m.list.Index()+1:]...,
			)
			listItems := m.list.Items()
			listItems = append(
				listItems[:m.list.Index()], listItems[m.list.Index()+1:]...,
			)
			m.list.SetItems(listItems)
			return m, m.UpdateSchema()
		case key.Matches(msg, m.detailKeyMap.SelectNextField):
			m.focusNextInput()
		case key.Matches(msg, m.detailKeyMap.SelectPrevField):
			m.focusPrevInput()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	if m.selected != nil {
		inputs := m.Attributes[*m.selected].inputs
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
	inputs := m.Attributes[*m.selected].inputs
	for i := range inputs {
		if inputs[i].Focused() {
			inputs[i].Blur()
			inputs[(i+1)%len(inputs)].Focus()
			return
		}
	}
}

func (m *Model) focusPrevInput() {
	inputs := m.Attributes[*m.selected].inputs
	for i, input := range inputs {
		if input.Focused() {
			inputs[i].Blur()
			inputs[(i-1+len(inputs))%len(inputs)].Focus()
			return
		}
	}
}

type UpdatedSchemaMsg struct {
	Attributes []config.Attribute
}

func (m Model) UpdateSchema() tea.Cmd {
	return func() tea.Msg {
		cfgAttributes := make([]config.Attribute, len(m.Attributes))
		for i, attr := range m.Attributes {
			cfgAttributes[i] = config.Attribute{
				Name:      attr.Name,
				Width:     attr.Width,
				Selectors: attr.Selectors,
				Type:      attr.Type,
				Format:    attr.Format,
			}
		}
		return UpdatedSchemaMsg{Attributes: cfgAttributes}
	}
}

func (m Model) View() string {
	attrList := m.list.View()
	selectionView := ""
	if m.selected != nil {
		selectionView = detailStyle.Height(m.list.Height()).Render(m.Attributes[*m.selected].View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, attrList, selectionView)
}

func (m *Model) selectAttr(i int) {
	m.selected = &i
	m.list.KeyMap.CursorDown.SetEnabled(false)
	m.list.KeyMap.CursorUp.SetEnabled(false)
	m.detailKeyMap.SelectNextField.SetEnabled(true)
	m.detailKeyMap.SelectPrevField.SetEnabled(true)
}

func (m *Model) deselect() {
	m.selected = nil
	m.list.KeyMap.CursorDown.SetEnabled(true)
	m.list.KeyMap.CursorUp.SetEnabled(true)
	m.detailKeyMap.SelectNextField.SetEnabled(false)
	m.detailKeyMap.SelectPrevField.SetEnabled(false)
}

type itemDelegate struct{ keys []key.Binding }

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	attr, ok := listItem.(*Attribute)
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
func (d itemDelegate) ShortHelp() []key.Binding { return d.keys }
func (d itemDelegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{d.ShortHelp()}
}
