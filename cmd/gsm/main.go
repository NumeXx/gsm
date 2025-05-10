package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/NumeXx/gsm/pkg/runner"
	"github.com/NumeXx/gsm/pkg/tui"
)

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

				if err := runner.Execute(selectedConn); err != nil {
					log.Printf("Session for %s ended with error: %v", selectedConn.Name, err)
				}
				fmt.Println("Returning to GSM main menu...")

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
