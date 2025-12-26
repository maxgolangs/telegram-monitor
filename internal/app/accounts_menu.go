package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"getclient/internal/store"
	"getclient/internal/ui"
)

var safeName = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func menuAddAccount(ctx context.Context, m *ui.Menu, st *store.State) error {
	m.Title("Добавить аккаунт")
	name, err := m.Prompt("Название аккаунта (например acc1)")
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("пустое название")
	}
	name = safeName.ReplaceAllString(name, "_")

	for _, a := range st.Accounts {
		if a.Name == name {
			return fmt.Errorf("аккаунт %q уже существует", name)
		}
	}

	sessionPath := filepath.Join("data", "sessions", name+".bin")
	ok, err := m.Confirm(fmt.Sprintf("Войти сейчас и создать session-файл %q?", sessionPath))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := loginAndSaveSession(ctx, name, sessionPath); err != nil {
		return err
	}

	st.Accounts = append(st.Accounts, store.Account{Name: name, SessionPath: sessionPath})
	return saveAccountsJSON(st.AccountsFile, st.Accounts)
}

func menuRemoveAccount(m *ui.Menu, st *store.State) error {
	m.Title("Удалить аккаунт")
	if len(st.Accounts) == 0 {
		m.Linef("Аккаунтов нет.")
		return nil
	}
	for i, a := range st.Accounts {
		m.Linef("%d) %s (%s)", i+1, a.Name, a.SessionPath)
	}
	s, err := m.Prompt("Введите номер аккаунта для удаления")
	if err != nil {
		return err
	}
	idx := 0
	_, _ = fmt.Sscanf(s, "%d", &idx)
	if idx < 1 || idx > len(st.Accounts) {
		return fmt.Errorf("неверный выбор")
	}
	acc := st.Accounts[idx-1]
	ok, err := m.Confirm(fmt.Sprintf("Удалить аккаунт %q?", acc.Name))
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	st.Accounts = append(st.Accounts[:idx-1], st.Accounts[idx:]...)
	_ = saveAccountsJSON(st.AccountsFile, st.Accounts)

	del, err := m.Confirm("Также удалить session-файл?")
	if err != nil {
		return err
	}
	if del {
		_ = os.Remove(acc.SessionPath)
	}
	return nil
}


