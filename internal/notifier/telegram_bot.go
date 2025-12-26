package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Notification struct {
	ChatTitle string
	From      string
	Link      string
	Text      string
}

type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}

type TelegramBot struct {
	token  string
	chatID int64
	http   *http.Client
}

func NewTelegramBot(token string, chatID int64) *TelegramBot {
	return &TelegramBot{
		token:  strings.TrimSpace(token),
		chatID: chatID,
		http: &http.Client{
			Timeout: 7 * time.Second,
		},
	}
}

func (b *TelegramBot) Enabled() bool {
	return b != nil && b.token != "" && b.chatID != 0
}

func (b *TelegramBot) Notify(ctx context.Context, n Notification) error {
	if !b.Enabled() {
		return nil
	}

	text, parseMode := format(n)
	form := url.Values{}
	form.Set("chat_id", fmt.Sprintf("%d", b.chatID))
	form.Set("text", text)
	if parseMode != "" {
		form.Set("parse_mode", parseMode)
	}
	form.Set("disable_web_page_preview", "true")

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)

	var lastErr error
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := b.http.Do(req)
		if err == nil {
			defer resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("status %s, body: %s", resp.Status, string(body))
		} else {
			lastErr = err
		}

		time.Sleep(time.Second * time.Duration(i+1))
	}

	return fmt.Errorf("after 3 attempts: %w", lastErr)
}

func format(n Notification) (string, string) {
	msg := strings.TrimSpace(n.Text)
	from := strings.TrimSpace(n.From)

	if n.Link != "" && msg != "" {
		linked := fmt.Sprintf(`<a href="%s">%s</a>`, htmlEscape(n.Link), htmlEscape(msg))
		if from != "" {
			return linked + "\n" + htmlEscape(from), "HTML"
		}
		return linked, "HTML"
	}

	if from != "" && msg != "" {
		return msg + "\n" + from, ""
	}
	if msg != "" {
		return msg, ""
	}
	return from, ""
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}


