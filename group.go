package main

import (
	"log"
	"sync"

	"github.com/gabrielperezs/CARBOnic/chats"
	"github.com/gabrielperezs/CARBOnic/inputs"
	"github.com/gabrielperezs/CARBOnic/lib"
)

const (
	groupMaxMessages = 100
)

type Group struct {
	sync.Mutex
	sync.WaitGroup

	Name  string
	Chat  []interface{}
	Input []interface{}

	chats  []lib.Chat
	inputs []lib.Input

	Ch chan *lib.Message

	done    chan bool
	exiting bool
}

func (g *Group) GetName() string {
	return g.Name
}

func (g *Group) GetChats() []lib.Chat {
	return g.chats
}

func (g *Group) GetInputs() []lib.Input {
	return g.inputs
}

func (g *Group) Chan() chan *lib.Message {
	return g.Ch
}

func (g *Group) start() {

	g.Ch = make(chan *lib.Message, groupMaxMessages)

	for _, cfg := range g.Chat {
		c, err := chats.Get(cfg)
		if err != nil {
			log.Printf("ERROR: %s", err)
			continue
		}

		c.SetGroup(g)

		g.Lock()
		g.chats = append(g.chats, c)
		g.Unlock()
	}

	for _, cfg := range g.Input {
		i, err := inputs.Get(cfg)
		if err != nil {
			log.Printf("ERROR: %s", err)
			continue
		}

		i.SetGroup(g)

		g.Lock()
		g.inputs = append(g.inputs, i)
		g.Unlock()
	}

	go g.listen()

	log.Printf("Group start: %s (inputs: %d, chats: %d)", g.Name, len(g.inputs), len(g.chats))

}

func (g *Group) listen() {
	for message := range g.Ch {
		log.Printf("[%s] - %d - %s", g.Name, message.Score, message.Msg)
		for _, chat := range g.chats {
			if message.Score >= chat.MinScore() {

				g.Lock()
				e := g.exiting
				g.Unlock()
				if e {
					break
				}

				select {
				case chat.Chan() <- message:
				default:
					log.Printf("ERROR [%s]: chat channel full", g.Name)
				}
			}
		}
	}
}

func (g *Group) Exit() {

	g.Lock()
	g.exiting = true
	g.Unlock()

	var wg sync.WaitGroup
	for _, c := range g.chats {
		wg.Add(1)
		go func(c lib.Chat) {
			defer wg.Done()
			c.Exit()
		}(c)
	}
	wg.Wait()

	for _, c := range g.inputs {
		wg.Add(1)
		go func(c lib.Input) {
			defer wg.Done()
			c.DelGroup(g)
		}(c)
	}
	wg.Wait()

	if g.Ch != nil {
		close(g.Ch)
		g.Ch = nil
	}
}
