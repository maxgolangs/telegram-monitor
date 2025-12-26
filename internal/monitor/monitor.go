package monitor

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"getclient/internal/notifier"
	"getclient/internal/store"
	"getclient/internal/telegramutil"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Monitor struct {
	matcher    *Matcher
	logger     *zap.Logger
	notify     notifier.Notifier
	account    string
	limiter    store.SenderLimiter
	globalSeen *sync.Map
}

func New(matcher *Matcher, logger *zap.Logger, notify notifier.Notifier, account string, limiter store.SenderLimiter, globalSeen *sync.Map) *Monitor {
	return &Monitor{
		matcher:    matcher,
		logger:     logger,
		notify:     notify,
		account:    account,
		limiter:    limiter,
		globalSeen: globalSeen,
	}
}

func (m *Monitor) ProcessMessage(ctx context.Context, e tg.Entities, msg tg.MessageClass) {
	message, ok := msg.(*tg.Message)
	if !ok || message == nil {
		return
	}
	m.process(ctx, e, message.PeerID, message.FromID, message.ID, message.Message)
}

func (m *Monitor) ProcessShort(ctx context.Context, e tg.Entities, peerID tg.PeerClass, fromID tg.PeerClass, msgID int, text string) {
	m.process(ctx, e, peerID, fromID, msgID, text)
}

func (m *Monitor) process(ctx context.Context, e tg.Entities, peerID tg.PeerClass, fromID tg.PeerClass, msgID int, text string) {
	if text == "" {
		return
	}

	isGroup := false
	switch p := peerID.(type) {
	case *tg.PeerChat:
		isGroup = true
	case *tg.PeerChannel:
		if ch, ok := e.Channels[p.ChannelID]; ok {
			if ch.Megagroup {
				isGroup = true
			}
		} else {
			isGroup = true
		}
	}

	if !isGroup {
		return
	}

	peerKey := telegramutil.PeerKey(peerID)
	dedupeKey := fmt.Sprintf("%s:%d", peerKey, msgID)
	if _, loaded := m.globalSeen.LoadOrStore(dedupeKey, struct{}{}); loaded {
		return
	}

	if !m.matcher.Match(text) {
		return
	}

	chatName := telegramutil.PeerTitle(peerID, e)
	var fromPeer tg.PeerClass = fromID
	if fromPeer == nil {
		fromPeer = &tg.PeerUser{UserID: 0}
	}
	sender := telegramutil.Sender(fromPeer, e)
	senderName := sender.Name
	if sender.Username != "" {
		senderName = "@" + sender.Username
	} else if senderName == "" && sender.ID != 0 {
		senderName = fmt.Sprintf("id:%d", sender.ID)
	}

	if m.limiter != nil {
		ok, err := m.limiter.Allow(ctx, m.account, sender.ID)
		if err != nil {
			m.logger.Warn("Limiter failed", zap.Error(err))
			return
		}
		if !ok {
			return
		}
	}

	link := telegramutil.MessageLink(peerID, msgID, e)

	m.logger.Info("Keyword found",
		zap.String("chat", chatName),
		zap.String("from", senderName),
		zap.String("account", m.account),
		zap.String("text", text),
	)

	fmt.Printf("\n\033[32m[ALERT]\033[0m Найдено аккаунтом: \033[1m%s\033[0m\n", m.account)
	fmt.Printf("Чат: \033[33m%s\033[0m\n", chatName)
	fmt.Printf("От: \033[36m%s\033[0m\n", senderName)
	if link != "" {
		fmt.Printf("Ссылка: \033[34m%s\033[0m\n", link)
	}
	fmt.Printf("Текст: %s\n\n", text)

	if m.notify != nil {
		if err := m.notify.Notify(ctx, notifier.Notification{
			ChatTitle: chatName,
			From:      fmt.Sprintf("%s (через %s)", senderName, m.account),
			Link:      link,
			Text:      text,
		}); err != nil {
			m.logger.Warn("Notify failed", zap.Error(err))
		}
	}
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}


