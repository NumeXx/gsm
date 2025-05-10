package runner

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/NumeXx/gsm/pkg/config" // Sesuaikan dengan path modul Go lu
)

const gsNetcatCommand = "gs-netcat"

// Execute menjalankan sesi gs-netcat interaktif untuk koneksi yang diberikan.
// Fungsi ini akan menampilkan pesan ke konsol selama proses koneksi dan diskoneksi.
func Execute(conn config.Connection) error {
	fmt.Printf("[+] Attempting to connect to: %s (Key: %s)\n", conn.Name, conn.Key)
	fmt.Println("    (Press Ctrl+C in the GSocket session to disconnect and return to GSM)")

	cmd := exec.Command(gsNetcatCommand, "-i", "-s", conn.Key)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run() // Ini akan nge-block sampai gs-netcat selesai atau di-interrupt

	if runErr != nil {
		// gs-netcat biasanya return error jika di-interrupt (Ctrl+C) atau ada masalah koneksi.
		// Kita tidak log detail errornya ke user karena Ctrl+C itu aksi normal.
		fmt.Printf("[<] Disconnected from %s (session ended, possibly with error: %v)\n", conn.Name, runErr)
		return runErr // Kembalikan errornya agar pemanggil tahu sesi tidak selesai normal
	} else {
		fmt.Printf("[<] Disconnected from %s successfully.\n", conn.Name)
	}
	return nil
}
