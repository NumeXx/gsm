package runner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/NumeXx/gsm/pkg/config"
)

const gsNetcatCommand = "gs-netcat"

func Execute(conn config.Connection) error {
	fmt.Printf("[+] Attempting to connect to: %s (Key: %s)\n", conn.Name, conn.Key)
	fmt.Println("    (Press Ctrl+C in the GSocket session to disconnect and return to GSM)")

	cmd := exec.Command(gsNetcatCommand, "-i", "-s", conn.Key)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()

	if runErr != nil {
		fmt.Printf("[<] Disconnected from %s (session ended, possibly with error: %v)\n", conn.Name, runErr)
		return runErr
	}
	fmt.Printf("[<] Disconnected from %s successfully.\n", conn.Name)
	return nil
}
