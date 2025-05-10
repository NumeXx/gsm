package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Connection mendefinisikan struktur untuk satu koneksi GSocket.
type Connection struct {
	Name  string   `json:"name"`
	Key   string   `json:"key"`
	Tags  []string `json:"tags"`
	Usage int      `json:"usage"`
}

// Config mendefinisikan struktur untuk keseluruhan konfigurasi GSM.
type Config struct {
	Connections []Connection `json:"connections"`
}

// DefaultConfigFilePath adalah path default ke file konfigurasi.
var DefaultConfigFilePath = filepath.Join(os.Getenv("HOME"), ".gsm", "config.json")

// currentConfig menyimpan konfigurasi yang sedang di-load atau di-update.
// Ini adalah variabel package-level, bukan global untuk seluruh aplikasi.
var currentConfig Config

// Load memuat konfigurasi dari DefaultConfigFilePath ke currentConfig.
// Fungsi ini akan mengembalikan error jika pembacaan atau unmarshalling gagal.
func Load() error {
	path := DefaultConfigFilePath
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File tidak ada, inisialisasi dengan config kosong.
			// Ini dianggap bukan error untuk Load, tapi mungkin perlu ditangani khusus oleh pemanggil.
			currentConfig = Config{Connections: []Connection{}}
			return nil // Tidak ada error jika file tidak ada, config kosong di-load
		}
		return fmt.Errorf("failed to read config file '%s': %w", path, err)
	}
	if err := json.Unmarshal(data, &currentConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config data from '%s': %w", path, err)
	}
	return nil
}

// Save menyimpan currentConfig ke DefaultConfigFilePath.
// Fungsi ini akan membuat direktori jika belum ada.
func Save() error {
	path := DefaultConfigFilePath
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory '%s': %w", dir, err)
	}

	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config data: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config to '%s': %w", path, err)
	}
	return nil
}

// GetCurrent returns salinan dari konfigurasi yang saat ini di-load.
func GetCurrent() Config {
	// Mengembalikan salinan untuk menghindari modifikasi tak sengaja dari luar package
	// meskipun dalam kasus ini, Connections adalah slice jadi tetap bisa dimodifikasi.
	// Untuk keamanan lebih, bisa dilakukan deep copy.
	return currentConfig
}

// AddConnection menambahkan koneksi baru ke currentConfig.
// Perubahan ini belum disimpan ke file sampai Save() dipanggil.
func AddConnection(conn Connection) {
	if currentConfig.Connections == nil {
		currentConfig.Connections = []Connection{}
	}
	currentConfig.Connections = append(currentConfig.Connections, conn)
}

// UpdateConnection (Contoh, belum diimplementasikan sepenuhnya)
// func UpdateConnection(updatedConn Connection) error {
// 	for i, conn := range currentConfig.Connections {
// 		if conn.Name == updatedConn.Name { // Asumsi nama unik
// 			currentConfig.Connections[i] = updatedConn
// 			return nil
// 		}
// 	}
// 	return fmt.Errorf("connection '%s' not found for update", updatedConn.Name)
// }

// TODO: Tambahkan fungsi DeleteConnection jika diperlukan.
