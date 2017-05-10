package main

import (
	"fmt"
	"log"
	"sync"
)

var (
	mHipChat       = &sync.Mutex{}
	hipchatClients = make(map[string]*HipChatPull)
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

	key := fmt.Sprintf("%s:%s", h.RoomID, h.Token)
	log.Printf("Group %s on HipChat %s", g.Name, h.RoomID)

	if _, ok := hipchatClients[key]; ok {
		h.Pull = hipchatClients[key]
	} else {
		h.Pull = newHipChatPull(h.RoomID, h.Token)
	}

	h.Pull.related = append(h.Pull.related, h)

	h.ParentGroup = g

	hipchatClients[key] = h.Pull

	h.start()

	return
}
