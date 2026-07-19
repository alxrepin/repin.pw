package telegram

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const botStatePath = "./var/telegram/bot/updates.json"

type BotHooks struct {
	OnPostsChanged func(ctx context.Context, ids []int) error
	OnPostsDeleted func(ctx context.Context, ids []int) error
	OnReady        func(ctx context.Context) error
}

type Bot struct {
	client  *Client
	channel string
}

func NewBot(client *Client, channel string) *Bot {
	return &Bot{client: client, channel: channel}
}

func (b *Bot) Run(ctx context.Context, hooks BotHooks) error {
	state, err := newUpdateState(botStatePath)
	if err != nil {
		return err
	}

	dispatcher := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler:      dispatcher,
		Storage:      state,
		AccessHasher: state,
	})

	b.client.updates = gaps

	return b.client.Run(ctx, func(ctx context.Context, tgClient *telegram.Client) error {
		log := zerolog.Ctx(ctx)

		channel, err := resolveChannel(ctx, tgClient, b.channel)
		if err != nil {
			return err
		}

		withSession := func(handlerCtx context.Context) context.Context {
			return context.WithValue(handlerCtx, sessionKey{}, tgClient)
		}

		report := func(event string, err error) {
			if err != nil {
				log.Error().Err(err).Str("event", event).Msg("handle update failed")
			}
		}

		dispatcher.OnNewChannelMessage(func(ctx context.Context, _ tg.Entities, u *tg.UpdateNewChannelMessage) error {
			if id, ok := channelMessageID(u.Message, channel.ID); ok {
				report("new", hooks.OnPostsChanged(withSession(ctx), []int{id}))
			}

			return nil
		})

		dispatcher.OnEditChannelMessage(func(ctx context.Context, _ tg.Entities, u *tg.UpdateEditChannelMessage) error {
			if id, ok := channelMessageID(u.Message, channel.ID); ok {
				report("edit", hooks.OnPostsChanged(withSession(ctx), []int{id}))
			}

			return nil
		})

		dispatcher.OnDeleteChannelMessages(func(ctx context.Context, _ tg.Entities, u *tg.UpdateDeleteChannelMessages) error {
			if u.ChannelID == channel.ID {
				report("delete", hooks.OnPostsDeleted(withSession(ctx), u.Messages))
			}

			return nil
		})

		self, err := tgClient.Self(ctx)
		if err != nil {
			return fmt.Errorf("get self: %w", err)
		}

		g, gctx := errgroup.WithContext(ctx)

		if hooks.OnReady != nil {
			g.Go(func() error { return hooks.OnReady(withSession(gctx)) })
		}

		g.Go(func() error {
			return gaps.Run(gctx, tgClient.API(), self.ID, updates.AuthOptions{
				IsBot: self.Bot,
				OnStart: func(context.Context) {
					log.Info().Str("channel", b.channel).Str("bot", self.Username).Msg("update loop started")
				},
			})
		})

		return g.Wait()
	})
}

func channelMessageID(msg tg.MessageClass, channelID int64) (int, bool) {
	m, ok := msg.(*tg.Message)
	if !ok {
		return 0, false
	}

	peer, ok := m.PeerID.(*tg.PeerChannel)
	if !ok || peer.ChannelID != channelID {
		return 0, false
	}

	return m.ID, true
}
