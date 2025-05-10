package tui

import (
	"fmt"
	"strings"

	"github.com/NumeXx/gsm/pkg/config" // Sesuaikan dengan path modul Go lu
	"github.com/charmbracelet/bubbles/list"
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
func (i Item) FilterValue() string { return i.Name } // Filter tetap berdasarkan nama

// Model adalah BubbleTea model untuk TUI gsm.
type Model struct {
	List list.Model
	// chosenConnection field di struct model tidak lagi jadi penentu utama lintas package,
	// kita gunakan ChosenConnectionGlobal.
	// Namun, bisa tetap ada untuk state internal model jika diperlukan di masa depan.
}

// NewModel membuat instance baru dari TUI Model.
// Ia menerima Config yang sudah di-load untuk mengisi list.
func NewModel(cfg config.Config) Model {
	items := []list.Item{}
	for _, c := range cfg.Connections {
		items = append(items, Item{Connection: c})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	// --- Styling TUI ---
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

	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	ChosenConnectionGlobal = nil // Reset global var tiap kali model baru dibuat
	return Model{List: l}
}

// Init untuk BubbleTea model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update untuk BubbleTea model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width, msg.Height-2)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			ChosenConnectionGlobal = nil
			return m, tea.Quit
		case "enter":
			selected, ok := m.List.SelectedItem().(Item)
			if ok {
				ChosenConnectionGlobal = &selected
				return m, tea.Quit
			}
		}
	}

	var listCmd tea.Cmd
	m.List, listCmd = m.List.Update(msg)
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

// View untuk BubbleTea model.
func (m Model) View() string {
	// Cek dari list internal model, bukan dari config global lagi
	if m.List.Items() == nil || len(m.List.Items()) == 0 {
		// Style pesan ini juga bisa, tapi untuk sekarang biarkan simpel
		return "\n   No connections configured. Run 'gsm config' to add one.\n\n   q quit"
	}

	footerText := "↑/↓ navigate • enter connect • q quit"
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("242")).
		Background(lipgloss.Color("235"))

	return m.List.View() + "\n" + footerStyle.Render(footerText)
}
