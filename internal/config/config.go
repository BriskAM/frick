package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const DefaultModel = "gemma-4-31b-it"

type Config struct {
	ApiKey          string            `json:"api_key"`
	Model           string            `json:"model"`
	ApiEndpoint     string            `json:"api_endpoint,omitempty"`
	SafetyOverrides map[string]string `json:"safety_overrides,omitempty"`
	SystemPrompt    string            `json:"system_prompt,omitempty"`
}

// GetConfigPath returns the absolute path to ~/.config/frick/config.json
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "frick", "config.json"), nil
}

// LoadConfig loads the configuration file. Returns a default config if file does not exist.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return default config
			return &Config{Model: DefaultModel}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}

	return &cfg, nil
}

// SaveConfig saves the configuration to the file.
func SaveConfig(cfg *Config) error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

type LastSuggestion struct {
	Command   string `json:"command"`
	Corrected string `json:"corrected"`
	ExitCode  int    `json:"exit_code"`
}

func GetLastSuggestionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "frick", "last_suggestion.json"), nil
}

func LoadLastSuggestion() (*LastSuggestion, error) {
	path, err := GetLastSuggestionPath()
	if err != nil {
		return nil, err
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ls LastSuggestion
	if err := json.Unmarshal(file, &ls); err != nil {
		return nil, err
	}
	return &ls, nil
}

func SaveLastSuggestion(ls *LastSuggestion) error {
	path, err := GetLastSuggestionPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(ls)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
