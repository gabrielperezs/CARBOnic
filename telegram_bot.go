package main

import (
	"fmt"
	"log"

	"gopkg.in/telegram-bot-api.v4"
)

func newTelegramBOT(token string) *TelegramBOT {
	b := &TelegramBOT{}

	err := b.connect(token)
	if err != nil {
		log.Println("Telegram error", err)
		return nil
	}

	go b.listener()

	return b

}

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

	if debug {
		tb.bot.Debug = true
	}
	log.Printf("Telegram Authorized on account %s", tb.bot.Self.UserName)

	return nil
}

func (tb *TelegramBOT) listener() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tb.bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("ERROR telegram: %s", err)
		return
	}

	if debug {
		log.Printf("Telegram start listen %s", tb.bot.Self.UserName)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		group := false

		for _, t := range tb.related {
			if t.Group == update.Message.Chat.ID {
				commands(t, update.Message.From.String(), update.Message.Text)
				group = true
			}
		}

		if !group {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, i'm alive. Please use the groups to send commands.")
			tb.bot.Send(msg)
		}
	}

	if debug {
		log.Printf("***** Telegram listener ended %s", tb.bot.Self.UserName)
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
