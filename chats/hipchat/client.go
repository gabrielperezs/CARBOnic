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
		token:  token,
		client: hipchat.NewClient(token),
	}

	go hcc.receiver()

	return hcc
}

type HipChatClient struct {
	sync.Mutex

	token  string
	client *hipchat.Client
	rooms  []*HipChat
}

func (hb *HipChatClient) sender(roomID string, message *lib.Message) {

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

	for {

		hb.Lock()
		rooms := hb.rooms
		hb.Unlock()

		for _, t := range rooms {

			// The room is exiting
			if t.exiting {
				continue
			}

			roomID := t.cfg.RoomID

			hist, resp, err := hb.client.Room.Latest(roomID, &hipchat.LatestHistoryOptions{
				MaxResults: t.maxResults,
				NotBefore:  t.lastMsgID,
			})

			if err != nil {
				log.Printf("HipChat Pull ERROR [RoomID %s]: %s", roomID, err)
				log.Printf("HipChat Pull ERROR [RoomID %s] response: %#v", roomID, resp)
				log.Printf("HipChat Pull [RoomID %s] re-try in %d seconds", roomID, hipchatRetry)

				time.Sleep(time.Second * hipchatRetry)
				continue
			}

			for _, m := range hist.Items {

				if m.ID == t.lastMsgID {
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
				t.lastMsgID = m.ID

				if t.maxResults == 1 {
					continue
				}

				if lib.IsDupMessage(fmt.Sprintf("hipchat_cmd_%s", roomID), msg) {
					continue
				}

				cmds.Commands(t, from, msg)
			}

			if t.maxResults == 1 {
				t.maxResults = 10
			}

			time.Sleep(1 * time.Second)
		}

		time.Sleep(time.Duration(hipChatInterval+rand.Int31n(5)) * time.Second)
	}

}
