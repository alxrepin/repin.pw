package telegram

import (
	"github.com/gotd/td/telegram/dcs"

	"repin/internal/pkg/proxyx"
)

func proxyDial(rawURL string) (dcs.DialFunc, error) {
	dial, err := proxyx.Dialer(rawURL)
	if err != nil {
		return nil, err
	}

	return dcs.DialFunc(dial), nil
}
