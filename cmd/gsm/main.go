package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/NumeXx/gsm/pkg/config"
	"github.com/NumeXx/gsm/pkg/runner"
	"github.com/NumeXx/gsm/pkg/tui"
)

var (
	version = "dev"     // Default value
	commit  = "none"    // Default value
	date    = "unknown" // Default value
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

			returnedModel, err := p.StartReturningModel()
			if err != nil {
				fmt.Println("Error running TUI program:", err)
				os.Exit(1)
			}

			if tui.ChosenConnectionGlobal != nil {
				selectedConnDetails := tui.ChosenConnectionGlobal.Connection
				tui.ChosenConnectionGlobal = nil

				now := time.Now()
				foundAndUpdate := false
				cfg := config.GetCurrent()
				for i, conn := range cfg.Connections {
					if conn.Name == selectedConnDetails.Name && conn.Key == selectedConnDetails.Key {
						cfg.Connections[i].LastConnected = &now
						cfg.Connections[i].Usage++
						if errUpdate := config.UpdateConnectionByIndex(i, cfg.Connections[i]); errUpdate == nil {
							if errSave := config.Save(); errSave != nil {
								log.Printf("Error saving config after updating LastConnected/Usage for %s: %v", selectedConnDetails.Name, errSave)
							}
						} else {
							log.Printf("Error updating connection %s in config for LastConnected/Usage: %v", selectedConnDetails.Name, errUpdate)
						}
						foundAndUpdate = true
						break
					}
				}
				if !foundAndUpdate {
					log.Printf("Warning: Could not find connection '%s' in config to update LastConnected/Usage time after TUI selection.", selectedConnDetails.Name)
				}

				if err := runner.Execute(selectedConnDetails); err != nil {
					keyPreview := selectedConnDetails.Key
					if len(keyPreview) > 8 {
						keyPreview = keyPreview[:8]
					}
					fmt.Fprintf(os.Stderr, "Session for %s (%s...) ended with error: %v\n", selectedConnDetails.Name, keyPreview, err)
				}
				fmt.Printf("Session for '%s' closed. Returning to GSM main menu...\n", selectedConnDetails.Name)
			} else {
				if _, ok := returnedModel.(tui.Model); ok {
					// _ = finalModel // Use finalModel if needed for specific exit status, currently not used.
				}
				fmt.Println("Exiting GSM. Thanks for using! See you, bro! 👋")
				break
			}
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GSM",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GSM Version: %s\n", version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("BuildDate: %s\n", date)
	},
}

// init function will be called when the package is initialized.
// We add our importCmd to the rootCmd here.
func init() {
	// If there are other initializations for rootCmd, they would be here.
	// e.g., rootCmd.PersistentFlags().StringVar(...)
	rootCmd.AddCommand(importCmd) // importCmd is defined in import.go (same package main)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
