package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/NumeXx/gsm/pkg/utils"
	"github.com/NumeXx/gsm/pkg/wordlist"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var ChosenConnectionGlobal *Item

type Item struct {
	config.Connection
}

func (i Item) Title() string { return i.Name }

func (i Item) Description() string {
	if len(i.Tags) == 0 {
		return ""
	}
	return "# " + strings.Join(i.Tags, ", ")
}

func (i Item) FilterValue() string { return i.Name + " " + strings.Join(i.Tags, " ") }

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
	List                 list.Model
	IsEditing            bool
	EditNameInput        textinput.Model
	EditKeyInput         textinput.Model
	EditTagsInput        textinput.Model
	EditingIndex         int
	EditFocusIndex       int
	lastKnownWidth       int
	lastKnownHeight      int
	StatusMessage        string
	StatusType           StatusMessageType
	IsConfirmingDelete   bool
	DeleteIndex          int
	DeleteConnectionName string
	detailViewport       viewport.Model
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

	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("23")).
		Foreground(lipgloss.Color("231")).
		Padding(0, 1).
		Bold(true)
	l.Title = "GSM | GSocket Manager"
	l.Styles.Title = titleStyle

	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).MaxHeight(1)
	delegate.Styles.NormalDesc = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).MaxHeight(1)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("231")).
		Background(lipgloss.Color("32")).
		Padding(0, 1).
		Bold(true)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")).
		Background(lipgloss.Color("32")).
		Padding(0, 1).MaxHeight(1)
	delegate.SetHeight(2)
	l.SetDelegate(delegate)

	l.SetShowStatusBar(true)
	l.SetStatusBarItemName("connection", "connections")
	l.Styles.StatusBar = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("235")).Foreground(lipgloss.Color("242"))
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("202"))

	l.SetShowHelp(false)

	ni := textinput.New()
	ni.Placeholder = "Name (optional, auto-gen from Key)"
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

	dvp := viewport.New(0, 0)

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
		detailViewport:       dvp,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.List.StartSpinner(),
		textinput.Blink,
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
					return m, tea.ClearScreen
				}

				reloadedCfg := config.GetCurrent()
				newM := NewModel(reloadedCfg)
				newM.lastKnownWidth = m.lastKnownWidth
				newM.lastKnownHeight = m.lastKnownHeight
				newM.StatusMessage = m.StatusMessage
				newM.StatusType = m.StatusType

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 {
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, tea.ClearScreen
			case "n", "esc", "ctrl+c":
				m.IsConfirmingDelete = false
				m.DeleteIndex = -1
				m.DeleteConnectionName = ""
				m.StatusMessage = "Delete cancelled."
				m.StatusType = StatusNone
				return m, tea.ClearScreen
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
				return m, tea.ClearScreen
			case tea.KeyTab, tea.KeyShiftTab:
				cmd = m.updateFocusEdit(msg.Type == tea.KeyTab)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case tea.KeyEnter:
				m.StatusMessage = ""
				m.StatusType = StatusNone

				nameFromForm := strings.TrimSpace(m.EditNameInput.Value())
				keyFromForm := strings.TrimSpace(m.EditKeyInput.Value())
				tagsRawFromForm := strings.TrimSpace(m.EditTagsInput.Value())

				finalName := nameFromForm
				generatedNameInfo := ""

				if m.EditingIndex == EditingIndexAddNew && finalName == "" && keyFromForm != "" {
					dictionary := wordlist.GetWords()
					if len(dictionary) > 0 {
						generatedName, err := utils.GenerateMnemonic(keyFromForm, 3, dictionary)
						if err == nil {
							finalName = generatedName
							generatedNameInfo = fmt.Sprintf(" (Name auto-generated: %s)", finalName)
						} else {
							m.StatusMessage = fmt.Sprintf("Error generating name: %v. Using key prefix.", err)
							m.StatusType = StatusError
							if len(keyFromForm) > 8 {
								finalName = "GsConn-" + keyFromForm[:8]
							} else {
								finalName = "GsConn-" + keyFromForm
							}
						}
					} else {
						m.StatusMessage = "Wordlist empty. Cannot generate name. Using key prefix."
						m.StatusType = StatusError
						if len(keyFromForm) > 8 {
							finalName = "GsConn-" + keyFromForm[:8]
						} else {
							finalName = "GsConn-" + keyFromForm
						}
					}
				}

				if finalName == "" {
					m.StatusMessage = "Connection name cannot be empty!"
					m.StatusType = StatusError
					return m, m.EditNameInput.Focus()
				}

				if keyFromForm == "" {
					m.StatusMessage = "GSocket key cannot be empty!"
					m.StatusType = StatusError
					return m, m.EditKeyInput.Focus()
				}

				var tags []string
				if tagsRawFromForm != "" {
					tagParts := strings.Split(tagsRawFromForm, ",")
					for _, t := range tagParts {
						tags = append(tags, strings.TrimSpace(t))
					}
				}

				var saveErr error
				var successMessage string

				if m.EditingIndex == EditingIndexAddNew {
					newConn := config.Connection{Name: finalName, Key: keyFromForm, Tags: tags, Usage: 0}
					for _, existingConn := range config.GetCurrent().Connections {
						if existingConn.Name == finalName {
							m.StatusMessage = fmt.Sprintf("Error: Connection name '%s' already exists!", finalName)
							m.StatusType = StatusError
							return m, m.EditNameInput.Focus()
						}
					}
					config.AddConnection(newConn)
					saveErr = config.Save()
					successMessage = fmt.Sprintf("Connection '%s' added%s.", finalName, generatedNameInfo)
				} else {
					originalName := config.GetCurrent().Connections[m.EditingIndex].Name
					if originalName != finalName {
						for i, existingConn := range config.GetCurrent().Connections {
							if i != m.EditingIndex && existingConn.Name == finalName {
								m.StatusMessage = fmt.Sprintf("Error: Connection name '%s' already exists!", finalName)
								m.StatusType = StatusError
								return m, m.EditNameInput.Focus()
							}
						}
					}
					updatedConn := config.Connection{
						Name:  finalName,
						Key:   keyFromForm,
						Tags:  tags,
						Usage: config.GetCurrent().Connections[m.EditingIndex].Usage,
					}
					saveErr = config.UpdateConnectionByIndex(m.EditingIndex, updatedConn)
					successMessage = fmt.Sprintf("Connection '%s' updated.", finalName)
				}

				if saveErr != nil {
					m.StatusMessage = fmt.Sprintf("Error saving: %v", saveErr)
					m.StatusType = StatusError
					return m, m.EditNameInput.Focus()
				} else {
					if m.StatusType != StatusError {
						m.StatusMessage = successMessage
						m.StatusType = StatusSuccess
					}
				}

				m.IsEditing = false
				m.EditingIndex = -1
				m.EditNameInput.Blur()
				m.EditKeyInput.Blur()
				m.EditTagsInput.Blur()

				if err := config.Load(); err != nil {
					m.StatusMessage = fmt.Sprintf("Error reloading config after save: %v. Please restart GSM.", err)
					m.StatusType = StatusError
					return m, tea.ClearScreen
				}

				newM := NewModel(config.GetCurrent())
				newM.lastKnownWidth = m.lastKnownWidth
				newM.lastKnownHeight = m.lastKnownHeight
				newM.StatusMessage = m.StatusMessage
				newM.StatusType = m.StatusType

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 {
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, tea.ClearScreen
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
		listHeight := m.lastKnownHeight - 2
		if m.StatusMessage != "" {
			listHeight--
		}

		listColumnWidth := (m.lastKnownWidth * 40) / 100
		detailColumnWidth := m.lastKnownWidth - listColumnWidth - 1

		m.List.SetSize(listColumnWidth, listHeight)
		m.detailViewport.Width = detailColumnWidth
		m.detailViewport.Height = listHeight

		if item, ok := m.List.SelectedItem().(Item); ok {
			m.detailViewport.SetContent(m.renderDetailPanel(item))
		} else {
			m.detailViewport.SetContent("No connection selected.")
		}
		return m, nil
	case tea.KeyMsg:
		if m.IsEditing || m.IsConfirmingDelete {
			break
		}

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
		if msg.Type == tea.KeyEnter && !m.IsEditing && !m.IsConfirmingDelete {
			if selected, ok := m.List.SelectedItem().(Item); ok {
				ChosenConnectionGlobal = &selected
				return m, tea.Quit
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	cmds = append(cmds, cmd)

	if item, ok := m.List.SelectedItem().(Item); ok {
		m.detailViewport.SetContent(m.renderDetailPanel(item))
	} else if len(m.List.Items()) == 0 {
		m.detailViewport.SetContent("No connections available.")
	} else {
		m.detailViewport.SetContent("No connection selected, or list is empty.")
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.IsConfirmingDelete {
		var b strings.Builder
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
		var formBuilder strings.Builder
		headerStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
		var formTitle string
		if m.EditingIndex == EditingIndexAddNew {
			formTitle = "Add New Connection (Esc to Cancel)"
		} else if m.EditingIndex >= 0 && m.EditingIndex < len(config.GetCurrent().Connections) {
			formTitle = fmt.Sprintf("Editing Connection: %s (Esc to Cancel)", config.GetCurrent().Connections[m.EditingIndex].Name)
		} else {
			formTitle = "Edit Connection (Esc to Cancel)"
		}
		formBuilder.WriteString(headerStyle.Render(formTitle) + "\n")

		inputStyle := lipgloss.NewStyle().MarginBottom(1)
		formBuilder.WriteString(inputStyle.Render("Name:  "+m.EditNameInput.View()) + "\n")
		formBuilder.WriteString(inputStyle.Render("Key:   "+m.EditKeyInput.View()) + "\n")
		formBuilder.WriteString(inputStyle.Render("Tags:  "+m.EditTagsInput.View()+" (comma-separated)") + "\n")

		if m.StatusMessage != "" && (m.StatusType == StatusError || strings.Contains(m.StatusMessage, "generated name")) {
			var statusStyle lipgloss.Style
			if m.StatusType == StatusError {
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).MarginTop(1)
			} else {
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).MarginTop(1)
			}
			formBuilder.WriteString("\n" + statusStyle.Render(m.StatusMessage))
		}

		hintText := "(Tab/Shift+Tab • Enter to Save)"
		formBuilder.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(hintText))

		return formBuilder.String()
	}

	// Main view: List on the left, Details on the right
	statusLine := ""
	if m.StatusMessage != "" {
		var statusStyle lipgloss.Style
		if m.StatusType == StatusError {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Padding(0, 1)
		} else if m.StatusType == StatusSuccess {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("40")).Bold(true).Padding(0, 1)
		} else {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Padding(0, 1)
		}
		statusLine = statusStyle.Render(m.StatusMessage)
	}

	listColumnWidth := (m.lastKnownWidth * 40) / 100
	if listColumnWidth < 30 {
		listColumnWidth = 30
	}
	if listColumnWidth > m.lastKnownWidth-25 {
		listColumnWidth = m.lastKnownWidth - 25
	}
	if m.lastKnownWidth < 50 {
		listColumnWidth = m.lastKnownWidth
	}

	// Calculate height for list and viewport (panel kanan)
	availableHeight := m.lastKnownHeight - 1 // Subtract 1 for footer
	if statusLine != "" {
		availableHeight-- // Subtract 1 more if status line is visible
	}
	if availableHeight < 1 {
		availableHeight = 1 // Ensure height is at least 1
	}

	m.List.SetHeight(availableHeight)         // Set list height
	m.detailViewport.Height = availableHeight // Match viewport height with list height

	listRender := m.List.View()
	var finalCombinedView string
	var detailPanelRenderedContent string // Declare here to ensure it's always available

	if m.lastKnownWidth > listColumnWidth+5 { // Only show detail panel if there is enough space
		detailColumnWidth := m.lastKnownWidth - listColumnWidth - 1
		m.detailViewport.Width = detailColumnWidth
		m.detailViewport.Height = availableHeight

		if item, ok := m.List.SelectedItem().(Item); ok {
			m.detailViewport.SetContent(m.renderDetailPanel(item))
		} else if len(m.List.Items()) == 0 {
			m.detailViewport.SetContent("No connections.")
		} else {
			m.detailViewport.SetContent("Select a connection.")
		}
		detailPanelRenderedContent = m.detailViewport.View()

		separatorStyle := lipgloss.NewStyle().SetString("│").Foreground(lipgloss.Color("239"))
		separatorHeight := m.detailViewport.Height
		if separatorHeight < 1 {
			separatorHeight = 1
		}
		separator := separatorStyle.Render(strings.Repeat("\n", separatorHeight-1))
		finalCombinedView = lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(listColumnWidth).Render(listRender),
			separator,
			lipgloss.NewStyle().Width(m.detailViewport.Width).PaddingLeft(1).Render(detailPanelRenderedContent),
		)
	} else { // Not enough space for detail panel, list takes full width
		m.List.SetSize(m.lastKnownWidth, availableHeight)
		listRender = m.List.View() // Re-render list with full width
		finalCombinedView = listRender
		// detailPanelRenderedContent remains empty string, which is fine
	}

	mainVerticalParts := []string{finalCombinedView}
	if statusLine != "" {
		mainVerticalParts = append(mainVerticalParts, statusLine)
	}

	footerText := "↑/↓ nav • q quit • / filter • e edit • d del • a add • Enter exec"
	if m.List.FilterState() == list.Filtering {
		footerText = "esc clear • enter select"
	}
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Background(lipgloss.Color("235")).Padding(0, 1)
	mainVerticalParts = append(mainVerticalParts, footerStyle.Render(footerText))

	return lipgloss.JoinVertical(lipgloss.Left, mainVerticalParts...)
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

