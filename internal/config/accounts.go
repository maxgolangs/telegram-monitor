package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type accountJSON struct {
	Name    string `json:"name"`
	Session string `json:"session"`
}

func LoadAccounts(accountsFile string, fallbackSessionPath string) ([]Account, error) {
	accountsFile = strings.TrimSpace(accountsFile)
	if accountsFile == "" {
		if strings.TrimSpace(fallbackSessionPath) == "" {
			fallbackSessionPath = "sessions/session.bin"
		}
		return []Account{{Name: "account-1", SessionPath: fallbackSessionPath}}, nil
	}

	data, err := os.ReadFile(accountsFile)
	if err != nil {
		return nil, fmt.Errorf("read accounts file: %w", err)
	}
	var raw []accountJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse accounts file (json): %w", err)
	}
	out := make([]Account, 0, len(raw))
	for i, a := range raw {
		name := strings.TrimSpace(a.Name)
		if name == "" {
			name = fmt.Sprintf("account-%d", i+1)
		}
		session := strings.TrimSpace(a.Session)
		if session == "" {
			return nil, fmt.Errorf("account %q has empty session path", name)
		}
		out = append(out, Account{Name: name, SessionPath: session})
	}
	return out, nil
}



