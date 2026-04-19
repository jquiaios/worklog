package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jquiaios/worklog/internal/db"
	"github.com/jquiaios/worklog/internal/entry"
)

// ── styles ────────────────────────────────────────────────────────────────────

var (
	highlightColor = lipgloss.Color("42")
	lowlightColor  = lipgloss.Color("203")

	colColors = [2]lipgloss.Color{highlightColor, lowlightColor}
	colTitles = [2]string{"Highlights", "Lowlights"}
	colTypes  = [2]entry.Type{entry.Highlight, entry.Lowlight}

	blurredBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))

	periodLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("255"))
	periodNavStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	confirmStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Padding(0, 1)
	errStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Padding(0, 1)
	keyStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
)

// ── period ────────────────────────────────────────────────────────────────────

type period struct {
	year    int
	quarter int // 1–4; 0 = all time
}

func currentPeriod() period {
	now := time.Now()
	return period{year: now.Year(), quarter: (int(now.Month())-1)/3 + 1}
}

func (p period) label() string {
	if p.quarter == 0 {
		return "All time"
	}
	return fmt.Sprintf("Q%d %d", p.quarter, p.year)
}

func (p period) bounds() (from, to time.Time) {
	if p.quarter == 0 {
		return time.Time{}, time.Time{}
	}
	month := time.Month((p.quarter-1)*3 + 1)
	from = time.Date(p.year, month, 1, 0, 0, 0, 0, time.Local)
	to = from.AddDate(0, 3, 0)
	return
}

func (p period) prev() period {
	if p.quarter == 0 {
		return p
	}
	if p.quarter == 1 {
		return period{year: p.year - 1, quarter: 4}
	}
	return period{year: p.year, quarter: p.quarter - 1}
}

func (p period) next() period {
	if p.quarter == 0 {
		return currentPeriod()
	}
	if p.quarter == 4 {
		return period{year: p.year + 1, quarter: 1}
	}
	return period{year: p.year, quarter: p.quarter + 1}
}

// ── list item ─────────────────────────────────────────────────────────────────

type item struct {
	e entry.Entry
}

func (i item) Title() string       { return i.e.Body }
func (i item) Description() string { return fmt.Sprintf("#%d · %s", i.e.ID, i.e.CreatedAt.Local().Format("2006-01-02")) }
func (i item) FilterValue() string { return i.e.Body }

// ── model ─────────────────────────────────────────────────────────────────────

type col int

const (
	colHighlight col = iota
	colLowlight
)

type viewState int

const (
	stateNormal viewState = iota
	stateConfirmDelete
	stateEditing
	stateAdding
)

type Model struct {
	cols        [2]list.Model
	focused     col
	period      period
	store       *db.DB
	width       int
	height      int
	err         error
	state       viewState
	pendingItem item
	textInput   textinput.Model
}

func New(store *db.DB) (Model, error) {
	hlDelegate := list.NewDefaultDelegate()
	hlDelegate.Styles.SelectedTitle = hlDelegate.Styles.SelectedTitle.Foreground(highlightColor).BorderForeground(highlightColor)
	hlDelegate.Styles.SelectedDesc = hlDelegate.Styles.SelectedDesc.Foreground(highlightColor)

	llDelegate := list.NewDefaultDelegate()
	llDelegate.Styles.SelectedTitle = llDelegate.Styles.SelectedTitle.Foreground(lowlightColor).BorderForeground(lowlightColor)
	llDelegate.Styles.SelectedDesc = llDelegate.Styles.SelectedDesc.Foreground(lowlightColor)

	hlList := list.New(nil, hlDelegate, 0, 0)
	hlList.SetShowHelp(false)

	llList := list.New(nil, llDelegate, 0, 0)
	llList.SetShowHelp(false)

	ti := textinput.New()
	ti.CharLimit = 500
	ti.Prompt = " "

	m := Model{
		cols:      [2]list.Model{hlList, llList},
		focused:   colHighlight,
		period:    currentPeriod(),
		store:     store,
		textInput: ti,
	}
	m.updateFocusStyles()
	if err := m.loadPeriod(); err != nil {
		return Model{}, err
	}
	return m, nil
}

func (m *Model) loadPeriod() error {
	from, to := m.period.bounds()
	for _, c := range []col{colHighlight, colLowlight} {
		entries, err := m.store.List(string(colTypes[c]), from, to)
		if err != nil {
			return err
		}
		m.cols[c].SetItems(toItems(entries))
	}
	return nil
}

func toItems(entries []entry.Entry) []list.Item {
	items := make([]list.Item, len(entries))
	for i, e := range entries {
		items[i] = item{e}
	}
	return items
}

func (m *Model) updateFocusStyles() {
	for i := range m.cols {
		focused := col(i) == m.focused
		titleStyle := lipgloss.NewStyle().Foreground(colColors[i])
		if focused {
			m.cols[i].Title = "▶ " + colTitles[i]
			titleStyle = titleStyle.Bold(true)
		} else {
			m.cols[i].Title = "  " + colTitles[i]
			titleStyle = titleStyle.Faint(true)
		}
		m.cols[i].Styles.Title = titleStyle
	}
}

