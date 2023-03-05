package table

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Model[E] defines a state for the table widget.
type Model[E any] struct {
	KeyMap KeyMap

	cols    []ColumnSpec[E]
	rows    []E
	cursor  int
	yOffset int
	hcursor int
	focus   bool
	styles  Styles

	viewport viewport.Model
}

// Row represents one line in the table.
// type Row string

type ColumnSpec[E any] interface {
	Title() string
	Width() int
	SetWidth(int)
	GetValue(E) string
}

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu menu.
type KeyMap struct {
	LineUp       key.Binding
	LineDown     key.Binding
	LineLeft     key.Binding
	LineRight    key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	HalfPageUp   key.Binding
	HalfPageDown key.Binding
	GotoTop      key.Binding
	GotoBottom   key.Binding
	ShrinkColumn key.Binding
	GrowColumn   key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		LineLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "left"),
		),
		LineRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "right"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b", "pgup"),
			key.WithHelp("ctrl+b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f", "pgdown"),
			key.WithHelp("ctrl+f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		ShrinkColumn: key.NewBinding(
			key.WithKeys("ctrl+left", "ctrl+h"),
			key.WithHelp("ctrl+←/ctrl+h", "shrink column"),
		),
		GrowColumn: key.NewBinding(
			key.WithKeys("ctrl+right", "ctrl+l"),
			key.WithHelp("ctrl+→/ctrl+l", "grow column"),
		),
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Header   lipgloss.Style
	Cell     lipgloss.Style
	Selected lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:   lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

// SetStyles sets the table styles.
func (m *Model[E]) SetStyles(s Styles) {
	m.styles = s
	m.UpdateViewport()
}

// Option[E] is used to set options in New. For example:
//
//	table := New(WithColumns([]Column{{Title: "ID", Width: 10}}))
type Option[E any] func(*Model[E])

// New creates a new model for the table widget.
func New[E any](opts ...Option[E]) Model[E] {
	m := Model[E]{
		cursor:   0,
		viewport: viewport.New(0, 20),

		KeyMap: DefaultKeyMap(),
		styles: DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	m.UpdateViewport()

	return m
}

// WithColumns sets the table columns (headers).
func WithColumns[E any](cols []ColumnSpec[E]) Option[E] {
	return func(m *Model[E]) {
		m.cols = cols
	}
}

// WithRows sets the table rows (data).
func WithRows[E any](rows []E) Option[E] {
	return func(m *Model[E]) {
		m.rows = rows
	}
}

// WithHeight sets the height of the table.
func WithHeight[E any](h int) Option[E] {
	return func(m *Model[E]) {
		m.viewport.Height = h
	}
}

// WithWidth sets the width of the table.
func WithWidth[E any](w int) Option[E] {
	return func(m *Model[E]) {
		m.viewport.Width = w
	}
}

// WithFocused sets the focus state of the table.
func WithFocused[E any](f bool) Option[E] {
	return func(m *Model[E]) {
		m.focus = f
	}
}

// WithStyles sets the table styles.
func WithStyles[E any](s Styles) Option[E] {
	return func(m *Model[E]) {
		m.styles = s
	}
}

// WithKeyMap sets the key map.
func WithKeyMap[E any](km KeyMap) Option[E] {
	return func(m *Model[E]) {
		m.KeyMap = km
	}
}

// Update is the Bubble Tea update loop.
func (m Model[E]) Update(msg tea.Msg) (Model[E], tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.LineLeft):
			m.MoveLeft(1)
		case key.Matches(msg, m.KeyMap.LineRight):
			m.MoveRight(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.viewport.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.viewport.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.KeyMap.ShrinkColumn):
			m.ShrinkColumn()
		case key.Matches(msg, m.KeyMap.GrowColumn):
			m.GrowColumn()
		}
	case tea.WindowSizeMsg:
		m.SetWidth(msg.Width - 2)
		m.SetHeight(msg.Height - 4)
	}

	return m, tea.Batch(cmds...)
}

// Focused returns the focus state of the table.
func (m Model[E]) Focused() bool {
	return m.focus
}

// Focus focusses the table, allowing the user to move around the rows and
// interact.
func (m *Model[E]) Focus() {
	m.focus = true
	m.UpdateViewport()
}

// Blur blurs the table, preventing selection or movement.
func (m *Model[E]) Blur() {
	m.focus = false
	m.UpdateViewport()
}

// View renders the component.
func (m Model[E]) View() string {
	return m.headersView() + "\n" + m.viewport.View()
}

