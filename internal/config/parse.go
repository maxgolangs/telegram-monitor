package config

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func Parse() (Config, error) {
	appIDFlag := flag.Int("app-id", 0, "App ID")
	appHashFlag := flag.String("app-hash", "", "App Hash")

	keywordsFile := flag.String("keywords-file", "keywords.txt", "File with phrases (one per line)")
	stopFile := flag.String("stopwords-file", "stopwords.txt", "File with stop-words (one per line)")
	useRegex := flag.Bool("regex", false, "Treat keywords as regular expressions")

	sessionPath := flag.String("session", "sessions/session.bin", "Session file (single-account fallback)")
	accountsFile := flag.String("accounts-file", "", "accounts.json for multi-account")

	pollInterval := flag.Duration("poll-interval", 0, "Dialogs polling fallback interval (0 = disabled)")
	pollLimit := flag.Int("poll-limit", 100, "Dialogs limit per poll")

	botToken := flag.String("bot-token", "", "Bot token")
	botChatID := flag.Int64("bot-chat-id", 0, "Bot chat id")

	flag.Parse()

	appID := *appIDFlag
	appHash := strings.TrimSpace(*appHashFlag)

	if appID == 0 || appHash == "" {
		return Config{}, fmt.Errorf("missing app credentials (API_ID/API_HASH)")
	}

	if *pollLimit <= 0 {
		return Config{}, fmt.Errorf("poll-limit must be > 0")
	}

	tok := strings.TrimSpace(*botToken)
	chatID := *botChatID

	accounts, err := LoadAccounts(*accountsFile, *sessionPath)
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppID:         appID,
		AppHash:       appHash,
		Accounts:      accounts,
		KeywordsFile:  strings.TrimSpace(*keywordsFile),
		StopwordsFile: strings.TrimSpace(*stopFile),
		UseRegex:      *useRegex,
		PollInterval:  *pollInterval,
		PollLimit:     *pollLimit,
		BotToken:      tok,
		BotChatID:     chatID,
	}, nil
}

var _ time.Duration


