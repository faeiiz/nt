package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/you/nt/storage"
)

type noteItem struct {
	n storage.Note
}

const (
	defaultWidth  = 70
	defaultHeight = 22
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE")).
			Padding(0, 1)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCDCDC")).
			Padding(0, 2)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7B68EE")).
			Padding(1)

	completedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#32CD32"))
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

// Modes
type mode int

const (
	listMode mode = iota
	viewMode
	addMode
	editMode
)

func (it noteItem) Title() string {
	status := "❌"
	if it.n.Completed {
		status = "✅"
	}
	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(it.n.CreatedAt.Format("02 Jan 15:04"))

	title := it.n.Title
	if it.n.Completed {
		title = completedStyle.Render(title)
	}
	return fmt.Sprintf("%s %s %s", title, status, ts)
}

func (it noteItem) Description() string { return it.n.Body }
func (it noteItem) FilterValue() string { return it.n.Title }

// softBreakLongTokens inserts zero-width spaces into long tokens to allow wrapping.
func softBreakLongTokens(s string, maxTokenLen int) string {
	if maxTokenLen <= 1 {
		return s
	}
	var out []rune
	tokenLen := 0
	for _, r := range s {
		out = append(out, r)
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			tokenLen = 0
			continue
		}
		tokenLen++
		if tokenLen >= maxTokenLen {
			out = append(out, '\u200B')
			tokenLen = 0
		}
	}
	return string(out)
}

// Model
type model struct {
	list       list.Model
	ctx        context.Context
	state      mode
	selected   noteItem
	titleInput textinput.Model
	bodyArea   textarea.Model
	bodyReady  bool // <-- track initialization
	inputStage int
	width      int
	height     int
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listWidth := msg.Width - 6
		if listWidth < 10 {
			listWidth = 10
		}
		listHeight := msg.Height - 8
		if listHeight < 3 {
			listHeight = 3
		}
		m.list.SetSize(listWidth, listHeight)
		if m.bodyReady {
			m.bodyArea.SetWidth(listWidth - 4)
		}
		m.titleInput.Width = listWidth - 6
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "a":
			if m.state == listMode {
				ti := textinput.New()
				ti.Placeholder = "Title"
				ti.Focus()
				ti.CursorEnd()
				ti.CharLimit = 0

				ta := textarea.New()
				ta.SetWidth(m.width - 10)
				ta.Placeholder = "Body (multiline). Use Enter to insert newline. Press Ctrl+S to save."
				ta.Blur()

				m.titleInput = ti
				m.bodyArea = ta
				m.bodyReady = true
				m.inputStage = 0
				m.state = addMode
				m.selected = noteItem{}
				return m, nil
			}

		case "enter":
			switch m.state {
			case listMode:
				if sel := m.list.SelectedItem(); sel != nil {
					m.selected = sel.(noteItem)
					m.state = viewMode
				}
			case addMode:
				if m.inputStage == 0 {
					m.selected.n.Title = m.titleInput.Value()
					m.titleInput.SetValue("")
					m.inputStage = 1
					if m.bodyReady {
						m.bodyArea.Focus()
					}
					return m, nil
				}
			case editMode:
				// handled by textarea
			}

		case "ctrl+s":
			if (m.state == addMode && m.inputStage == 1) || m.state == editMode {
				m.selected.n.Body = m.bodyArea.Value()
				if m.selected.n.Title == "" {
					lines := strings.Split(m.bodyArea.Value(), "\n")
					if len(lines) > 0 {
						m.selected.n.Title = lines[0]
					}
				}
				if m.selected.n.Title != "" || m.selected.n.Body != "" {
					if m.state == addMode {
						_, _ = store.Add(m.selected.n.Title, m.selected.n.Body)
					} else {
						_ = store.Update(&m.selected.n)
					}
					refreshList(m)
				}
				m.state = listMode
				m.inputStage = 0
				return m, nil
			}

		case "esc":
			if m.state == addMode {
				if m.inputStage == 1 {
					m.inputStage = 0
					m.titleInput.SetValue(m.selected.n.Title)
					m.titleInput.Focus()
					if m.bodyReady {
						m.bodyArea.Blur()
					}
					return m, nil
				}
				m.state = listMode
				m.inputStage = 0
				return m, nil
			} else if m.state == editMode {
				m.state = viewMode
				if m.bodyReady {
					m.bodyArea.Blur()
				}
				return m, nil
			} else if m.state == viewMode {
				m.state = listMode
				return m, nil
			}

		case "c":
			toggleComplete(m)
			return m, nil

		case "e":
			if m.state == viewMode && m.bodyReady {
				m.state = editMode
				ta := textarea.New()
				ta.SetWidth(m.width - 10)
				ta.SetValue(m.selected.n.Body)
				ta.Focus()
				m.bodyArea = ta
				m.bodyReady = true
				return m, nil
			}

		case "d":
			if m.state == viewMode {
				_ = store.Delete(m.selected.n.ID)
				refreshList(m)
				m.state = listMode
				return m, nil
			}
		}
	}

	// Route keys to widgets
	if m.state == listMode {
		m.list, cmd = m.list.Update(msg)
	} else if m.state == addMode || m.state == editMode {
		if m.state == addMode && m.inputStage == 0 {
			m.titleInput, cmd = m.titleInput.Update(msg)
		} else if m.bodyReady {
			m.bodyArea, cmd = m.bodyArea.Update(msg)
		}
	}

	return m, cmd
}

