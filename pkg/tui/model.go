package tui

import (
	"fmt"
	"strings"

	"github.com/NumeXx/gsm/pkg/config" // Sesuaikan dengan path modul Go lu
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
	ni.Placeholder = "Connection Name"
	ni.CharLimit = 100
	ni.Width = 50

	ki := textinput.New()
	ki.Placeholder = "GSocket Key"
	ki.CharLimit = 256
	ki.Width = 50
	// ki.EchoMode = textinput.EchoPassword // Jika ingin menyembunyikan key

	ti := textinput.New() // ti untuk tagsInput
	ti.Placeholder = "tag1,tag2,another-tag"
	ti.CharLimit = 200
	ti.Width = 50

	ChosenConnectionGlobal = nil

	return Model{
		List:           l,
		IsEditing:      false,
		EditNameInput:  ni,
		EditKeyInput:   ki,
		EditTagsInput:  ti,
		EditingIndex:   -1,
		EditFocusIndex: 0, // Default fokus ke input nama
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
				return m, nil
			case tea.KeyTab, tea.KeyShiftTab:
				cmd = m.updateFocusEdit(msg.Type == tea.KeyTab)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			case tea.KeyEnter: // Logika Simpan Perubahan
				newName := strings.TrimSpace(m.EditNameInput.Value())
				newKey := strings.TrimSpace(m.EditKeyInput.Value())
				newTagsRaw := strings.TrimSpace(m.EditTagsInput.Value())

				// Validasi sederhana (minimal nama tidak boleh kosong)
				if newName == "" {
					// TODO: Tampilkan pesan error ke user di TUI, jangan cuma fmt.Println
					fmt.Println("Connection name cannot be empty!") // Sementara
					return m, nil                                   // Tetap di mode edit
				}
				if newKey == "" { // Key juga tidak boleh kosong
					// TODO: Tampilkan pesan error ke user di TUI
					fmt.Println("GSocket key cannot be empty!") // Sementara
					return m, nil                               // Tetap di mode edit
				}

				var tags []string
				if newTagsRaw != "" {
					tagParts := strings.Split(newTagsRaw, ",")
					for _, t := range tagParts {
						tags = append(tags, strings.TrimSpace(t))
					}
				}

				updatedConn := config.Connection{
					Name:  newName,
					Key:   newKey,
					Tags:  tags,
					Usage: config.GetCurrent().Connections[m.EditingIndex].Usage, // Pertahankan Usage count lama
				}

				if err := config.UpdateConnectionByIndex(m.EditingIndex, updatedConn); err != nil {
					// TODO: Tampilkan error save ke user di TUI
					fmt.Printf("Error saving connection: %v\n", err) // Sementara
					return m, nil                                    // Tetap di mode edit jika save gagal
				}

				// Sukses menyimpan
				m.IsEditing = false
				m.EditingIndex = -1
				m.EditNameInput.Blur()
				m.EditKeyInput.Blur()
				m.EditTagsInput.Blur()

				// Panggil config.Load() lagi untuk memastikan GetCurrent() mengambil data terbaru
				if err := config.Load(); err != nil {
					// TODO: Tampilkan error ini ke user di TUI dengan lebih baik
					fmt.Printf("Error reloading config after save for TUI refresh: %v\n", err)
					// Jika gagal load, lebih baik jangan lanjut ke NewModel dengan config lama/salah
					// Kembalikan state IsEditing false saja, user mungkin perlu quit dan run ulang.
					return m, nil
				}
				// Kembalikan Model baru yang diinisialisasi dengan config yang sudah di-update
				// Ini akan me-refresh seluruh TUI.
				reloadedCfg := config.GetCurrent()
				newM := NewModel(reloadedCfg)          // NewModel internalnya gak set size dari param lagi
				newM.lastKnownWidth = m.lastKnownWidth // Oper ukuran lama
				newM.lastKnownHeight = m.lastKnownHeight

				// Langsung panggil SetSize di model baru sebelum di-return
				// Ini penting biar list di model baru ukurannya pas sebelum View() pertama.
				if newM.lastKnownWidth > 0 && newM.lastKnownHeight > 0 { // Pastiin ukurannya valid
					newM.List.SetSize(newM.lastKnownWidth, newM.lastKnownHeight-1)
				}
				return newM, nil // Return model baru
			}
		}

		// Forward input ke field yang sedang fokus
		switch m.EditFocusIndex {
		case 0:
			m.EditNameInput, cmd = m.EditNameInput.Update(msg)
		case 1:
			m.EditKeyInput, cmd = m.EditKeyInput.Update(msg)
		case 2:
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
						m.EditFocusIndex = 0
						return m, m.EditNameInput.Focus()
					}
				}
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
	if m.IsEditing {
		var b strings.Builder
		headerStyle := lipgloss.NewStyle().Bold(true).MarginBottom(1)
		b.WriteString(headerStyle.Render(fmt.Sprintf("Editing Connection: %s (Esc to Cancel)", config.GetCurrent().Connections[m.EditingIndex].Name)) + "\n")

		inputStyle := lipgloss.NewStyle().MarginBottom(1)
		b.WriteString(inputStyle.Render("Name:  "+m.EditNameInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Key:   "+m.EditKeyInput.View()) + "\n")
		b.WriteString(inputStyle.Render("Tags:  "+m.EditTagsInput.View()+" (comma-separated)") + "\n")

		hintText := "(Tab/Shift+Tab to navigate fields • Enter to Save - SOON!)"
		b.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(hintText))
		return b.String()
	}

	footerText := "↑/↓ navigate • q quit • / filter • e edit"
	if m.List.FilterState() == list.Filtering {
		footerText = "esc to clear filter • enter to select (if any)"
	}
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Background(lipgloss.Color("235")).Padding(0, 1)
	return m.List.View() + "\n" + footerStyle.Render(footerText)
}

// updateFocusEdit mengalihkan fokus antar input field di form edit
func (m *Model) updateFocusEdit(forward bool) tea.Cmd {
	// Blur input field saat ini
	switch m.EditFocusIndex {
	case 0:
		m.EditNameInput.Blur()
	case 1:
		m.EditKeyInput.Blur()
	case 2:
		m.EditTagsInput.Blur()
	}

	if forward {
		m.EditFocusIndex = (m.EditFocusIndex + 1) % 3 // 3 adalah jumlah input field
	} else {
		m.EditFocusIndex = (m.EditFocusIndex - 1 + 3) % 3
	}

	// Fokus ke input field baru
	switch m.EditFocusIndex {
	case 0:
		return m.EditNameInput.Focus()
	case 1:
		return m.EditKeyInput.Focus()
	case 2:
		return m.EditTagsInput.Focus()
	}
	return nil
}
