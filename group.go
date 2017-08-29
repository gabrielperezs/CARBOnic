package main

import (
	"log"
	"sync"
)

const (
	groupMaxMessages = 100
)

type Group struct {
	Name     string
	Telegram *Telegram
	HipChat  *HipChat
	SQS      []*SQS
	chReciv  chan *Message
}

func (g *Group) start() {

	g.chReciv = make(chan *Message, groupMaxMessages)

	getTelegram(g)
	getHipChat(g)

	for _, sqs := range g.SQS {
		sqs.startSession()
	}

	go g.groupMessages()
	go g.pullSQS()

	log.Println(g.Name, "stated")

}

func (g *Group) pullSQS() {
	for {
		if len(g.SQS) > 0 {
			wg := sync.WaitGroup{}
			for _, sqs := range g.SQS {
				wg.Add(1)
				go func(sqs *SQS) {
					defer wg.Done()
					sqs.pullSQS(g.chReciv)
				}(sqs)
			}
			wg.Wait()
			continue
		}

		log.Printf("ERROR: No SQS resources in the group %s", g.Name)
		return
	}
}

func (g *Group) groupMessages() {

	for {

		message := <-g.chReciv

		log.Printf("[%s] - %d - %s", g.Name, message.score, message.msg)

		if g.Telegram != nil && message.score >= g.Telegram.MinScore {
			select {
			case g.Telegram.chSender <- message:
			default:
				log.Printf("ERROR: telegram %d channel full", g.Telegram.Group)
			}
		}

		if g.HipChat != nil && message.score >= g.HipChat.MinScore {
			select {
			case g.HipChat.chSender <- message:
			default:
				log.Printf("ERROR: HipChat %s channel full", g.HipChat.RoomID)
			}
		}

	}

}
