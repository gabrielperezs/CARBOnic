package main

import "gopkg.in/telegram-bot-api.v4"
import "log"

type Telegram struct {
	Token string
	Name  string

	Group int64

	Bot *tgbotapi.BotAPI

	MinScore int
	chSender chan *Message

	ParentGroup *Group
}

func (t *Telegram) start() {
	t.chSender = make(chan *Message, 5)
	go t.sender()
	go t.receiver()
}

func (t *Telegram) sender() {

	for {
		if t.Bot == nil {
			log.Println("Telegram was not correctly initialiced", t.Group, t.Name)
			getTelegram(t.ParentGroup)
			return
		}

		alarm := <-t.chSender

		msg := tgbotapi.NewMessage(t.Group, alarm.msg)
		t.Bot.Send(msg)
	}
}

func (t *Telegram) receiver() {
	t.Bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := t.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		commands(t, update.Message.From.String(), update.Message.Text)
	}
}

func (t *Telegram) getParentGroup() *Group {
	return t.ParentGroup
}

func (t *Telegram) getMinScore() int {
	return t.MinScore
}

func (t *Telegram) getName() string {
	return t.Name
}
