package main

import (
	"log"
	"time"
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
			for _, sqs := range g.SQS {
				sqs.pullSQS(g.chReciv)
			}
		}

		<-time.After(8 * 1000 * time.Millisecond)
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
				log.Println("ERROR: telegram channel full")
			}
		}

		if g.HipChat != nil && message.score >= g.HipChat.MinScore {
			select {
			case g.HipChat.chSender <- message:
			default:
				log.Println("ERROR: HipChat channel full")
			}
		}

	}

}
