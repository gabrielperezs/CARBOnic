package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

func newHipChatPull(roomID, token string) *HipChatPull {
	hb := &HipChatPull{
		roomID:     roomID,
		token:      token,
		maxResults: 1,
		client:     hipchat.NewClient(token),
	}

	go hb.receiver()

	return hb
}

type HipChatPull struct {
	roomID    string
	token     string
	client    *hipchat.Client
	lastMsgID string
	related   []*HipChat

	maxResults int
}

func (hb *HipChatPull) sender(message *Message) {

	if !ignoreDup(fmt.Sprintf("hipchat:%s:%s", hb.roomID, message.msg)) {
		log.Printf("Ignore repeated message on %s %s", "hipchat", hb.roomID)
		return
	}

	if message.score > 5 {
		notifRq := &hipchat.NotificationRequest{
			Color:   "red",
			Message: fmt.Sprintf("@all %s", message.msg),
		}
		_, err := hb.client.Room.Notification(hb.roomID, notifRq)
		if err != nil {
			log.Printf("ERROR [%s]: %s", hb.roomID, err)
		}
	} else {
		msgReq := &hipchat.RoomMessageRequest{
			Message: fmt.Sprintf("%s", message.msg),
		}
		_, err := hb.client.Room.Message(hb.roomID, msgReq)
		if err != nil {
			log.Printf("ERROR [%s]: %s", hb.roomID, err)
		}
	}
}

func (hb *HipChatPull) receiver() {

	hb.maxResults = 1

	for {

		hist, resp, err := hb.client.Room.Latest(hb.roomID, &hipchat.LatestHistoryOptions{
			MaxResults: hb.maxResults,
			NotBefore:  hb.lastMsgID,
		})

		if err != nil {
			log.Printf("ERROR [%s]: %s", hb.roomID, err)
			log.Printf("Server returns %+v (re-try in 15 seconds)\n", resp)

			// Don't try again in the next 15 seconds
			<-time.After(time.Second * 15)
			continue
		}

		for _, m := range hist.Items {

			if m.ID == hb.lastMsgID {
				continue
			}

			from := ""
			switch m.From.(type) {
			case string:
				from = m.From.(string)
			case map[string]interface{}:
				f := m.From.(map[string]interface{})
				from = f["name"].(string)
			}

			msg := m.Message
			hb.lastMsgID = m.ID

			if hb.maxResults == 1 {
				continue
			}

			for _, h := range hb.related {
				commands(h, from, msg)
			}
		}

		if hb.maxResults == 1 {
			hb.maxResults = 10
		}

		<-time.After(time.Duration(hipChatInterval+rand.Int31n(4)) * time.Second)
	}

}
