package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

const (
	userSessionPath   = "./var/telegram/session.json"
	botSessionPath    = "./var/telegram/bot/session.json"
	workerSessionPath = "./var/telegram/worker/session.json"
)

type Client struct {
	apiID       int
	apiHash     string
	phone       string
	botToken    string
	proxyURL    string
	sessionPath string
	updates     telegram.UpdateHandler
}

func NewClient(apiID int, apiHash, phone, proxyURL string) *Client {
	return &Client{apiID: apiID, apiHash: apiHash, phone: phone, proxyURL: proxyURL, sessionPath: userSessionPath}
}

func NewBotClient(apiID int, apiHash, botToken, proxyURL string) *Client {
	return &Client{apiID: apiID, apiHash: apiHash, botToken: botToken, proxyURL: proxyURL, sessionPath: botSessionPath}
}

func NewWorkerClient(apiID int, apiHash, phone, botToken, proxyURL string) *Client {
	c := &Client{apiID: apiID, apiHash: apiHash, phone: phone, proxyURL: proxyURL, sessionPath: userSessionPath}

	if botToken != "" {
		c.botToken = botToken
		c.sessionPath = workerSessionPath
	}

	return c
}

func (c *Client) Run(ctx context.Context, fn func(ctx context.Context, client *telegram.Client) error) error {
	if err := os.MkdirAll(filepath.Dir(c.sessionPath), 0o700); err != nil {
		return fmt.Errorf("create session dir: %w", err)
	}

	opts := telegram.Options{
		SessionStorage: &session.FileStorage{Path: c.sessionPath},
		UpdateHandler:  c.updates,
	}

	if c.proxyURL != "" {
		dial, err := proxyDial(c.proxyURL)
		if err != nil {
			return err
		}

		opts.Resolver = dcs.Plain(dcs.PlainOptions{Dial: dial})
	}

	client := telegram.NewClient(c.apiID, c.apiHash, opts)

	return client.Run(ctx, func(ctx context.Context) error {
		if err := c.authorize(ctx, client); err != nil {
			return fmt.Errorf("auth failed: %w", err)
		}

		return fn(ctx, client)
	})
}

func (c *Client) authorize(ctx context.Context, client *telegram.Client) error {
	if c.botToken == "" {
		flow := auth.NewFlow(Authenticator{phone: c.phone}, auth.SendCodeOptions{})
		return client.Auth().IfNecessary(ctx, flow)
	}

	status, err := client.Auth().Status(ctx)
	if err != nil {
		return fmt.Errorf("auth status: %w", err)
	}

	if status.Authorized {
		return nil
	}

	if _, err := client.Auth().Bot(ctx, c.botToken); err != nil {
		return fmt.Errorf("bot login: %w", err)
	}

	return nil
}

type sessionKey struct{}

func (c *Client) WithSession(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.Run(ctx, func(ctx context.Context, client *telegram.Client) error {
		return fn(context.WithValue(ctx, sessionKey{}, client))
	})
}

func clientFrom(ctx context.Context) (*telegram.Client, bool) {
	client, ok := ctx.Value(sessionKey{}).(*telegram.Client)
	return client, ok
}

func resolveChannel(ctx context.Context, client *telegram.Client, username string) (*tg.Channel, error) {
	resolve, err := client.API().ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: username})
	if err != nil {
		return nil, fmt.Errorf("resolve username failed: %w", err)
	}

	if len(resolve.Chats) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	channel, ok := resolve.Chats[0].(*tg.Channel)
	if !ok {
		return nil, fmt.Errorf("not a channel")
	}

	return channel, nil
}

func (c *Client) FetchMessages(ctx context.Context, client *telegram.Client, username string, minID int) ([]tg.MessageClass, error) {
	channel, err := resolveChannel(ctx, client, username)
	if err != nil {
		return nil, err
	}

	var (
		all      []tg.MessageClass
		offsetID int
	)

	for {
		resp, err := client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
			Peer:     &tg.InputPeerChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash},
			OffsetID: offsetID,
			MinID:    minID,
			Limit:    100,
		})
		if err != nil {
			return nil, fmt.Errorf("get history failed: %w", err)
		}

		msgs, err := messagesOf(resp)
		if err != nil {
			return nil, err
		}

		if len(msgs) == 0 {
			break
		}

		all = append(all, msgs...)

		last, ok := msgs[len(msgs)-1].(*tg.Message)
		if !ok {
			break
		}

		offsetID = last.ID
	}

	return all, nil
}

func (c *Client) FetchMessage(ctx context.Context, client *telegram.Client, username string, id int) (*tg.Message, error) {
	channel, err := resolveChannel(ctx, client, username)
	if err != nil {
		return nil, err
	}

	resp, err := client.API().ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash},
		ID:      []tg.InputMessageClass{&tg.InputMessageID{ID: id}},
	})
	if err != nil {
		return nil, fmt.Errorf("get message failed: %w", err)
	}

	msgs, err := messagesOf(resp)
	if err != nil {
		return nil, err
	}

	if len(msgs) == 0 {
		return nil, fmt.Errorf("message %d not found", id)
	}

	msg, ok := msgs[0].(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected message type")
	}

	return msg, nil
}

func (c *Client) FetchMessagesByIDs(ctx context.Context, client *telegram.Client, username string, ids []int) ([]tg.MessageClass, error) {
	channel, err := resolveChannel(ctx, client, username)
	if err != nil {
		return nil, err
	}

	const chunkSize = 100 // channels.getMessages accepts at most 100 ids per call

	var all []tg.MessageClass

	for start := 0; start < len(ids); start += chunkSize {
		end := min(start+chunkSize, len(ids))

		input := make([]tg.InputMessageClass, 0, end-start)
		for _, id := range ids[start:end] {
			input = append(input, &tg.InputMessageID{ID: id})
		}

		resp, err := client.API().ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{ChannelID: channel.ID, AccessHash: channel.AccessHash},
			ID:      input,
		})
		if err != nil {
			return nil, fmt.Errorf("get messages failed: %w", err)
		}

		msgs, err := messagesOf(resp)
		if err != nil {
			return nil, err
		}

		all = append(all, msgs...)
	}

	return all, nil
}

func (c *Client) FetchChannelInfo(ctx context.Context, client *telegram.Client, username string) (*tg.Channel, string, int64, error) {
	channel, err := resolveChannel(ctx, client, username)
	if err != nil {
		return nil, "", 0, err
	}

	full, err := client.API().ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	})
	if err != nil {
		return nil, "", 0, fmt.Errorf("get full channel failed: %w", err)
	}

	cf, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, "", 0, fmt.Errorf("unexpected full channel type")
	}

	subscribers, _ := cf.GetParticipantsCount()

	return channel, cf.GetAbout(), int64(subscribers), nil
}

func messagesOf(resp tg.MessagesMessagesClass) ([]tg.MessageClass, error) {
	switch m := resp.(type) {
	case *tg.MessagesMessages:
		return m.Messages, nil
	case *tg.MessagesChannelMessages:
		return m.Messages, nil
	default:
		return nil, fmt.Errorf("unexpected messages response type")
	}
}
