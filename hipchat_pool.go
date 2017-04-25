package main

import (
	"log"
	"sync"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

const (
	hipChatInterval = 5
)

var (
	mHipChat   = &sync.Mutex{}
	CwHipChats = make(map[string]*HipChat)
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
	h.ParentGroup = g

	if _, ok := CwHipChats[h.Token]; ok {
		return
	}

	if h.Token == "" {
		return
	}

	if h.RoomID == "" {
		return
	}

	h.Client = hipchat.NewClient(h.Token)

	CwHipChats[h.Token] = h

	h.start()

	return
}
