package main

import (
	"log"
	"sync"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	mTelegram       = &sync.Mutex{}
	telegramClients = make(map[string]*TelegramBOT)
)

func getTelegram(g *Group) {

	if g.Telegram == nil {
		log.Printf("[%s] Telegram disabled", g.Name)
		return
	}

	if g.Telegram.Token == "" {
		log.Printf("[%s] Telegram disabled, empty token", g.Name)
		return
	}

	mTelegram.Lock()
	defer mTelegram.Unlock()

	t := g.Telegram

	if _, ok := telegramClients[t.Token]; ok {
		t.Bot = telegramClients[t.Token]
		t.Bot.related = append(t.Bot.related, t)
	} else {
		t.Bot = &TelegramBOT{}
		err := g.Telegram.Bot.connect(t.Token)
		if err != nil {
			return
		}
		t.Bot.related = append(t.Bot.related, t)
		go g.Telegram.Bot.listener()
	}

	t.ParentGroup = g

	telegramClients[t.Token] = t.Bot

	t.start()

	return
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

func (tb *TelegramBOT) Send(msg tgbotapi.MessageConfig) {
	tb.bot.Send(msg)
}