func toggleComplete(m *model) {
	switch m.state {
	case viewMode:
		m.selected.n.Completed = !m.selected.n.Completed
		_ = store.Update(&m.selected.n)
		refreshList(m)
	case listMode:
		if sel := m.list.SelectedItem(); sel != nil {
			note := sel.(noteItem)
			note.n.Completed = !note.n.Completed
			_ = store.Update(&note.n)
			refreshList(m)
		}
	}
}

func refreshList(m *model) {
	notes, _ := store.List()
	var items []list.Item
	for _, n := range notes {
		items = append(items, noteItem{n: *n})
	}
	m.list.SetItems(items)
}

func (m *model) View() string {
	switch m.state {
	case listMode:
		content := m.list.View()
		count := fmt.Sprintf(" %d notes ", len(m.list.Items()))
		status := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf("(↑/k up • ↓/j down • / filter) • %s", count))
		footer := helpStyle.Render("(a) add • (enter) view • (c) toggle complete • / filter")
		return borderStyle.Render(fmt.Sprintf("%s\n\n%s", content, status+"\n"+footer))

	case viewMode:
		contentWidth := m.width - 8
		if contentWidth < 10 {
			contentWidth = 10
		}

		titleRendered := titleStyle.Copy().Width(contentWidth).Render(noteItem{n: m.selected.n}.Title())

		bodyText := softBreakLongTokens(m.selected.n.Body, 30)
		bodyWrapped := bodyStyle.Copy().Width(contentWidth).Render(bodyText)
		footer := helpStyle.Render("(e) edit • (d) delete • (c) toggle complete • (Esc) back ")
		return borderStyle.Render(fmt.Sprintf("%s\n\n%s\n\n%s", titleRendered, bodyWrapped, footer))

	case addMode:
		if m.inputStage == 0 {
			stage := lipgloss.NewStyle().Bold(true).Render("Add Note — Title")
			help := helpStyle.Render("(Enter) go to Body • (Esc) cancel")
			content := fmt.Sprintf("%s\n\n%s\n\n%s", stage, m.titleInput.View(), help)
			return borderStyle.Render(content)
		}
		stage := lipgloss.NewStyle().Bold(true).Render("Add Note — Body (Ctrl+S to save)")
		help := helpStyle.Render("(Ctrl+S) save • (Esc) back to Title ")
		content := fmt.Sprintf("%s\n\n%s\n\n%s", stage, m.bodyArea.View(), help)
		return borderStyle.Render(content)

	case editMode:
		stage := lipgloss.NewStyle().Bold(true).Render("Edit Note — Body (Ctrl+S to save)")
		help := helpStyle.Render("(Ctrl+S) save • (Esc) cancel")
		content := fmt.Sprintf("%s\n\n%s\n\n%s", stage, m.bodyArea.View(), help)
		return borderStyle.Render(content)
	}

	return ""
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI to browse notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := store.List()
		if err != nil {
			return err
		}
		sort.Slice(notes, func(i, j int) bool { return notes[i].CreatedAt.After(notes[j].CreatedAt) })
		var items []list.Item
		for _, n := range notes {
			items = append(items, noteItem{n: *n})
		}

		l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
		l.Title = "Notes"
		l.SetShowStatusBar(false)
		l.SetShowHelp(false)
		l.SetFilteringEnabled(true)
		l.Paginator.InactiveDot = "-"
		l.Paginator.ActiveDot = "•"
		l.SetSize(defaultWidth-4, defaultHeight-6)

		p := tea.NewProgram(&model{
			list:   l,
			ctx:    context.Background(),
			width:  defaultWidth,
			height: defaultHeight,
		}, tea.WithAltScreen())

		return p.Start()
	},
}
