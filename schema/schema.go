package schema

import (
	"fmt"
	"io"
	"log"
	"strconv"

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
	Name       string
	nameInput  textinput.Model
	Width      int
	widthInput textinput.Model
	Selectors  []string
	Type       string
	Format     *string
}

func (a Attribute) FilterValue() string { return "" }

func (a Attribute) View() string {
	log.Println("PrintAttr", a.nameInput.Value(), a.nameInput.View())
	return "Name\n" + a.nameInput.View() + "\n\n" + "Width\n" + a.widthInput.View()
}

type Model struct {
	Attributes []Attribute
	list       list.Model
	selected   *Attribute
}

func FromLogView(lv config.LogView, width, height int) Model {
	attrs := make([]Attribute, len(lv.Attrs))
	for i, attr := range lv.Attrs {
		nameInput := textinput.New()
		nameInput.Placeholder = "Title/Name/Barcode/Host/..."
		nameInput.SetValue(attr.Name)
		nameInput.CharLimit = 30
		nameInput.Focus()
		widthInput := textinput.New()
		widthInput.Placeholder = "15"
		widthInput.SetValue(strconv.Itoa(attr.Width))
		attrs[i] = Attribute{
			Name:       attr.Name,
			nameInput:  nameInput,
			Width:      attr.Width,
			widthInput: widthInput,
			Selectors:  []string{attr.Selector},
			Type:       attr.Type,
			Format:     attr.Format,
		}
	}
	items := make([]list.Item, len(attrs))
	for i, attr := range attrs {
		items[i] = attr
	}

	log.Println("Create schema")
	l := list.New(items, itemDelegate{}, width, height)
	l.Title = "Schema attributes"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle
	l.DisableQuitKeybindings()
	return Model{Attributes: attrs, list: l}
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
		switch keypress := msg.String(); keypress {
		case "enter":
			attr, ok := m.list.SelectedItem().(Attribute)
			if ok {
				m.selected = &attr
				m.list.KeyMap.CursorDown.SetEnabled(false)
				m.list.KeyMap.CursorUp.SetEnabled(false)
			}
			return m, nil
		case "esc":
			if m.selected == nil {
				return m, func() tea.Msg { return Close{} }
			}
			m.selected = nil
			m.list.KeyMap.CursorDown.SetEnabled(true)
			m.list.KeyMap.CursorUp.SetEnabled(true)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
	if m.selected != nil {
		log.Println("Selected", m.selected.nameInput.Value())
		m.selected.nameInput, cmd = m.selected.nameInput.Update(msg)
		log.Println("Selected2", m.selected.nameInput.Value())
		cmds = append(cmds, cmd)
		m.selected.widthInput, cmd = m.selected.widthInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.selected != nil {
		log.Println("View", m.selected.nameInput.Value())
	}
	attrList := m.list.View()
	selectionView := ""
	if m.selected != nil {
		selectionView = detailStyle.Height(m.list.Height()).Render(m.selected.View())
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, attrList, selectionView)
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
