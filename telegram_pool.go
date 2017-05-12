package main

import (
	"log"
	"sync"
	"time"
)

var (
	mTelegram       = &sync.Mutex{}
	telegramClients = make(map[string]*TelegramBOT)

	limitRepeatMsg time.Duration = time.Second * 10
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
	log.Printf("Group %s on Telegram %d", g.Name, t.Group)

	if _, ok := telegramClients[t.Token]; ok {
		t.Bot = telegramClients[t.Token]
		t.Bot.related = append(t.Bot.related, t)
	} else {
		t.Bot = newTelegramBOT(t.Token)
		t.Bot.related = append(t.Bot.related, t)
	}

	t.ParentGroup = g

	telegramClients[t.Token] = t.Bot

	t.start()

	return
}
