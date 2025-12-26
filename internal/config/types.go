package config

import "time"

type Account struct {
	Name        string
	SessionPath string
}

type Config struct {
	AppID   int
	AppHash string

	Accounts []Account

	KeywordsFile  string
	StopwordsFile string
	UseRegex      bool

	PollInterval time.Duration
	PollLimit    int

	BotToken  string
	BotChatID int64
}


