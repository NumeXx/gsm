// GSM (GSocket Manager) - Restructured
// Dependencies: cobra, bubbletea, lipgloss, list

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	// "os/exec" // Tidak lagi dipakai langsung di sini
	// "runtime" // Tidak lagi dipakai langsung di sini
	// "strings" // Mungkin tidak dipakai lagi di sini secara langsung

	tea "github.com/charmbracelet/bubbletea" // Untuk tea.NewProgram
	"github.com/spf13/cobra"

	"github.com/NumeXx/gsm/pkg/config" // Path modul Go lu
	"github.com/NumeXx/gsm/pkg/runner" // Package Runner yang baru
	"github.com/NumeXx/gsm/pkg/tui"    // Package TUI yang baru
	// "github.com/NumeXx/gsm/pkg/utils" // Tidak jadi dipakai untuk ClearScreen
)

// var chosenConnectionGlobal *item // Sudah pindah ke pkg/tui sebagai tui.ChosenConnectionGlobal
// const gsNetcatCommand = "gs-netcat" // Sudah pindah ke pkg/runner

// Fungsi clearScreen() sudah tidak relevan lagi dan dihapus.

// === CLI ROOT ===
var rootCmd = &cobra.Command{
	Use:   "gsm",
	Short: "GSocket Manager - Connect seamlessly",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(); err != nil {
			fmt.Printf("Critical error loading config from '%s': %v\n", config.DefaultConfigFilePath, err)
			os.Exit(1)
		}

		for {
			if err := config.Load(); err != nil {
				fmt.Printf("Error reloading config for TUI: %v. Exiting.\n", err)
				os.Exit(1)
			}

			tuiModel := tui.NewModel(config.GetCurrent())
			p := tea.NewProgram(tuiModel)

			if err := p.Start(); err != nil {
				fmt.Println("Error running TUI program:", err)
				os.Exit(1)
			}

			if tui.ChosenConnectionGlobal != nil {
				selectedConn := tui.ChosenConnectionGlobal.Connection

				// Panggil runner untuk mengeksekusi koneksi
				if err := runner.Execute(selectedConn); err != nil {
					// runner.Execute sudah print pesan disconnected.
					// Di sini kita bisa log error internal jika perlu, tapi biasanya tidak untuk user.
					// log.Printf("Session for %s ended with error: %v", selectedConn.Name, err)
				}
				fmt.Println("Returning to GSM main menu...") // Pesan ini tetap di sini

			} else {
				fmt.Println("Exiting GSM. Thanks for using! See you, bro! 👋")
				break
			}
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "config",
	Short: "Add or update gsocket connection configurations",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(); err != nil {
			fmt.Printf("Warning: Problem loading existing config (%v). Operations will proceed on a potentially empty/new configuration.\n", err)
		}

		var name, key, tagRaw string
		fmt.Print("Connection name: ")
		fmt.Scanln(&name)
		fmt.Print("GSocket key (-s value): ")
		fmt.Scanln(&key)
		fmt.Print("Tags (comma separated, e.g., work,personal): ")
		fmt.Scanln(&tagRaw)
		tags := []string{}
		if strings.TrimSpace(tagRaw) != "" { // Perlu import "strings" jika belum ada
			tagParts := strings.Split(tagRaw, ",")
			for _, t := range tagParts {
				tags = append(tags, strings.TrimSpace(t))
			}
		}
		newConn := config.Connection{Name: name, Key: key, Tags: tags, Usage: 0}
		config.AddConnection(newConn)

		if err := config.Save(); err != nil {
			fmt.Println("ERROR: Failed to save config:", err)
			os.Exit(1)
		}
		fmt.Println("✅ Config saved successfully to", config.DefaultConfigFilePath)
	},
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	rootCmd.AddCommand(addCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