// UpdateViewport updates the list content based on the previously defined
// columns and rows.
func (m *Model[E]) UpdateViewport() {
	renderedRows := make([]string, 0, m.viewport.Height)
	start, end := m.yOffset, min(m.yOffset+m.viewport.Height, len(m.rows))
	for i := range m.rows[start:end] {
		renderedRows = append(renderedRows, m.renderRow(start+i))
	}

	m.viewport.SetContent(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

// SelectedRow returns the selected row.
// You can cast it to your own implementation.
func (m Model[E]) SelectedRow() E {
	return m.rows[m.cursor]
}

// SetColumns sets the table columns (headers).
func (m *Model[E]) SetColumns(cols []ColumnSpec[E]) {
	m.cols = cols
	m.UpdateViewport()
}

// SetRows set a new rows state.
func (m *Model[E]) SetRows(r []E) {
	m.rows = r
	m.UpdateViewport()
}

// SetWidth sets the width of the viewport of the table.
func (m *Model[E]) SetWidth(w int) {
	m.viewport.Width = w
	m.UpdateViewport()
}

// SetHeight sets the height of the viewport of the table.
func (m *Model[E]) SetHeight(h int) {
	m.viewport.Height = h
	m.UpdateViewport()
}

// Height returns the viewport height of the table.
func (m Model[E]) Height() int {
	return m.viewport.Height
}

// Width returns the viewport width of the table.
func (m Model[E]) Width() int {
	return m.viewport.Width
}

// Cursor returns the index of the selected row.
func (m Model[E]) Cursor() int {
	return m.cursor
}

// SetCursor sets the cursor position in the table.
func (m *Model[E]) SetCursor(n int) {
	m.cursor = clamp(n, 0, len(m.rows)-1)
	m.UpdateViewport()
}

// MoveUp moves the selection up by any number of row.
// It can not go above the first row.
func (m *Model[E]) MoveUp(n int) {
	m.cursor = clamp(m.cursor-n, 0, len(m.rows)-1)

	if m.cursor < m.yOffset {
		m.yOffset = m.cursor
	}
	m.UpdateViewport()
}

// MoveDown moves the selection down by any number of row.
// It can not go below the last row.
func (m *Model[E]) MoveDown(n int) {
	m.cursor = clamp(m.cursor+n, 0, len(m.rows)-1)

	if m.cursor > (m.yOffset + (m.viewport.Height - 1)) {
		m.yOffset = m.cursor - (m.viewport.Height - 1)
	}
	m.UpdateViewport()
}

// MoveLeft moves the selection left by any number of columns.
// It can not go left of the first column.
func (m *Model[E]) MoveLeft(n int) {
	m.hcursor = clamp(m.hcursor-n, 0, len(m.cols)-1)
	m.UpdateViewport()
}

// MoveRight moves the selection right by any number of columns.
// It can not go right of the last column.
func (m *Model[E]) MoveRight(n int) {
	m.hcursor = clamp(m.hcursor+n, 0, len(m.cols)-1)
	m.UpdateViewport()
}

// GotoTop moves the selection to the first row.
func (m *Model[E]) GotoTop() {
	m.MoveUp(m.cursor)
}

// GotoBottom moves the selection to the last row.
func (m *Model[E]) GotoBottom() {
	m.MoveDown(len(m.rows))
}

// ShrinkColumn shrinks the current column by one character.
func (m *Model[E]) ShrinkColumn() {
	m.cols[m.hcursor].SetWidth(m.cols[m.hcursor].Width() - 1)
	m.UpdateViewport()
}

// GrowColumn grows the current column by one character.
func (m *Model[E]) GrowColumn() {
	m.cols[m.hcursor].SetWidth(m.cols[m.hcursor].Width() + 1)
	m.UpdateViewport()
}

func (m Model[E]) headersView() string {
	s := make([]string, 0, len(m.cols))
	remainingWidth := m.Width()
	padding := m.styles.Header.GetHorizontalPadding()
	for i, col := range m.cols {
		if i < m.hcursor {
			continue
		}
		colWidthWithPadding := min(col.Width()+padding, remainingWidth)
		remainingWidth -= colWidthWithPadding
		colWidth := colWidthWithPadding - padding
		if colWidth < 1 {
			continue
		}
		style := lipgloss.NewStyle().Width(colWidth).MaxWidth(colWidth).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title(), colWidth, "…"))
		s = append(s, m.styles.Header.Render(renderedCell))
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, s...)
}

func (m *Model[E]) renderRow(rowID int) string {
	s := make([]string, 0, len(m.cols))
	remainingWidth := m.Width()
	padding := m.styles.Cell.GetHorizontalPadding()
	for i, col := range m.cols {
		if i < m.hcursor {
			continue
		}
		row := m.rows[rowID]
		colWidthWithPadding := min(col.Width()+padding, remainingWidth)
		remainingWidth -= colWidthWithPadding
		colWidth := colWidthWithPadding - padding
		if colWidth < 1 {
			continue
		}
		style := lipgloss.NewStyle().
			Width(colWidth).
			MaxWidth(colWidth).
			Inline(true)
		value := col.GetValue(row)
		content := style.Render(runewidth.Truncate(value, colWidth, "…"))
		renderedCell := m.styles.Cell.Render(content)
		s = append(s, renderedCell)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, s...)

	if rowID == m.cursor {
		return m.styles.Selected.Render(row)
	}

	return row
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
