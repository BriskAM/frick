package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CachedSuggestion matches the gemini.Suggestion structure to avoid circular imports.
type CachedSuggestion struct {
	Command     string `json:"command"`
	SafetyLevel string `json:"safety_level"`
}

type CacheEntry struct {
	Suggestions []CachedSuggestion `json:"suggestions"`
}

type Cache struct {
	Entries map[string]CacheEntry `json:"entries"`
}

func GetCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "frick", "cache.json"), nil
}

func LoadCache() (*Cache, error) {
	path, err := GetCachePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Cache{Entries: make(map[string]CacheEntry)}, nil
		}
		return nil, err
	}
	var cache Cache
	if err := json.Unmarshal(data, &cache); err != nil {
		return &Cache{Entries: make(map[string]CacheEntry)}, nil
	}
	if cache.Entries == nil {
		cache.Entries = make(map[string]CacheEntry)
	}
	return &cache, nil
}

func SaveCache(cache *Cache) error {
	path, err := GetCachePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
