package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Account struct {
	Name        string `json:"name"`
	SessionPath string `json:"session"`
}

type State struct {
	AppID   int    `json:"app_id"`
	AppHash string `json:"app_hash"`

	Accounts []Account `json:"accounts"`

	BotToken  string `json:"bot_token"`
	BotChatID int64  `json:"bot_chat_id"`

	KeywordsFile   string `json:"keywords_file"`
	StopwordsFile  string `json:"stopwords_file"`
	AccountsFile   string `json:"accounts_file"`
	PollIntervalMs int64  `json:"poll_interval_ms"`
	PollLimit      int    `json:"poll_limit"`
	UseRegex       bool   `json:"use_regex"`
}

func Default() State {
	return State{
		KeywordsFile:   "data/keywords.txt",
		StopwordsFile:  "data/stopwords.txt",
		AccountsFile:   "data/accounts.json",
		PollLimit:      100,
		PollIntervalMs: 3000,
	}
}

func Load(path string) (State, error) {
	s := Default()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return Default(), fmt.Errorf("config.json parse: %w", err)
	}
	if s.PollLimit <= 0 {
		s.PollLimit = 100
	}
	if strings.TrimSpace(s.KeywordsFile) == "" {
		s.KeywordsFile = "data/keywords.txt"
	}
	if strings.TrimSpace(s.StopwordsFile) == "" {
		s.StopwordsFile = "data/stopwords.txt"
	}
	if strings.TrimSpace(s.AccountsFile) == "" {
		s.AccountsFile = "data/accounts.json"
	}
	return s, nil
}

func Save(path string, s State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}


