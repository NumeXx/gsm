package tui

import (
	"fmt"
	"strings"

	// "time" // Tidak dipakai lagi untuk auto-clear status

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ChosenConnectionGlobal adalah variabel package-level untuk menyimpan koneksi yang dipilih.
// Ini akan di-set oleh Update() dan dibaca oleh cmd/gsm/main.go setelah TUI quit.
var ChosenConnectionGlobal *Item // Menggunakan tipe Item dari package ini

// Item adalah representasi satu koneksi di TUI list.
// Ini meng-embed config.Connection untuk mendapatkan field-fieldnya.
type Item struct {
	config.Connection // Embed Connection dari pkg/config
}

// Title untuk list.Item interface.
func (i Item) Title() string { return fmt.Sprintf("Name : %s", i.Name) }

// Description untuk list.Item interface.
func (i Item) Description() string {
	if len(i.Tags) == 0 {
		return "Tag  : -"
	}
	return fmt.Sprintf("Tag  : %s", strings.Join(i.Tags, ", "))
}

// FilterValue untuk list.Item interface.
func (i Item) FilterValue() string { return i.Name + " " + strings.Join(i.Tags, " ") } // Filter berdasarkan Nama dan Tag

// Konstanta untuk EditFocusIndex
const (
	focusEditName = iota
	focusEditKey
	focusEditTags
)

// Konstanta untuk EditingIndexAddNew
const (
	EditingIndexAddNew = -2 // Penanda khusus untuk mode Add New
)

// StatusMessageType mendefinisikan tipe dari pesan status
type StatusMessageType int

const (
	StatusNone StatusMessageType = iota
	// StatusSuccess // Tidak dipakai lagi
	StatusError
)

// Model adalah BubbleTea model untuk TUI gsm.
type Model struct {
	List            list.Model
	IsEditing       bool            // True jika sedang dalam mode edit
	EditNameInput   textinput.Model // Input untuk Nama Koneksi
	EditKeyInput    textinput.Model // Input untuk Key GSocket
	EditTagsInput   textinput.Model // Input untuk Tags (comma-separated)
	EditingIndex    int             // Indeks item di cfg.Connections yang sedang diedit, -1 jika tidak ada
	EditFocusIndex  int             // 0: Name, 1: Key, 2: Tags (untuk navigasi form)
	lastKnownWidth  int             // Simpen lebar terakhir
	lastKnownHeight int             // Simpen tinggi terakhir

	StatusMessage string            // Pesan status untuk ditampilkan ke user
	StatusType    StatusMessageType // Tipe pesan status (success, error, none)
	// statusClearCmd tea.Cmd // Untuk auto-clear message, implementasi nanti

	// State untuk Delete Confirmation
	IsConfirmingDelete   bool   // True jika sedang dalam mode konfirmasi delete
	DeleteIndex          int    // Index item di cfg.Connections yang akan dihapus
	DeleteConnectionName string // Nama koneksi yang akan dihapus (untuk pesan konfirmasi)
}

// NewModel membuat instance baru dari TUI Model.
// Ia menerima Config yang sudah di-load untuk mengisi list.
func NewModel(cfg config.Config) Model {
	items := []list.Item{}
	for _, c := range cfg.Connections {
		items = append(items, Item{Connection: c})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.SetFilteringEnabled(true)
	l.FilterInput.Placeholder = "Filter by name or tag... (type to search)"
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.FilterInput.TextStyle = lipgloss.NewStyle() // Default text style
	// l.FilterInput.Focus() // Tidak perlu fokus di awal, list yang fokus

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

	l.SetShowStatusBar(true)                            // Kita akan coba tampilkan status bar bawaan list untuk filter
	l.SetStatusBarItemName("connection", "connections") // Custom nama item di status bar
	l.Styles.StatusBar = lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("235")).Foreground(lipgloss.Color("242"))
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("202")) // Kursor filter warna oranye

	l.SetShowHelp(false) // Kita akan handle help text di footer manual jika perlu

	// Inisialisasi TextInputs untuk Mode Edit
	ni := textinput.New()
	ni.Placeholder = "Connection Name (required)"
	ni.CharLimit = 100
	ni.Width = 50

	ki := textinput.New()
	ki.Placeholder = "GSocket Key (required)"
	ki.CharLimit = 256
	ki.Width = 50
	// ki.EchoMode = textinput.EchoPassword // Jika ingin menyembunyikan key

	ti := textinput.New() // ti untuk tagsInput
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
		StatusMessage:        "", // Init status message kosong
		StatusType:           StatusNone,
		IsConfirmingDelete:   false,
		DeleteIndex:          -1,
		DeleteConnectionName: "",
		// lastKnownWidth dan lastKnownHeight akan diisi oleh WindowSizeMsg pertama
	}
}

// Init untuk BubbleTea model.
func (m Model) Init() tea.Cmd {
	// Kita coba kirim command yang mungkin dibutuhin list buat render awal
	// atau buat refresh viewport-nya.
	// list.Model.Init() sendiri biasanya return nil, tapi kita bisa coba:
	return tea.Batch(
		m.List.StartSpinner(), // Kalo pake spinner, ini bisa jadi pemicu
		textinput.Blink,       // Kalo ada textinput yang aktif (filter)
		tea.ClearScreen,       // Coba clear screen dulu (walaupun BubbleTea udah handle alt screen)
	)
	// Atau cuma return nil jika tidak ada command khusus yang jelas.
	// return nil
}

