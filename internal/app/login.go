package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	authutil "getclient/internal/auth"
	"getclient/internal/store"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"go.uber.org/zap"
)

func loginAndSaveSession(ctx context.Context, accountName string, sessionPath string) error {
	st, _ := store.Load(statePath)
	appID := st.AppID
	appHash := strings.TrimSpace(st.AppHash)
	if appID == 0 || appHash == "" {
		return fmt.Errorf("не заданы API_ID/API_HASH. Откройте меню → «Настройки приложения»")
	}
	if err := os.MkdirAll(filepath.Dir(sessionPath), 0o700); err != nil {
		return err
	}

	client := telegram.NewClient(appID, appHash, telegram.Options{
		Logger: zap.NewNop(),
		SessionStorage: &session.FileStorage{
			Path: sessionPath,
		},
		NoUpdates: true,
	})

	pa := authutil.PromptAuth{In: bufio.NewReader(os.Stdin), Out: os.Stdout, Tag: accountName}
	flow := auth.NewFlow(pa, auth.SendCodeOptions{})

	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}
		return nil
	})
}


