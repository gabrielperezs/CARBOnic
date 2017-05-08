package main

import "log"

type Telegram struct {
	Token string
	Name  string

	Group int64

	Bot *TelegramBOT

	MinScore int
	chSender chan *Message

	ParentGroup *Group
}

const (
	telegramMaxMessages = 100
)

func (t *Telegram) start() {
	t.chSender = make(chan *Message, telegramMaxMessages)
	go t.sender()
}

func (t *Telegram) sender() {

	for {
		message := <-t.chSender

		if t.Bot == nil {
			log.Println("Telegram was not correctly initialiced", t.Group, t.Name)
			getTelegram(t.ParentGroup)
			return
		}

		t.Bot.Send(t.Group, message)
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
