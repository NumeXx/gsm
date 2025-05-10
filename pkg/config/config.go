package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Connection struct {
	Name  string   `json:"name"`
	Key   string   `json:"key"`
	Tags  []string `json:"tags"`
	Usage int      `json:"usage"`
}

type Config struct {
	Connections []Connection `json:"connections"`
}

var DefaultConfigFilePath = filepath.Join(os.Getenv("HOME"), ".gsm", "config.json")

var currentConfig Config

func Load() error {
	path := DefaultConfigFilePath
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			currentConfig = Config{Connections: []Connection{}}
			return nil
		}
		return fmt.Errorf("failed to read config file '%s': %w", path, err)
	}
	if err := json.Unmarshal(data, &currentConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config data from '%s': %w", path, err)
	}
	if currentConfig.Connections == nil {
		currentConfig.Connections = []Connection{}
	}
	return nil
}

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

func GetCurrent() Config {
	return currentConfig
}

func AddConnection(conn Connection) {
	if currentConfig.Connections == nil {
		currentConfig.Connections = []Connection{}
	}
	currentConfig.Connections = append(currentConfig.Connections, conn)
}

func UpdateConnectionByIndex(index int, updatedConn Connection) error {
	if index < 0 || index >= len(currentConfig.Connections) {
		return fmt.Errorf("index %d out of bounds for connections list (len %d)", index, len(currentConfig.Connections))
	}
	currentConfig.Connections[index] = updatedConn
	return Save()
}

func DeleteConnectionByIndex(index int) error {
	if index < 0 || index >= len(currentConfig.Connections) {
		return fmt.Errorf("index %d out of bounds for connections list (len %d)", index, len(currentConfig.Connections))
	}
	currentConfig.Connections = append(currentConfig.Connections[:index], currentConfig.Connections[index+1:]...)
	return Save()
}
