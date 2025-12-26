package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	"getclient/internal/config"
	"getclient/internal/notifier"
	"getclient/internal/store"
	"getclient/internal/ui"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func RunNonInteractive(ctx context.Context) int {
	cfg, err := config.Parse()
	if err != nil {
		fmt.Fprintln(os.Stdout, err.Error())
		ui.WaitEnter()
		return 2
	}
	res := RunWithConfig(ctx, cfg)
	if res != 0 {
		ui.WaitEnter()
	}
	return res
}

func RunWithConfig(ctx context.Context, cfg config.Config) int {
	loggerCfg := zap.NewDevelopmentConfig()
	loggerCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	loggerCfg.DisableStacktrace = true
	loggerCfg.DisableCaller = true
	loggerCfg.EncoderConfig.TimeKey = ""
	logger, _ := loggerCfg.Build()
	defer logger.Sync()

	bot := notifier.NewTelegramBot(cfg.BotToken, cfg.BotChatID)
	var n notifier.Notifier
	if bot.Enabled() {
		n = bot
	}
	logger.Info("Бот-уведомления",
		zap.Bool("enabled", bot.Enabled()),
		zap.Int64("chat_id", cfg.BotChatID),
	)

	db, err := store.OpenBaseDB("data/base.json")
	if err != nil {
		logger.Error("Base error", zap.Error(err))
		return 2
	}
	defer db.Close()

	var globalSeen sync.Map

	g, ctx := errgroup.WithContext(ctx)
	for _, acc := range cfg.Accounts {
		acc := acc
		g.Go(func() error {
			return runAccount(ctx, cfg, acc, n, db, &globalSeen, logger)
		})
	}

	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			return 0
		}
		logger.Error("Client error", zap.Error(err))
		return 1
	}
	return 0
}



