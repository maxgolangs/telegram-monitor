package monitor

import (
	"context"
	"time"

	"getclient/internal/telegramutil"

	"github.com/gotd/td/tg"
)

type DialogPoller struct {
	api      *tg.Client
	monitor  *Monitor
	interval time.Duration
	limit    int
}

func NewDialogPoller(api *tg.Client, monitor *Monitor, interval time.Duration, limit int) *DialogPoller {
	return &DialogPoller{api: api, monitor: monitor, interval: interval, limit: limit}
}

func (p *DialogPoller) Run(ctx context.Context) {
	t := time.NewTicker(p.interval)
	defer t.Stop()

	p.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.tick(ctx)
		}
	}
}

func (p *DialogPoller) tick(ctx context.Context) {
	resp, err := p.api.MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{
		OffsetPeer: &tg.InputPeerEmpty{},
		Limit:      p.limit,
	})
	if err != nil {
		return
	}

	var users []tg.UserClass
	var chats []tg.ChatClass
	var messages []tg.MessageClass

	switch d := resp.(type) {
	case *tg.MessagesDialogs:
		users = d.Users
		chats = d.Chats
		messages = d.Messages
	case *tg.MessagesDialogsSlice:
		users = d.Users
		chats = d.Chats
		messages = d.Messages
	default:
		return
	}

	e := telegramutil.BuildEntities(users, chats)
	for _, msg := range messages {
		p.monitor.ProcessMessage(ctx, e, msg)
	}
}


