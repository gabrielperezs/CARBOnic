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
		cfg:   &Config{},
		chIn:  make(chan *lib.Message, 100),
		conn:  havConn.(*HipChatClient),
		group: nil,
	}
	// Copy configuration
	*t.cfg = *cfg

	if iR, ok := t.conn.rooms.Load(t.cfg.RoomID); ok {
		r := iR.([]*HipChat)
		r = append(r, t)
		t.conn.rooms.Store(t.cfg.RoomID, r)
	} else {
		r := []*HipChat{t}
		t.conn.rooms.Store(t.cfg.RoomID, r)
	}

	go t.listener()

	return t, nil
}

type HipChat struct {
	sync.Mutex

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

	if iR, ok := t.conn.rooms.Load(t.cfg.RoomID); ok {
		var n []*HipChat
		for _, i := range iR.([]*HipChat) {
			if i.cfg.RoomID != t.cfg.RoomID {
				n = append(n, i)
			}
		}

		if len(n) == 0 {
			t.conn.rooms.Delete(t.cfg.RoomID)
			close(t.chIn)
			return
		}

		t.conn.rooms.Store(t.cfg.RoomID, n)
	}
}
