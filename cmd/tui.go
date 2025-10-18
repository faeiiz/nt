package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/you/nt/storage"
)

// wrapper type — do NOT use a type alias to storage.Note
type noteItem struct {
	n storage.Note
}

func (it noteItem) Title() string {
	status := "❌"
	if it.n.Completed {
		status = "✅"
	}

	ts := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true).
		Render(it.n.CreatedAt.Format("02 Jan 15:04"))

	return fmt.Sprintf("%s %s %s", it.n.Title, status, ts)
}

func (it noteItem) Description() string { return it.n.Body }
func (it noteItem) FilterValue() string { return it.n.Title }

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE")). // soft purple
			Padding(0, 1)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#DCDCDC")). // light gray
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#4B0082")). // dark indigo
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7B68EE")).
			Padding(1)
)

type mode int

const (
	listMode mode = iota
	viewMode
	addMode
	editMode
)

type model struct {
	list       list.Model
	ctx        context.Context
	state      mode
	selected   noteItem
	input      textinput.Model
	inputStage int // 0 = title, 1 = body (for adding/editing)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "c":
			if m.state == viewMode {
				m.selected.n.Completed = !m.selected.n.Completed
				_ = store.Update(&m.selected.n)
				// refresh list items
				notes, _ := store.List()
				var items []list.Item
				for _, n := range notes {
					items = append(items, noteItem{n: *n})
				}
				m.list.SetItems(items)
			} else if m.state == listMode {
				sel := m.list.SelectedItem()
				if sel != nil {
					note := sel.(noteItem)
					note.n.Completed = !note.n.Completed
					_ = store.Update(&note.n)
					// refresh list items
					notes, _ := store.List()

					var items []list.Item
					for _, n := range notes {
						items = append(items, noteItem{n: *n})
					}
					m.list.SetItems(items)
				}
			}

		case "enter":
			switch m.state {
			case listMode:
				if sel := m.list.SelectedItem(); sel != nil {
					m.selected = sel.(noteItem)
					m.state = viewMode
				}
			case addMode, editMode:
				if m.inputStage == 0 {
					// Title entered, move to Body
					m.selected.n.Title = m.input.Value()
					m.input.SetValue("")
					m.input.Placeholder = "Body"
					m.inputStage = 1
				} else {
					// Body entered, save note
					m.selected.n.Body = m.input.Value()
					if m.state == addMode {
						n, err := store.Add(m.selected.n.Title, m.selected.n.Body)
						if err != nil {
							// handle error
						}
						m.selected.n = *n
					} else {
						_ = store.Update(&m.selected.n)
					}

					// Refresh list
					notes, _ := store.List()
					var items []list.Item
					for _, n := range notes {
						items = append(items, noteItem{n: *n})
					}
					m.list.SetItems(items)

					m.state = listMode
					m.inputStage = 0
				}
			}

		case "a":
			if m.state == listMode {
				m.state = addMode
				m.input = textinput.New()
				m.input.Placeholder = "Title"
				m.input.SetValue("")
				m.input.Focus()
				m.input.CursorEnd()
				m.inputStage = 0
				m.selected = noteItem{}
				return m, nil // consume "a" key
			}

		case "e":
			if m.state == viewMode {
				m.state = editMode
				m.input = textinput.New()
				m.input.Placeholder = "Body"
				m.input.SetValue(m.selected.n.Body)
				m.input.Focus()
				m.input.CursorEnd()
				m.inputStage = 1
				return m, nil // consume "e" key
			}

		case "d":
			if m.state == viewMode {
				_ = store.Delete(m.selected.n.ID)
				// refresh list
				notes, _ := store.List()
				var items []list.Item
				for _, n := range notes {
					items = append(items, noteItem{n: *n})
				}
				m.list.SetItems(items)
				m.state = listMode
			}

		case "esc", "backspace":
			if m.state == addMode || m.state == editMode {
				if m.inputStage == 1 {
					// Body stage
					if m.input.Value() == "" {
						// Body empty → go back to Title stage
						m.inputStage = 0
						m.input.SetValue(m.selected.n.Title)
						m.input.Placeholder = "Title"
						m.input.CursorEnd()
						return m, nil // consume key
					}
					// Body not empty → allow backspace normally
				} else {
					// Title stage
					if m.input.Value() == "" {
						// Empty Title → cancel to list
						m.state = listMode
						m.inputStage = 0
						return m, nil
					}
					// Title not empty → allow backspace normally
				}
			} else if m.state == viewMode {
				m.state = listMode
				m.inputStage = 0
				return m, nil
			}
		}

	}

	// Let list handle remaining messages
	if m.state == listMode {
		m.list, cmd = m.list.Update(msg)
	} else if m.state == addMode || m.state == editMode {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.state {
	case listMode:
		return borderStyle.Render(m.list.View() + "\n(a) add • (enter) view • q quit")
		// return m.list.View()
	case viewMode:
		content := fmt.Sprintf("%s\n\n%s\n\n(e) edit • (d) delete • (backspace) back • q quit",
			titleStyle.Render(m.selected.n.Title),
			bodyStyle.Render(m.selected.n.Body),
		)

		return borderStyle.Render(content)

	case addMode, editMode:
		stage := "Title"
		if m.inputStage == 1 {
			stage = "Body"
		}
		content := fmt.Sprintf("Add/Edit Note - %s:\n\n%s\n\n(Enter) next/save • (Esc) cancel",
			stage,
			m.input.View(),
		)
		return borderStyle.Render(content)

	}
	return ""
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "launch interactive TUI to browse notes",
	RunE: func(cmd *cobra.Command, args []string) error {
		notes, err := store.List()
		// sort the noted by CreatedAt descending
		sort.Slice(notes, func(i, j int) bool { return notes[i].CreatedAt.After(notes[j].CreatedAt) })
		if err != nil {
			return err
		}
		var items []list.Item
		for _, n := range notes {
			items = append(items, noteItem{n: *n})
		}

		const defaultWidth = 60
		const defaultHeight = 20

		l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
		l.Title = "Notes"
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(true)
		l.Paginator.InactiveDot = "-"
		l.Paginator.ActiveDot = "•"

		p := tea.NewProgram(model{list: l, ctx: context.Background()})
		if err := p.Start(); err != nil {
			return err
		}
		return nil
	},
}
