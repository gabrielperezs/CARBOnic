package main

import (
	"fmt"
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

type TelegramBOT struct {
	bot     *tgbotapi.BotAPI
	related []*Telegram
}

func (tb *TelegramBOT) connect(token string) error {

	var err error
	tb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Println("Telegram connection error:", token, err)
		return err
	}

	return nil
}

func (tb *TelegramBOT) listener() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tb.bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("ERROR telegram: %s", err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		for _, t := range tb.related {
			if t.Group == update.Message.Chat.ID {
				commands(t, update.Message.From.String(), update.Message.Text)
			}
		}
	}
}

func (tb *TelegramBOT) Send(group int64, message *Message) {

	if !ignoreDup(fmt.Sprintf("telegram:%d:%s", group, message.msg)) {
		log.Printf("Ignore repeated message on %s %d", "telegram", group)
		return
	}

	msg := tgbotapi.NewMessage(group, message.msg)
	tb.bot.Send(msg)
}
