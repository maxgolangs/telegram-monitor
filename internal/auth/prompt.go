package authutil

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type PromptAuth struct {
	In  *bufio.Reader
	Out io.Writer
	Tag string
}

func (p PromptAuth) prompt(line string) (string, error) {
	if p.Tag != "" {
		line = "[" + p.Tag + "] " + line
	}
	fmt.Fprintln(p.Out, line)
	fmt.Fprint(p.Out, "> ")
	s, err := p.In.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(s), nil
}

func (p PromptAuth) Phone(ctx context.Context) (string, error) {
	return p.prompt("Введите номер телефона (с кодом страны, например +79991234567):")
}

func (p PromptAuth) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	if sentCode != nil {
		method := "неизвестно"
		switch sentCode.Type.(type) {
		case *tg.AuthSentCodeTypeApp:
			method = "в Telegram (чат «Telegram»)"
		case *tg.AuthSentCodeTypeSMS:
			method = "SMS"
		case *tg.AuthSentCodeTypeCall:
			method = "звонок"
		case *tg.AuthSentCodeTypeFlashCall:
			method = "flash call"
		case *tg.AuthSentCodeTypeMissedCall:
			method = "пропущенный звонок"
		case *tg.AuthSentCodeTypeEmailCode:
			method = "код на email"
		}
		fmt.Fprintf(p.Out, "Код отправлен: %s\n", method)
		if sentCode.Timeout != 0 {
			fmt.Fprintf(p.Out, "Возможно нужно подождать до %d секунд.\n", sentCode.Timeout)
		}
	}
	return p.prompt("Введите код:")
}

func (p PromptAuth) Password(ctx context.Context) (string, error) {
	return p.prompt("Введите пароль 2FA (если включён):")
}

func (p PromptAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

func (p PromptAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("регистрация не поддерживается")
}


