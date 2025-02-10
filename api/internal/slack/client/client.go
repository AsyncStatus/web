package slack

import (
	"api/config"
	"context"

	"github.com/slack-go/slack"
)

type Client struct {
	client *slack.Client
	cfg    *config.Config
}

func NewClient(cfg *config.Config) *Client {
	return &Client{slack.New(cfg.SlackClientID), cfg}
}
func (c *Client) PostToStatusChannel(ctx context.Context, text string) error {
	return c.postMessage(ctx, StatusChannel, text)
}
func (c *Client) postMessage(ctx context.Context, channel *Channel, text string) error {
	if _, _, err := c.client.PostMessageContext(
		ctx,
		channel.ToString(c.cfg.AppEnv),
		slack.MsgOptionText(text, false),
	); err != nil {
		return err
	}

	return nil
}
