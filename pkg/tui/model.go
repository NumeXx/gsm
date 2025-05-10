package tui

import (
	"fmt"
	"strings"

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ChosenConnectionGlobal *Item

type Item struct {
	config.Connection
}

func (i Item) Title() string { return fmt.Sprintf("Name : %s", i.Name) }

func (i Item) Description() string {
	if len(i.Tags) == 0 {
		return "Tag  : -"
	}
	return fmt.Sprintf("Tag  : %s", strings.Join(i.Tags, ", "))
}

func (i Item) FilterValue() string { return i.Name + " " + strings.Join(i.Tags, " ") } // Filter berdasarkan Nama dan Tag

const (
	focusEditName = iota
	focusEditKey
	focusEditTags
)

const (
	EditingIndexAddNew = -2
)

type StatusMessageType int

const (
	StatusNone StatusMessageType = iota
	StatusSuccess
	StatusError
)

type Model struct {
	List            list.Model
	IsEditing       bool
	EditNameInput   textinput.Model
	EditKeyInput    textinput.Model
	EditTagsInput   textinput.Model
	EditingIndex    int
	EditFocusIndex  int
	lastKnownWidth  int
	lastKnownHeight int

	StatusMessage        string
	StatusType           StatusMessageType
	IsConfirmingDelete   bool
	DeleteIndex          int
	DeleteConnectionName string
}

func NewModel(cfg config.Config) Model {
	items := []list.Item{}
	for _, c := range cfg.Connections {
		items = append(items, Item{Connection: c})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetFilteringEnabled(true)
	l.FilterInput.Placeholder = "Filter by name or tag... (type to search)"
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.FilterInput.TextStyle = lipgloss.NewStyle()

	// Styling TUI
	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("23")).
		Foreground(lipgloss.Color("231")).
		Padding(0, 1).
		Bold(true)
	l.Title = "GSM | GSocket Manager"
	l.Styles.Title = titleStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("32")).
		Padding(0, 1).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Background(lipgloss.Color("32")).
		Padding(0, 1)
	l.SetDelegate(delegate)

	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("connection", "connections")
	l.Styles.StatusBar = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("235")).Foreground(lipgloss.Color("242"))
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	l.SetShowHelp(false)

	ni := textinput.New()
	ni.Placeholder = "Connection Name (required)"
	ni.CharLimit = 100
	ni.Width = 50

	ki := textinput.New()
	ki.Placeholder = "GSocket Key (required)"
	ki.CharLimit = 256
	ki.Width = 50
	ti := textinput.New()
	ti.Placeholder = "tag1,tag2 (optional)"
	ti.CharLimit = 200
	ti.Width = 50

	ChosenConnectionGlobal = nil

	return Model{
		List:                 l,
		IsEditing:            false,
		EditNameInput:        ni,
		EditKeyInput:         ki,
		EditTagsInput:        ti,
		EditingIndex:         -1,
		EditFocusIndex:       focusEditName,
		StatusMessage:        "",
		StatusType:           StatusNone,
		IsConfirmingDelete:   false,
		DeleteIndex:          -1,
		DeleteConnectionName: "",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.List.StartSpinner(),
		textinput.Blink,
		tea.ClearScreen,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.IsConfirmingDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch strings.ToLower(msg.String()) {
			case "y":
				if err := config.DeleteConnectionByIndex(m.DeleteIndex); err != nil {
					m.StatusMessage = fmt.Sprintf("Error deleting '%s': %v", m.DeleteConnectionName, err)
					m.StatusType = StatusError
				} else {
					m.StatusMessage = fmt.Sprintf("Connection '%s' deleted.", m.DeleteConnectionName)
					m.StatusType = StatusSuccess
				}
				m.IsConfirmingDelete = false
				m.DeleteIndex = -1
				m.DeleteConnectionName = ""
				if err := config.Load(); err != nil {
					m.StatusMessage = fmt.Sprintf("Error reloading config after delete: %v", err)
					m.StatusType = StatusError
					return m, nil
				}

				reloadedCfg := config.GetCurrent()
				newM := NewModel(reloadedCfg)
				newM.lastKnownWidth = m.lastKnownWidth
				newM.lastKnownHeight = m.lastKnownHeight

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 {
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, nil
			case "n", "esc", "ctrl+c":
				m.IsConfirmingDelete = false
				m.DeleteIndex = -1
				m.DeleteConnectionName = ""
				m.StatusMessage = "Delete cancelled."
				m.StatusType = StatusNone
				return m, nil
			}
		}
		return m, nil
	}

	if m.IsEditing {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.IsEditing = false
				m.EditingIndex = -1
				m.EditNameInput.Blur()
				m.EditKeyInput.Blur()
				m.EditTagsInput.Blur()
				m.StatusMessage = "Edit cancelled."
				m.StatusType = StatusNone
				return m, nil
			case tea.KeyTab, tea.KeyShiftTab:
				cmd = m.updateFocusEdit(msg.Type == tea.KeyTab)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case tea.KeyEnter:
				m.StatusMessage = ""
				m.StatusType = StatusNone

				newName := strings.TrimSpace(m.EditNameInput.Value())
				newKey := strings.TrimSpace(m.EditKeyInput.Value())
				newTagsRaw := strings.TrimSpace(m.EditTagsInput.Value())

				if newName == "" {
					m.StatusMessage = "Connection name cannot be empty!"
					m.StatusType = StatusError
					return m, m.EditNameInput.Focus()
				}
				if newKey == "" {
					m.StatusMessage = "GSocket key cannot be empty!"
					m.StatusType = StatusError
					return m, m.EditKeyInput.Focus()
				}

				var tags []string
				if newTagsRaw != "" {
					tagParts := strings.Split(newTagsRaw, ",")
					for _, t := range tagParts {
						tags = append(tags, strings.TrimSpace(t))
					}
				}

				var saveErr error
				if m.EditingIndex == EditingIndexAddNew {
					newConn := config.Connection{Name: newName, Key: newKey, Tags: tags, Usage: 0}
					config.AddConnection(newConn)
					saveErr = config.Save()
				} else {
					updatedConn := config.Connection{
						Name:  newName,
						Key:   newKey,
						Tags:  tags,
						Usage: config.GetCurrent().Connections[m.EditingIndex].Usage,
					}
					saveErr = config.UpdateConnectionByIndex(m.EditingIndex, updatedConn)
				}

				if saveErr != nil {
					m.StatusMessage = fmt.Sprintf("Error saving: %v", saveErr)
					m.StatusType = StatusError
					return m, nil
				}

				m.IsEditing = false
				m.EditingIndex = -1
				m.EditNameInput.Blur()
				m.EditKeyInput.Blur()
				m.EditTagsInput.Blur()

				if err := config.Load(); err != nil {
					m.StatusMessage = fmt.Sprintf("Error reloading config after save: %v", err)
					m.StatusType = StatusError
					m.IsEditing = true
					return m, m.EditNameInput.Focus()
				}

				newM := NewModel(config.GetCurrent())
				newM.lastKnownWidth = m.lastKnownWidth
				newM.lastKnownHeight = m.lastKnownHeight

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 {
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, nil
			}
		}

		switch m.EditFocusIndex {
		case focusEditName:
			m.EditNameInput, cmd = m.EditNameInput.Update(msg)
		case focusEditKey:
			m.EditKeyInput, cmd = m.EditKeyInput.Update(msg)
		case focusEditTags:
			m.EditTagsInput, cmd = m.EditTagsInput.Update(msg)
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastKnownWidth = msg.Width
		m.lastKnownHeight = msg.Height
		m.List.SetSize(m.lastKnownWidth, m.lastKnownHeight-1)
		return m, nil
	case tea.KeyMsg:
		if m.StatusMessage != "" && m.StatusType == StatusError {
			m.StatusMessage = ""
			m.StatusType = StatusNone
		}
		if m.List.FilterState() != list.Filtering {
			switch msg.String() {
			case "q", "ctrl+c":
				ChosenConnectionGlobal = nil
				return m, tea.Quit
			case "e":
				if len(m.List.VisibleItems()) > 0 {
					selected, ok := m.List.SelectedItem().(Item)
					if ok {
						m.IsEditing = true
						cfg := config.GetCurrent()
						foundIndex := -1
						for i, conn := range cfg.Connections {
							if conn.Name == selected.Name && conn.Key == selected.Key {
								foundIndex = i
								break
							}
						}
						if foundIndex != -1 {
							m.EditingIndex = foundIndex
						} else {
							m.IsEditing = false
							return m, nil
						}
						m.EditNameInput.SetValue(selected.Name)
						m.EditKeyInput.SetValue(selected.Key)
						m.EditTagsInput.SetValue(strings.Join(selected.Tags, ", "))
						m.EditFocusIndex = focusEditName
						m.StatusMessage = ""
						m.StatusType = StatusNone
						return m, m.EditNameInput.Focus()
					}
				}
			case "d":
				if len(m.List.VisibleItems()) > 0 {
					selected, ok := m.List.SelectedItem().(Item)
					if ok {
						cfg := config.GetCurrent()
						foundIndex := -1
						for i, connVal := range cfg.Connections {
							if connVal.Name == selected.Name && connVal.Key == selected.Key {
								foundIndex = i
								break
							}
						}
						if foundIndex != -1 {
							m.IsConfirmingDelete = true
							m.DeleteIndex = foundIndex
							m.DeleteConnectionName = selected.Name
							m.StatusMessage = ""
							m.StatusType = StatusNone
							return m, nil
						}
					}
				}
			case "a":
				m.IsEditing = true
				m.EditingIndex = EditingIndexAddNew
				m.EditNameInput.SetValue("")
				m.EditKeyInput.SetValue("")
				m.EditTagsInput.SetValue("")
				m.EditFocusIndex = focusEditName
				m.StatusMessage = ""
				m.StatusType = StatusNone
				return m, m.EditNameInput.Focus()
			}
		}
		if msg.Type == tea.KeyEnter && !m.IsEditing {
			selected, ok := m.List.SelectedItem().(Item)
			if ok {
				ChosenConnectionGlobal = &selected
				return m, tea.Quit
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var b strings.Builder

	if m.StatusMessage != "" && m.StatusType == StatusError {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1).Bold(true)
		b.WriteString(statusStyle.Render(m.StatusMessage) + "\n\n")
	}

	if m.IsConfirmingDelete {
		headerStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1).Foreground(lipgloss.Color("196"))
		b.WriteString(headerStyle.Render(fmt.Sprintf("DELETE Connection: %s?", m.DeleteConnectionName)) + "\n\n")
		promptStyle := lipgloss.NewStyle().MarginBottom(1)
		b.WriteString(promptStyle.Render(fmt.Sprintf("Are you sure you want to delete '%s'?", m.DeleteConnectionName)) + "\n")
		b.WriteString(promptStyle.Render("This action cannot be undone.") + "\n\n")
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		b.WriteString(hintStyle.Render("(Y)es, delete it! / (N)o or (Esc) to cancel."))
		return b.String()
	}

	if m.IsEditing {
		headerStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
		var formTitle string
		if m.EditingIndex == EditingIndexAddNew {
			formTitle = "Add New Connection (Esc to Cancel)"
		} else if m.EditingIndex >= 0 && m.EditingIndex < len(config.GetCurrent().Connections) {
			formTitle = fmt.Sprintf("Editing Connection: %s (Esc to Cancel)", config.GetCurrent().Connections[m.EditingIndex].Name)
		} else {
			formTitle = "Edit Connection (Esc to Cancel)"
		}
		b.WriteString(headerStyle.Render(formTitle) + "\n")

		inputStyle := lipgloss.NewStyle().MarginBottom(1)
		b.WriteString(inputStyle.Render("Name:  "+m.EditNameInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Key:   "+m.EditKeyInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Tags:  "+m.EditTagsInput.View()+" (comma-separated)") + "\n")

		hintText := "(Tab/Shift+Tab • Enter to Save)"
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(hintText))
		return b.String()
	}

	footerText := "↑/↓ navigate • q quit • / filter • e edit • d delete • a add"
	if m.List.FilterState() == list.Filtering {
		footerText = "esc to clear filter • enter to select (if any)"
	}
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Background(lipgloss.Color("235")).Padding(0, 1)
	b.WriteString(m.List.View() + "\n" + footerStyle.Render(footerText))
	return b.String()
}

func (m *Model) updateFocusEdit(forward bool) tea.Cmd {
	switch m.EditFocusIndex {
	case focusEditName:
		m.EditNameInput.Blur()
	case focusEditKey:
		m.EditKeyInput.Blur()
	case focusEditTags:
		m.EditTagsInput.Blur()
	}
	if forward {
		m.EditFocusIndex = (m.EditFocusIndex + 1) % 3
	} else {
		m.EditFocusIndex = (m.EditFocusIndex - 1 + 3) % 3
	}
	switch m.EditFocusIndex {
	case focusEditName:
		return m.EditNameInput.Focus()
	case focusEditKey:
		return m.EditKeyInput.Focus()
	case focusEditTags:
		return m.EditTagsInput.Focus()
	}
	return nil
}
