package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatID int64
}

func NewTelegramNotifier(token string, chatID int64) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize telegram bot: %w", err)
	}
	return &TelegramNotifier{
		bot:    bot,
		chatID: chatID,
	}, nil
}

func (tn *TelegramNotifier) SendAlert(errType string, count int, window string) error {
	text := fmt.Sprintf("🚨 *High Error Rate Detected*\n\n"+
		"*Type:* `%s`\n"+
		"*Count:* `%d` occurrences\n"+
		"*Window:* last %s", errType, count, window)

	msg := tgbotapi.NewMessage(tn.chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2 // Clean formatting for logs/code snippets

	_, err := tn.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}
