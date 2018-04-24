package telegram

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gabrielperezs/CARBOnic/lib"
)

var connPool sync.Map

type Config struct {
	Name     string
	Token    string
	Group    int64
	MinScore int
}

func NewOrGet(c map[string]interface{}) (*Telegram, error) {

	cfg := &Config{}

	for k, v := range c {
		switch strings.ToLower(k) {
		case "token":
			cfg.Token = v.(string)
		case "minscore":
			cfg.MinScore = int(v.(int64))
		case "group":
			cfg.Group = v.(int64)
		}
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("TELEGRAM ERROR: Token not found or invalid")
	}

	if cfg.Group == 0 {
		return nil, fmt.Errorf("TELEGRAM ERROR: Group not defined")
	}

	var ok bool

	var havConn interface{}

	if havConn, ok = connPool.Load(cfg.Token); !ok {
		var err error
		havConn, err = NewConnection(cfg.Token)
		if err != nil {
			return nil, err
		}
	}

	connPool.Store(cfg.Token, havConn)

	conn := havConn.(*TelegramBOT)

	t := &Telegram{
		cfg:   &Config{},
		ch:    make(chan *lib.Message, 100),
		conn:  conn,
		group: nil,
	}
	// Copy configuration
	*t.cfg = *cfg

	t.conn.r = append(t.conn.r, t)

	go t.listener()

	return t, nil
}

type Telegram struct {
	sync.Mutex

	cfg     *Config
	ch      chan *lib.Message
	conn    *TelegramBOT
	group   lib.Group
	exiting bool
}

func (t *Telegram) listener() {
	for m := range t.ch {
		go t.conn.Send(t.cfg.Group, m)
	}
}

func (t *Telegram) GetLabel() string {
	return fmt.Sprintf("Telegram %d (%d)", t.cfg.Group, t.cfg.MinScore)
}

func (t *Telegram) MinScore() int {
	return t.cfg.MinScore
}

func (t *Telegram) SetGroup(g lib.Group) {
	t.group = g
}

func (t *Telegram) Group() lib.Group {
	return t.group
}

func (t *Telegram) Chan() chan *lib.Message {
	return t.ch
}

func (t *Telegram) Exit() {

	t.Lock()
	e := t.exiting
	t.Unlock()

	if e == true {
		return
	}

	defer log.Printf("TELEGRAM closed %d", t.cfg.Group)

	t.Lock()
	t.exiting = true
	close(t.ch)
	t.ch = nil
	t.Unlock()

	t.conn.Lock()
	defer t.conn.Unlock()

	var n []*Telegram
	for _, i := range t.conn.r {
		if i != t && t != nil {
			n = append(n, i)
		}
	}
	t.conn.r = n

}
