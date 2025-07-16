package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

var (
	configDir  string
	configFile string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get home directory: %v", err))
	}
	configDir = filepath.Join(homeDir, ".ask")
	configFile = filepath.Join(configDir, "config.json")
}

// Load loads the configuration from file
func Load() (*Config, error) {
	config := &Config{
		Model: "gpt-3.5-turbo", // default model
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// Save saves the configuration to file
func Save(config *Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetAvailableModels returns a list of available OpenAI models
func GetAvailableModels() []string {
	return []string{
		"gpt-4.1-nano",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-4-turbo",
		"gpt-4",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configFile
}
