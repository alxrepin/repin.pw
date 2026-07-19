package telegram

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Authenticator struct {
	phone string
}

func (a Authenticator) Phone(_ context.Context) (string, error) {
	return a.phone, nil
}

func (Authenticator) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")

	var code string
	if _, err := fmt.Scanln(&code); err != nil {
		return "", err
	}

	return code, nil
}

func (Authenticator) Password(_ context.Context) (string, error) {
	fmt.Print("Enter 2FA password: ")

	pass, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(pass), nil
}

func (Authenticator) AcceptTermsOfService(_ context.Context, _ tg.HelpTermsOfService) error {
	return nil
}

func (Authenticator) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("sign up is not supported")
}
