package main

import (
	"log"
	"sync"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

var (
	mHipChat       = &sync.Mutex{}
	hipchatClients = make(map[string]*hipchat.Client)
)

func getHipChat(g *Group) {

	if g.HipChat == nil {
		log.Printf("[%s] HipChat disabled", g.Name)
		return
	}

	if g.HipChat.Token == "" {
		log.Printf("[%s] HipChat disabled, empty token", g.Name)
		return
	}

	mHipChat.Lock()
	defer mHipChat.Unlock()

	h := g.HipChat

	if _, ok := hipchatClients[h.Token]; ok {
		g.HipChat.Client = hipchatClients[h.Token]
	} else {
		h.Client = hipchat.NewClient(h.Token)
	}

	h.ParentGroup = g

	hipchatClients[h.Token] = h.Client

	h.start()

	return
}