func (m Model) renderDetailPanel(item Item) string {
	var s strings.Builder
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	valueStyle := lipgloss.NewStyle().Bold(true)

	s.WriteString(valueStyle.Render(item.Name) + "\n\n")
	s.WriteString(keyStyle.Render("Key: ") + valueStyle.Render(item.Key[:min(len(item.Key), 20)]+"...") + "\n")
	if len(item.Tags) > 0 {
		s.WriteString(keyStyle.Render("Tags: ") + valueStyle.Render(strings.Join(item.Tags, ", ")) + "\n")
	} else {
		s.WriteString(keyStyle.Render("Tags: ") + valueStyle.Render("-") + "\n")
	}
	s.WriteString(keyStyle.Render("Usage: ") + valueStyle.Render(fmt.Sprintf("%d times", item.Usage)) + "\n")

	lastConnectedStr := "Never"
	if item.LastConnected != nil {
		if time.Since(*item.LastConnected).Hours() < 24*7 {
			if time.Now().YearDay() == item.LastConnected.YearDay() && time.Now().Year() == item.LastConnected.Year() {
				lastConnectedStr = "Today, " + item.LastConnected.Format("15:04")
			} else if time.Now().YearDay()-1 == item.LastConnected.YearDay() && time.Now().Year() == item.LastConnected.Year() {
				lastConnectedStr = "Yesterday, " + item.LastConnected.Format("15:04")
			} else {
				lastConnectedStr = item.LastConnected.Format("Mon, 2 Jan 15:04")
			}
		} else {
			lastConnectedStr = item.LastConnected.Format("2 Jan 2006")
		}
	}
	s.WriteString(keyStyle.Render("Last Seen: ") + valueStyle.Render(lastConnectedStr) + "\n")

	return s.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
