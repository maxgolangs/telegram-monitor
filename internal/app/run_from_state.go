package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"getclient/internal/config"
	"getclient/internal/store"
	"getclient/internal/ui"
)

func RunFromState(ctx context.Context, st store.State) int {
	cfg, err := configFromEnvAndState(st)
	if err != nil {
		fmt.Fprintln(os.Stdout, ui.Red(err.Error()))
		return 2
	}
	return RunWithConfig(ctx, cfg)
}

func configFromEnvAndState(st store.State) (config.Config, error) {
	appID := st.AppID
	appHash := strings.TrimSpace(st.AppHash)
	if appID == 0 || appHash == "" {
		return config.Config{}, fmt.Errorf("не заданы API_ID/API_HASH. Откройте меню → «Настройки приложения»")
	}

	accounts := st.Accounts
	if strings.TrimSpace(st.AccountsFile) != "" {
		if a2, err := loadAccountsJSON(st.AccountsFile); err == nil && len(a2) > 0 {
			accounts = a2
		}
	}

	return config.Config{
		AppID:        appID,
		AppHash:      appHash,
		Accounts:     toCfgAccounts(accounts),
		KeywordsFile:  st.KeywordsFile,
		StopwordsFile: st.StopwordsFile,
		UseRegex:     st.UseRegex,
		PollInterval: pollDuration(st),
		PollLimit:    st.PollLimit,
		BotToken:     st.BotToken,
		BotChatID:    st.BotChatID,
	}, nil
}

func loadAccountsJSON(path string) ([]store.Account, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a []store.Account
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return a, nil
}

func saveAccountsJSON(path string, a []store.Account) error {
	if path == "" {
		path = "accounts.json"
	}
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o700)
	}
	return os.WriteFile(path, data, 0o600)
}

func toCfgAccounts(a []store.Account) []config.Account {
	out := make([]config.Account, 0, len(a))
	for i, x := range a {
		name := strings.TrimSpace(x.Name)
		if name == "" {
			name = fmt.Sprintf("account-%d", i+1)
		}
		out = append(out, config.Account{Name: name, SessionPath: strings.TrimSpace(x.SessionPath)})
	}
	return out
}