// Update untuk BubbleTea model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle Konfirmasi Delete dulu jika aktif (override mode lain)
	if m.IsConfirmingDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch strings.ToLower(msg.String()) {
			case "y":
				if err := config.DeleteConnectionByIndex(m.DeleteIndex); err != nil {
					m.StatusMessage = fmt.Sprintf("Error deleting '%s': %v", m.DeleteConnectionName, err)
					m.StatusType = StatusError
				} else {
					// Pesan sukses tidak ditampilkan sesuai request
					// m.StatusMessage = fmt.Sprintf("Connection '%s' deleted.", m.DeleteConnectionName)
					// m.StatusType = StatusSuccess
				}
				m.IsConfirmingDelete = false
				m.DeleteIndex = -1
				m.DeleteConnectionName = ""
				if err := config.Load(); err != nil { // Reload config
					m.StatusMessage = fmt.Sprintf("Error reloading config after delete: %v", err)
					m.StatusType = StatusError
					return m, nil
				}

				// REFRESH TUI DENGAN CARA YANG BENER (OPER UKURAN)
				reloadedCfg := config.GetCurrent()
				newM := NewModel(reloadedCfg)
				newM.lastKnownWidth = m.lastKnownWidth // Oper ukuran lama
				newM.lastKnownHeight = m.lastKnownHeight

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 { // Pastiin ukurannya valid
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, nil // Return model baru
			case "n", "esc", "ctrl+c":
				m.IsConfirmingDelete = false
				m.DeleteIndex = -1
				m.DeleteConnectionName = ""
				m.StatusMessage = "Delete cancelled."
				m.StatusType = StatusNone // Atau tipe info jika ada
				return m, nil
			}
		}
		return m, nil // Jika bukan KeyMsg atau key tidak dikenal, jangan lakukan apa-apa di mode konfirmasi
	}

	if m.IsEditing {
		// === UPDATE SAAT MODE EDIT ===
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
					return m, m.EditNameInput.Focus() // Fokus kembali ke input nama
				}
				if newKey == "" {
					m.StatusMessage = "GSocket key cannot be empty!"
					m.StatusType = StatusError
					return m, m.EditKeyInput.Focus() // Fokus kembali ke input key
				}

				var tags []string
				if newTagsRaw != "" {
					tagParts := strings.Split(newTagsRaw, ",")
					for _, t := range tagParts {
						tags = append(tags, strings.TrimSpace(t))
					}
				}

				var saveErr error
				if m.EditingIndex == EditingIndexAddNew { // Mode Add New
					newConn := config.Connection{Name: newName, Key: newKey, Tags: tags, Usage: 0}
					config.AddConnection(newConn)
					saveErr = config.Save()
				} else { // Mode Edit Existing
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
				// Tidak ada transfer status message sukses ke newM

				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 {
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, nil // Langsung refresh TUI tanpa pesan sukses
			}
		}

		// Forward input ke field yang sedang fokus
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

	// === UPDATE SAAT MODE LIST (NORMAL) ===
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.lastKnownWidth = msg.Width
		m.lastKnownHeight = msg.Height
		m.List.SetSize(m.lastKnownWidth, m.lastKnownHeight-1) // -1 untuk footer
		return m, nil
	case tea.KeyMsg:
		// Hapus status message error dari mode edit jika ada input lain di mode list
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
						// Hapus status message lama sebelum masuk mode edit
						m.StatusMessage = ""
						m.StatusType = StatusNone
						return m, m.EditNameInput.Focus()
					}
				}
			case "d": // Tombol untuk masuk mode Konfirmasi Delete
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
							m.StatusMessage = "" // Hapus status message lama
							m.StatusType = StatusNone
							return m, nil // View akan menampilkan konfirmasi
						} // else: item tidak ditemukan, jangan lakukan apa-apa
					}
				}
			case "a": // Tombol untuk masuk mode Add New
				m.IsEditing = true
				m.EditingIndex = EditingIndexAddNew // Gunakan penanda Add New
				// Kosongkan form
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

// View untuk BubbleTea model.
func (m Model) View() string {
	var b strings.Builder

	if m.StatusMessage != "" && m.StatusType == StatusError { // Hanya tampilkan status jika error
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Padding(0, 1).Bold(true) // Merah
		b.WriteString(statusStyle.Render(m.StatusMessage) + "\n\n")
	}

	if m.IsConfirmingDelete { // Tampilan Konfirmasi Delete
		headerStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1).Foreground(lipgloss.Color("196")) // Merah untuk delete
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
			// Pastikan index valid sebelum akses, meskipun seharusnya sudah aman dari Update
			formTitle = fmt.Sprintf("Editing Connection: %s (Esc to Cancel)", config.GetCurrent().Connections[m.EditingIndex].Name)
		} else {
			formTitle = "Edit Connection (Esc to Cancel)" // Fallback jika index aneh
		}
		b.WriteString(headerStyle.Render(formTitle) + "\n")

		inputStyle := lipgloss.NewStyle().MarginBottom(1)
		b.WriteString(inputStyle.Render("Name:  "+m.EditNameInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Key:   "+m.EditKeyInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Tags:  "+m.EditTagsInput.View()+" (comma-separated)") + "\n")

		hintText := "(Tab/Shift+Tab • Enter to Save)" // Hint lebih singkat
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

// updateFocusEdit mengalihkan fokus antar input field di form edit
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

// Definisi ClearStatusMessageMsg sudah tidak diperlukan dan dihapus
// type ClearStatusMessageMsg struct{}
