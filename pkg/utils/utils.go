package utils

import (
	"os"
	"os/exec"
	"runtime"
)

// ClearScreen membersihkan layar terminal.
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	// Error sengaja diabaikan untuk fungsi utilitas sederhana ini,
	// karena kegagalan clear screen biasanya bukan error kritis.
	_ = cmd.Run()
}
