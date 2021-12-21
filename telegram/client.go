package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Client struct {
	b      *tgbotapi.BotAPI
	chatID int64
}

func NewClient(telegramToken string, chatID int64) (*Client, error) {
	b, err := tgbotapi.NewBotAPI(telegramToken)
	if err != nil {
		return nil, fmt.Errorf("unable to create new bot api: %w", err)
	}
	b.Debug = true

	return &Client{
		b:      b,
		chatID: chatID,
	}, nil
}

func (c *Client) SendMsg(text string) error {
	msg := tgbotapi.NewMessage(c.chatID, text)
	msg.ParseMode = "HTML"
	if _, err := c.b.Send(msg); err != nil {
		return err
	}

	return nil
}

func (c *Client) NewUpdates() (tgbotapi.UpdatesChannel, error) {
	updatesCh, err := c.b.GetUpdatesChan(tgbotapi.UpdateConfig{Timeout: 60})
	if err != nil {
		return nil, fmt.Errorf("unable to acquire updates channel: %w", err)
	}

	return updatesCh, nil
}
