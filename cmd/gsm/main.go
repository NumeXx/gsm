// GSM (GSocket Manager) - Restructured
// Dependencies: cobra, bubbletea, lipgloss, list

package main

import (
	"fmt"
	"log"
	"os"

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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