// ── update ────────────────────────────────────────────────────────────────────

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.setSizes()
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateConfirmDelete:
			if msg.String() == "y" {
				if _, err := m.store.Delete(m.pendingItem.e.ID); err != nil {
					m.err = err
				} else {
					m.err = nil
					m.cols[m.focused].RemoveItem(m.cols[m.focused].Index())
				}
			}
			m.state = stateNormal
			return m, nil

		case stateAdding:
			switch msg.String() {
			case "enter":
				body := strings.TrimSpace(m.textInput.Value())
				if body != "" {
					e := entry.Entry{Type: colTypes[m.focused], Body: body, CreatedAt: time.Now()}
					if _, err := m.store.Insert(e); err != nil {
						m.err = err
					} else {
						m.err = nil
						if err := m.loadPeriod(); err != nil {
							m.err = err
						}
					}
				}
				m.state = stateNormal
				m.textInput.Blur()
				return m, nil
			case "esc":
				m.state = stateNormal
				m.textInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		case stateEditing:
			switch msg.String() {
			case "enter":
				newBody := strings.TrimSpace(m.textInput.Value())
				if newBody != "" && newBody != m.pendingItem.e.Body {
					if err := m.store.Update(m.pendingItem.e.ID, newBody); err != nil {
						m.err = err
					} else {
						m.err = nil
						updated := m.pendingItem
						updated.e.Body = newBody
						m.cols[m.focused].SetItem(m.cols[m.focused].Index(), updated)
					}
				}
				m.state = stateNormal
				m.textInput.Blur()
				return m, nil
			case "esc":
				m.state = stateNormal
				m.textInput.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		case stateNormal:
			if m.cols[m.focused].FilterState() == list.Filtering {
				break
			}
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "tab":
				m.focused = 1 - m.focused
				m.updateFocusStyles()
				return m, nil
			case "[":
				m.period = m.period.prev()
				m.err = nil
				if err := m.loadPeriod(); err != nil {
					m.err = err
				}
				return m, nil
			case "]":
				m.period = m.period.next()
				m.err = nil
				if err := m.loadPeriod(); err != nil {
					m.err = err
				}
				return m, nil
			case "a":
				m.period = period{quarter: 0}
				m.err = nil
				if err := m.loadPeriod(); err != nil {
					m.err = err
				}
				return m, nil
			case "d":
				sel, ok := m.cols[m.focused].SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.pendingItem = sel
				m.state = stateConfirmDelete
				return m, nil
			case "e":
				sel, ok := m.cols[m.focused].SelectedItem().(item)
				if !ok {
					return m, nil
				}
				m.pendingItem = sel
				m.textInput.SetValue(sel.e.Body)
				m.textInput.CursorEnd()
				m.textInput.Width = m.width - 6
				m.textInput.Focus()
				m.state = stateEditing
				return m, nil
			case "n":
				m.textInput.SetValue("")
				m.textInput.Width = m.width - 6
				m.textInput.Focus()
				m.state = stateAdding
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.cols[m.focused], cmd = m.cols[m.focused].Update(msg)
	return m, cmd
}

func (m *Model) setSizes() {
	colW := m.width/2 - 2
	colH := m.height - 4 // -1 header, -3 footer
	for i := range m.cols {
		m.cols[i].SetSize(colW, colH)
	}
	m.textInput.Width = m.width - 6
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m Model) headerView() string {
	nav := periodNavStyle.Render("◀") + "  " +
		periodLabelStyle.Render(m.period.label()) + "  " +
		periodNavStyle.Render("▶")
	hint := helpStyle.Render("[/]: period   a: all time")
	content := nav + "   " + hint
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(content)
}

func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}

	colW := m.width/2 - 2
	colH := m.height - 4

	renderCol := func(l list.Model, c col) string {
		s := blurredBorder
		if c == m.focused {
			s = blurredBorder.BorderForeground(colColors[c])
		}
		return s.Width(colW).Height(colH).Render(l.View())
	}

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		renderCol(m.cols[0], colHighlight),
		renderCol(m.cols[1], colLowlight),
	)

	var footer string
	switch m.state {
	case stateConfirmDelete:
		body := m.pendingItem.e.Body
		if len(body) > 50 {
			body = body[:50] + "…"
		}
		footer = confirmStyle.Render(fmt.Sprintf("Delete %q?", body)) +
			"  " + keyStyle.Render("y") + helpStyle.Render(" confirm") +
			"  " + keyStyle.Render("n / esc") + helpStyle.Render(" cancel")

	case stateAdding, stateEditing:
		label := "New " + string(colTypes[m.focused])
		if m.state == stateEditing {
			label = "Edit"
		}
		inputStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colColors[m.focused]).
			Padding(0, 1).
			Width(m.width - 6)
		footer = inputStyle.Render(label + ": " + m.textInput.View())

	default:
		footer = helpStyle.Render("tab: switch   n: new   d: delete   e: edit   /: filter   q: quit")
		if m.err != nil {
			footer += errStyle.Render("error: " + m.err.Error())
		}
	}

	return m.headerView() + "\n" + columns + "\n" + footer
}
