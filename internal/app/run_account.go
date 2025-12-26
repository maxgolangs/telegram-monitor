package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	authutil "getclient/internal/auth"
	"getclient/internal/config"
	"getclient/internal/monitor"
	"getclient/internal/notifier"
	"getclient/internal/store"
	"getclient/internal/telegramutil"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"sync"
)

func runAccount(ctx context.Context, cfg config.Config, acc config.Account, n notifier.Notifier, limiter store.SenderLimiter, globalSeen *sync.Map, logger *zap.Logger) error {
	if acc.SessionPath != "" {
		if err := os.MkdirAll(filepath.Dir(acc.SessionPath), 0o700); err != nil {
			return fmt.Errorf("failed to create session dir (%s): %w", acc.Name, err)
		}
	}

	keywords, stopwords := mustReadWords(cfg.KeywordsFile, cfg.StopwordsFile)
	logger.Info("Загружены слова", 
		zap.Int("keywords", len(keywords)), 
		zap.Int("stopwords", len(stopwords)),
		zap.String("account", acc.Name),
	)
	matcher := monitor.NewMatcher(keywords, stopwords, cfg.UseRegex)
	mon := monitor.New(matcher, logger, n, acc.Name, limiter, globalSeen)

	dispatcher := tg.NewUpdateDispatcher()

	updatesMgr := updates.New(updates.Config{
		Handler: dispatcher,
		Logger:  zap.NewNop(),
	})

	rawHandler := telegram.UpdateHandlerFunc(func(ctx context.Context, u tg.UpdatesClass) error {
		go func(u tg.UpdatesClass) {
			switch v := u.(type) {
			case *tg.UpdateShortMessage:
				peer := &tg.PeerUser{UserID: v.UserID}
				mon.ProcessShort(ctx, tg.Entities{}, peer, peer, v.ID, v.Message)
			case *tg.UpdateShortChatMessage:
				peer := &tg.PeerChat{ChatID: v.ChatID}
				from := &tg.PeerUser{UserID: v.FromID}
				mon.ProcessShort(ctx, tg.Entities{}, peer, from, v.ID, v.Message)
			case *tg.Updates:
				entities := telegramutil.BuildEntities(v.Users, v.Chats)
				for _, sub := range v.Updates {
					switch us := sub.(type) {
					case *tg.UpdateNewMessage:
						mon.ProcessMessage(ctx, entities, us.Message)
					case *tg.UpdateNewChannelMessage:
						mon.ProcessMessage(ctx, entities, us.Message)
					case *tg.UpdateEditMessage:
						mon.ProcessMessage(ctx, entities, us.Message)
					case *tg.UpdateEditChannelMessage:
						mon.ProcessMessage(ctx, entities, us.Message)
					}
				}
			}
		}(u)

		return updatesMgr.Handle(ctx, u)
	})

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		Logger: zap.NewNop(),
		SessionStorage: &session.FileStorage{
			Path: acc.SessionPath,
		},
		UpdateHandler: rawHandler,
	})

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		mon.ProcessMessage(ctx, e, u.Message)
		return nil
	})
	dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
		mon.ProcessMessage(ctx, e, u.Message)
		return nil
	})
	dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditMessage) error {
		mon.ProcessMessage(ctx, e, u.Message)
		return nil
	})
	dispatcher.OnEditChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditChannelMessage) error {
		mon.ProcessMessage(ctx, e, u.Message)
		return nil
	})

	pa := authutil.PromptAuth{In: bufio.NewReader(os.Stdin), Out: os.Stdout, Tag: acc.Name}
	flow := auth.NewFlow(pa, auth.SendCodeOptions{})

	logger.Info("Подключение к Telegram...", zap.String("account", acc.Name))
	return client.Run(ctx, func(ctx context.Context) error {
		for {
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				logger.Error("Ошибка авторизации", zap.String("account", acc.Name), zap.Error(err))
				fmt.Fprintf(os.Stdout, "[%s] Ошибка авторизации. Повторить? (y/N)\n", acc.Name)
				fmt.Fprint(os.Stdout, "> ")
				var ans string
				_, _ = fmt.Scanln(&ans)
				if ans != "y" && ans != "Y" {
					return err
				}
				continue
			}
			break
		}

		logger.Info("Успешно подключено", zap.String("account", acc.Name))
		api := client.API()
		self, err := client.Self(ctx)
		if err != nil {
			return err
		}

		if cfg.PollInterval > 0 {
			go monitor.NewDialogPoller(api, mon, cfg.PollInterval, cfg.PollLimit).Run(ctx)
		}

		logger.Info("Мониторинг запущен. Нажмите Ctrl+C для остановки.", zap.String("account", acc.Name))
		return updatesMgr.Run(ctx, api, self.ID, updates.AuthOptions{})
	})
}


