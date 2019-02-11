package slack

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
	Group    string
	MinScore int
}

func NewOrGet(c map[string]interface{}) (*Slack, error) {

	cfg := &Config{}

	for k, v := range c {
		switch strings.ToLower(k) {
		case "token":
			cfg.Token = v.(string)
		case "minscore":
			cfg.MinScore = int(v.(int64))
		case "group":
			cfg.Group = v.(string)
		}
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("Slack ERROR: Token not found or invalid")
	}

	if cfg.Group == "" {
		return nil, fmt.Errorf("Slack ERROR: Group not defined")
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

	conn := havConn.(*SlackBOT)

	t := &Slack{
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

type Slack struct {
	sync.Mutex

	cfg     *Config
	ch      chan *lib.Message
	conn    *SlackBOT
	group   lib.Group
	exiting bool
}

func (t *Slack) listener() {
	for m := range t.ch {
		go t.conn.Send(t.cfg.Group, m)
	}
}

func (t *Slack) GetLabel() string {
	return fmt.Sprintf("Slack %s (%d)", t.cfg.Group, t.cfg.MinScore)
}

func (t *Slack) MinScore() int {
	return t.cfg.MinScore
}

func (t *Slack) SetGroup(g lib.Group) {
	t.group = g
}

func (t *Slack) Group() lib.Group {
	return t.group
}

func (t *Slack) Chan() chan *lib.Message {
	return t.ch
}

func (t *Slack) Exit() {

	t.Lock()
	e := t.exiting
	t.Unlock()

	if e == true {
		return
	}

	defer log.Printf("Slack closed %s", t.cfg.Group)

	t.Lock()
	t.exiting = true
	close(t.ch)
	t.ch = nil
	t.Unlock()

	t.conn.Lock()
	defer t.conn.Unlock()

	var n []*Slack
	for _, i := range t.conn.r {
		if i != t && t != nil {
			n = append(n, i)
		}
	}
	t.conn.r = n
}
