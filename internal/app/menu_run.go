package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"getclient/internal/store"
	"getclient/internal/ui"
)

const statePath = "data/config.json"

func Run(ctx context.Context) int {
	for _, a := range os.Args[1:] {
		if a == "--no-menu" {
			return RunNonInteractive(ctx)
		}
	}
	return RunMenu(ctx)
}

func ensureFile(path, content string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.WriteFile(path, []byte(content), 0o600)
	}
}

func RunMenu(ctx context.Context) int {
	m := ui.New()

	if err := os.MkdirAll("data/sessions", 0o700); err != nil {
		fmt.Fprintf(os.Stdout, "Ошибка создания папки data: %v\n", err)
	}
	ensureFile("data/keywords.txt", "")
	ensureFile("data/stopwords.txt", "")
	ensureFile("data/accounts.json", "[]")
	ensureFile("data/base.json", "{}")

	st, err := store.Load(statePath)
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		return 2
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		_ = store.Save(statePath, st)
	}

	if st.PollIntervalMs == 0 {
		st.PollIntervalMs = 3000
	}

	for {
		kw, sw := mustReadWords(st.KeywordsFile, st.StopwordsFile)
		botStatus := ui.Red("выключен")
		if st.BotToken != "" && st.BotChatID != 0 {
			botStatus = ui.Green("включен")
		}
		
		info := fmt.Sprintf("Аккаунтов: %s | Фраз: %s | Стоп-слов: %s | Бот: %s",
			ui.Cyan(fmt.Sprintf("%d", len(st.Accounts))),
			ui.Cyan(fmt.Sprintf("%d", len(kw))),
			ui.Cyan(fmt.Sprintf("%d", len(sw))),
			botStatus,
		)

		act, err := m.Choose(ctx, info)
		if err != nil {
			fmt.Fprintln(os.Stdout, err.Error())
			return 2
		}
		switch act {
		case ui.ActionExit:
			_ = store.Save(statePath, st)
			return 0
		case ui.ActionBotSettings:
			if err := menuBotSettings(m, &st); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionAppSettings:
			if err := menuAppSettings(m, &st); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionKeywordsAdd:
			if err := appendLine(m, st.KeywordsFile, "Ключевая фраза"); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionStopwordsAdd:
			if err := appendLine(m, st.StopwordsFile, "Стоп-слово"); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionAddAccount:
			if err := menuAddAccount(ctx, m, &st); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionRemoveAccount:
			if err := menuRemoveAccount(m, &st); err != nil {
				m.Linef("Ошибка: %v", err)
			}
		case ui.ActionResetBase:
			_ = os.Remove("data/base.json")
			m.Linef("%s", ui.Green("База сброшена!"))
			time.Sleep(1 * time.Second)
		case ui.ActionStart:
			_ = store.Save(statePath, st)
			runMonitoringSession(ctx, m, st)
			continue
		}
		_ = store.Save(statePath, st)
	}
}

func runMonitoringSession(parent context.Context, m *ui.Menu, st store.State) {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()

	ui.Clear()
	m.Title("Мониторинг")
	m.Linef("%s", ui.Green("Мониторинг запущен."))
	m.Linef("Чтобы остановить мониторинг: введите %s и нажмите Enter.", ui.Cyan("три пробела"))
	m.Linef("--------------------------------------------------")

	stopCh := make(chan struct{}, 1)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			s, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if ctx.Err() != nil {
				select {
				case stopCh <- struct{}{}:
				default:
				}
				return
			}
			if strings.Trim(s, "\r\n") == "   " {
				select {
				case stopCh <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	done := make(chan struct{}, 1)
	go func() {
		_ = RunFromState(ctx, st)
		done <- struct{}{}
	}()

	select {
	case <-parent.Done():
		return
	case <-stopCh:
		cancel()
		<-done
		fmt.Printf("\n%s\n", ui.Green("Остановка..."))
		time.Sleep(800 * time.Millisecond)
	case <-done:
		fmt.Printf("\n%s\n", ui.Red("Сессия завершена."))
		fmt.Print("Нажмите Enter, чтобы вернуться в меню...")
		cancel()
		<-stopCh
		time.Sleep(500 * time.Millisecond)
	}
}

func menuAppSettings(m *ui.Menu, st *store.State) error {
	m.Title("Настройки приложения (Telegram API)")
	m.Linef("API_ID/API_HASH можно получить на https://my.telegram.org")
	idStr, err := m.Prompt("API_ID (пусто = оставить как есть)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(idStr) != "" {
		var id int
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id == 0 {
			return fmt.Errorf("неверный API_ID")
		}
		st.AppID = id
	}
	hash, err := m.Prompt("API_HASH (пусто = оставить как есть)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(hash) != "" {
		st.AppHash = strings.TrimSpace(hash)
	}
	return nil
}

func menuBotSettings(m *ui.Menu, st *store.State) error {
	m.Title("Настройки бота")
	tok, err := m.Prompt("BOT_TOKEN (пусто = оставить как есть)")
	if err != nil {
		return err
	}
	if strings.TrimSpace(tok) != "" {
		st.BotToken = strings.TrimSpace(tok)
	}
	chatID, err := m.PromptInt64("BOT_CHAT_ID (пусто = оставить как есть)")
	if err != nil {
		return err
	}
	if chatID != 0 {
		st.BotChatID = chatID
	}
	pollMs, err := m.PromptInt64("Интервал polling (мс). 0 = выключить (только realtime)")
	if err != nil {
		return err
	}
	if pollMs >= 0 {
		st.PollIntervalMs = pollMs
	}
	return nil
}

func appendLine(m *ui.Menu, filePath, label string) error {
	v, err := m.Prompt(label)
	if err != nil {
		return err
	}
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil && filepath.Dir(filePath) != "." {
		return err
	}
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintln(f, v)
	return err
}

func pollDuration(st store.State) time.Duration {
	if st.PollIntervalMs <= 0 {
		return 0
	}
	return time.Duration(st.PollIntervalMs) * time.Millisecond
}


