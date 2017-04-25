package main

import (
	"log"
	"sync"

	"gopkg.in/telegram-bot-api.v4"
)

var (
	mTelegram   = &sync.Mutex{}
	CwTelegrams = make(map[string]*Telegram)
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
	t.ParentGroup = g

	if _, ok := CwTelegrams[t.Token]; ok {
		return
	}

	var err error
	t.Bot, err = tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Println("Telegram connection error:", t.Name, err)
		t.Bot = nil
		return
	}

	t.ParentGroup = g

	CwTelegrams[t.Token] = t

	t.start()

	return
}
