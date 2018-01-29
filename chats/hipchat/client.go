package hipchat

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/gabrielperezs/CARBOnic/cmds"
	"github.com/gabrielperezs/CARBOnic/lib"
	"github.com/tbruyelle/hipchat-go/hipchat"
)

const (
	hipchatRetry          = 60 // 60 seconds
	hipChatInterval int32 = 15 // Will be modify with random 0 to 5
)

func newConnection(token string) *HipChatClient {

	hcc := &HipChatClient{
		token:      token,
		maxResults: 1,
		client:     hipchat.NewClient(token),
	}

	go hcc.receiver()

	return hcc
}

type HipChatClient struct {
	sync.Mutex

	token     string
	client    *hipchat.Client
	lastMsgID string
	rooms     sync.Map

	maxResults int
}

func (hb *HipChatClient) sender(roomID string, message *lib.Message) {

	// If the same message was sent few seconds ago just discard it
	if lib.IsDupMessage(fmt.Sprintf("hipchat_%s", roomID), message) {
		return
	}

	var err error

	if message.Score > 5 {
		notifRq := &hipchat.NotificationRequest{
			Color:   "red",
			Message: fmt.Sprintf("@all %s", message.Msg),
		}
		_, err = hb.client.Room.Notification(roomID, notifRq)
	} else {
		msgReq := &hipchat.RoomMessageRequest{
			Message: fmt.Sprintf("%s", message.Msg),
		}
		_, err = hb.client.Room.Message(roomID, msgReq)
	}

	if err != nil {
		log.Printf("HipChat Pull ERROR [%s]: %s", roomID, err)
	}

}

func (hb *HipChatClient) receiver() {

	hb.maxResults = 1

	for {

		hb.rooms.Range(func(key, value interface{}) bool {

			roomID := key.(string)

			hist, resp, err := hb.client.Room.Latest(roomID, &hipchat.LatestHistoryOptions{
				MaxResults: hb.maxResults,
				NotBefore:  hb.lastMsgID,
			})

			if err != nil {
				log.Printf("HipChat Pull ERROR [RoomID %s]: %s", roomID, err)
				log.Printf("HipChat Pull ERROR [RoomID %s] response: %#v", roomID, resp)
				log.Printf("HipChat Pull [RoomID %s] re-try in %d seconds", roomID, hipchatRetry)

				time.Sleep(time.Second * hipchatRetry)
				return true
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

				for _, g := range value.([]*HipChat) {
					cmds.Commands(g, from, msg)
				}
			}

			if hb.maxResults == 1 {
				hb.maxResults = 10
			}

			time.Sleep(1 * time.Second)
			return true
		})

		time.Sleep(time.Duration(hipChatInterval+rand.Int31n(5)) * time.Second)
	}

}
