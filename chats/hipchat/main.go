package hipchat

import (
	"fmt"
	"log"
	"sync"

	"github.com/gabrielperezs/CARBOnic/lib"
)

var connPool sync.Map

type Config struct {
	Name     string
	Token    string
	RoomID   string
	MinScore int
}

func NewOrGet(c map[string]interface{}) (*HipChat, error) {

	if _, ok := c["Token"]; !ok {
		return nil, fmt.Errorf("HIPCHAT ERROR: Token not found or invalid")
	}

	if _, ok := c["MinScore"]; !ok {
		return nil, fmt.Errorf("HIPCHAT ERROR: MinScore not defined")
	}

	if _, ok := c["RoomID"]; !ok {
		return nil, fmt.Errorf("HIPCHAT ERROR: RoomID not defined")
	}

	cfg := &Config{
		Token:    c["Token"].(string),
		MinScore: int(c["MinScore"].(int64)),
		RoomID:   c["RoomID"].(string),
	}

	var ok bool

	var havConn interface{}

	if havConn, ok = connPool.Load(cfg.Token); !ok {
		havConn = newConnection(cfg.Token)
	}

	connPool.Store(cfg.Token, havConn)

	t := &HipChat{
		cfg:        &Config{},
		chIn:       make(chan *lib.Message, 100),
		conn:       havConn.(*HipChatClient),
		group:      nil,
		maxResults: 1,
	}
	// Copy configuration
	*t.cfg = *cfg

	t.conn.Lock()
	t.conn.rooms = append(t.conn.rooms, t)
	t.conn.Unlock()

	go t.listener()

	return t, nil
}

type HipChat struct {
	sync.Mutex

	lastMsgID  string
	maxResults int

	cfg     *Config
	chIn    chan *lib.Message
	conn    *HipChatClient
	group   lib.Group
	exiting bool
}

func (t *HipChat) listener() {
	for m := range t.chIn {
		t.conn.sender(t.cfg.RoomID, m)
	}
}

func (t *HipChat) GetLabel() string {
	return fmt.Sprintf("HipChat %s (%d)", t.cfg.RoomID, t.cfg.MinScore)
}

func (t *HipChat) MinScore() int {
	return t.cfg.MinScore
}

func (t *HipChat) SetGroup(g lib.Group) {
	t.group = g
}

func (t *HipChat) Group() lib.Group {
	return t.group
}

func (t *HipChat) Chan() chan *lib.Message {
	return t.chIn
}

func (t *HipChat) Exit() {

	t.Lock()
	e := t.exiting
	t.Unlock()

	if e {
		return
	}

	defer log.Printf("HIPCHAT closed %s", t.cfg.RoomID)

	t.Lock()
	t.exiting = true
	t.Unlock()

	t.conn.Lock()
	defer t.conn.Unlock()

	var n []*HipChat
	for _, i := range t.conn.rooms {
		if i.cfg.RoomID != t.cfg.RoomID {
			n = append(n, i)
		}
	}

	t.conn.rooms = n

	if len(t.conn.rooms) == 0 {
		close(t.chIn)
		return
	}

}
