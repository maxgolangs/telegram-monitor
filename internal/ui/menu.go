package ui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Menu struct {
	In  *bufio.Reader
	Out io.Writer
}

func New() *Menu {
	return &Menu{In: bufio.NewReader(os.Stdin), Out: os.Stdout}
}

func (m *Menu) Title(s string) {
	fmt.Fprintf(m.Out, "\n\033[1m%s\033[0m\n", s)
}

func (m *Menu) Linef(format string, a ...any) {
	fmt.Fprintf(m.Out, format+"\n", a...)
}

func (m *Menu) Prompt(label string) (string, error) {
	fmt.Fprintf(m.Out, "\n\033[36m%s\033[0m\n> ", label)
	s, err := m.In.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func (m *Menu) PromptInt64(label string) (int64, error) {
	s, err := m.Prompt(label)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("неверное число")
	}
	return v, nil
}

func (m *Menu) Confirm(label string) (bool, error) {
	s, err := m.Prompt(label + " (y/N)")
	if err != nil {
		return false, err
	}
	return strings.ToLower(s) == "y", nil
}

type Action int

const (
	ActionExit Action = iota
	ActionStart
	ActionAddAccount
	ActionRemoveAccount
	ActionAppSettings
	ActionBotSettings
	ActionKeywordsAdd
	ActionStopwordsAdd
	ActionResetBase
)

func (m *Menu) Choose(ctx context.Context, info string) (Action, error) {
	Clear()
	if info != "" {
		fmt.Fprintln(m.Out, info)
	}
	m.Title("Telegram монитор")
	m.Linef("1) Запустить мониторинг")
	m.Linef("2) Добавить аккаунт")
	m.Linef("3) Удалить аккаунт")
	m.Linef("4) Настройки приложения (API_ID/API_HASH)")
	m.Linef("5) Настройки бота")
	m.Linef("6) Добавить ключевую фразу")
	m.Linef("7) Добавить стоп-слово")
	m.Linef("8) Сбросить базу (лимит 24ч)")
	m.Linef("0) Выход")
	s, err := m.Prompt("Выберите пункт меню")
	if err != nil {
		return ActionExit, err
	}
	select {
	case <-ctx.Done():
		return ActionExit, ctx.Err()
	default:
	}
	switch s {
	case "1":
		return ActionStart, nil
	case "2":
		return ActionAddAccount, nil
	case "3":
		return ActionRemoveAccount, nil
	case "4":
		return ActionAppSettings, nil
	case "5":
		return ActionBotSettings, nil
	case "6":
		return ActionKeywordsAdd, nil
	case "7":
		return ActionStopwordsAdd, nil
	case "8":
		return ActionResetBase, nil
	default:
		return ActionExit, nil
	}
}


