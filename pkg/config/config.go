package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultConfigFileName is the standard name for the configuration file.
const DefaultConfigFileName = "config.json"

// DefaultConfigDirName is the standard name for the configuration directory.
const DefaultConfigDirName = ".gsm"

// DefaultConfigFilePath is the computed default path to the configuration file.
var DefaultConfigFilePath string

// Connection struct holds all data for a single GSocket connection entry.
type Connection struct {
	Name          string     `json:"name"`
	Key           string     `json:"key"`
	Tags          []string   `json:"tags,omitempty"`
	Usage         int        `json:"usage,omitempty"`
	LastConnected *time.Time `json:"last_connected,omitempty"`
}

// Config struct holds all connections and global settings.
// Currently, only connections are stored.
type Config struct {
	Connections []Connection `json:"connections"`
}

var currentConfig Config

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		// This is a critical error, but we'll let Load() handle printing it if it occurs there.
		// For now, DefaultConfigFilePath might be incorrect if home dir is not found.
		DefaultConfigFilePath = filepath.Join(DefaultConfigDirName, DefaultConfigFileName)
		return
	}
	DefaultConfigFilePath = filepath.Join(home, DefaultConfigDirName, DefaultConfigFileName)
}

// Load reads the configuration from the default file path into currentConfig.
// It creates the directory and an empty config file if they don't exist.
func Load() error {
	configDir := filepath.Dir(DefaultConfigFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if mkDirErr := os.MkdirAll(configDir, 0700); mkDirErr != nil {
			return fmt.Errorf("failed to create config directory '%s': %w", configDir, mkDirErr)
		}
	}

	if _, err := os.Stat(DefaultConfigFilePath); os.IsNotExist(err) {
		fmt.Printf("Config file not found at '%s'. Creating a new empty config.\n", DefaultConfigFilePath)
		currentConfig = Config{Connections: []Connection{}}
		return Save() // Save the new empty config
	}

	data, err := os.ReadFile(DefaultConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to read config file '%s': %w", DefaultConfigFilePath, err)
	}

	if len(data) == 0 { // File exists but is empty
		currentConfig = Config{Connections: []Connection{}}
		return nil
	}

	err = json.Unmarshal(data, &currentConfig)
	if err != nil {
		return fmt.Errorf("failed to parse config file '%s': %w", DefaultConfigFilePath, err)
	}
	return nil
}

// GetCurrent returns a copy of the currently loaded configuration.
func GetCurrent() Config {
	return currentConfig
}

// Save writes the currentConfig to the default file path.
func Save() error {
	data, err := json.MarshalIndent(currentConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	configDir := filepath.Dir(DefaultConfigFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if mkDirErr := os.MkdirAll(configDir, 0700); mkDirErr != nil {
			return fmt.Errorf("failed to create config directory '%s' for saving: %w", configDir, mkDirErr)
		}
	}

	return os.WriteFile(DefaultConfigFilePath, data, 0600)
}

// AddConnection adds a new connection to the current configuration.
// It does not automatically save; Save() must be called separately.
func AddConnection(conn Connection) {
	currentConfig.Connections = append(currentConfig.Connections, conn)
}

// UpdateConnectionByIndex updates an existing connection at a specific index.
// It returns an error if the index is out of bounds.
// It does not automatically save; Save() must be called separately.
func UpdateConnectionByIndex(index int, conn Connection) error {
	if index < 0 || index >= len(currentConfig.Connections) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	currentConfig.Connections[index] = conn
	return nil
}

// DeleteConnectionByIndex removes a connection at a specific index.
// It returns an error if the index is out of bounds.
// It does not automatically save; Save() must be called separately.
func DeleteConnectionByIndex(index int) error {
	if index < 0 || index >= len(currentConfig.Connections) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	currentConfig.Connections = append(currentConfig.Connections[:index], currentConfig.Connections[index+1:]...)
	return nil
}
