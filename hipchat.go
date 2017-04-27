package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tbruyelle/hipchat-go/hipchat"
)

type HipChat struct {
	Token  string
	Name   string
	RoomID string

	lastMsgID  string
	maxResults int

	Client *hipchat.Client

	MinScore int
	chSender chan *Message

	ParentGroup *Group
}

const (
	hipChatInterval    = 15
	hipChatMaxMessages = 100
)

func (h *HipChat) start() {
	h.chSender = make(chan *Message, hipChatMaxMessages)

	go h.receiver()
	go h.sender()
}

func (h *HipChat) sender() {
	for {
		message := <-h.chSender

		if message.score > 5 {
			notifRq := &hipchat.NotificationRequest{
				Color:   "red",
				Message: fmt.Sprintf("@all %s", message.msg),
			}
			_, err := h.Client.Room.Notification(h.RoomID, notifRq)
			if err != nil {
				log.Println("ERROR [%s]: %s", h.Name, err)
			}
		} else {
			msgReq := &hipchat.RoomMessageRequest{
				Message: fmt.Sprintf("%s", message.msg),
			}
			_, err := h.Client.Room.Message(h.RoomID, msgReq)
			if err != nil {
				log.Println("ERROR [%s]: %s", h.Name, err)
			}
		}
	}
}

func (h *HipChat) receiver() {

	h.maxResults = 1

	for {

		hist, resp, err := h.Client.Room.Latest(h.RoomID, &hipchat.LatestHistoryOptions{
			MaxResults: h.maxResults,
			NotBefore:  h.lastMsgID,
		})

		if err != nil {
			log.Printf("Error during room history req %q\n", err)
			log.Printf("Server returns %+v\n", resp)
			continue
		}

		for _, m := range hist.Items {

			if m.ID == h.lastMsgID {
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
			h.lastMsgID = m.ID

			if h.maxResults == 1 {
				continue
			}

			commands(h, from, msg)
		}

		if h.maxResults == 1 {
			h.maxResults = 10
		}

		<-time.After(hipChatInterval * time.Second)
	}

}

func (h *HipChat) getParentGroup() *Group {
	return h.ParentGroup
}

func (h *HipChat) getMinScore() int {
	return h.MinScore
}

func (h *HipChat) getName() string {
	return h.Name
}
