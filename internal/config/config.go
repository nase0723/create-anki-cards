package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	OpenAIKey      string            `json:"openai_api_key"`
	AnkiConnectURL string            `json:"ankiconnect_url"`
	DeckName       string            `json:"deck_name"`
	NoteType       string            `json:"note_type"`
	FieldMapping   map[string]string `json:"field_mapping"`
	PollIntervalMs int               `json:"poll_interval_ms"`
	TriggerMode    string            `json:"trigger_mode"` // "hotkey" or "polling"
	PixabayAPIKey  string            `json:"pixabay_api_key"`
	PexelsAPIKey   string            `json:"pexels_api_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		AnkiConnectURL: "http://localhost:8765",
		DeckName:       "English Vocabulary",
		NoteType:       "EnglishVocab",
		FieldMapping: map[string]string{
			"word":    "Word",
			"meaning": "Meaning",
			"example": "Example",
		},
		PollIntervalMs: 500,
		TriggerMode:    "polling",
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.OpenAIKey == "" {
		return nil, fmt.Errorf("openai_api_key is required")
	}

	return cfg, nil
}
