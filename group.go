package main

import (
	"log"
	"sync/atomic"
	"time"
)

type Group struct {
	Name         string
	Telegram     *Telegram
	HipChat      *HipChat
	SQS          []*SQS
	chReciv      chan *Message
	lastMsgScore int64
}

func (g *Group) start() {

	g.chReciv = make(chan *Message)

	getTelegram(g)
	getHipChat(g)

	go g.pullSQS()
	go g.groupMessages()

}

func (g *Group) setMsgScore(score int) {
	atomic.StoreInt64(&g.lastMsgScore, int64(score))
}

func (g *Group) getMsgScore() int64 {
	return atomic.LoadInt64(&g.lastMsgScore)
}

func (g *Group) purgeSQS() {
	if len(g.SQS) > 0 {
		for _, sqs := range g.SQS {
			sqs.purge()
		}
	}
}

func (g *Group) pullSQS() {
	for {
		if len(g.SQS) > 0 {
			for _, sqs := range g.SQS {
				sqs.pullSQS(g.chReciv)
			}
		}

		<-time.After(15 * 1000 * time.Millisecond)
	}
}

func (g *Group) groupMessages() {

	for {

		message := <-g.chReciv

		g.setMsgScore(message.score)

		log.Printf("[%s] - %d - %s", g.Name, message.score, message.msg)

		if g.Telegram != nil && message.score >= g.Telegram.MinScore {
			select {
			case g.Telegram.chSender <- message:
			default:
			}
		}

		if g.HipChat != nil && message.score >= g.HipChat.MinScore {
			select {
			case g.HipChat.chSender <- message:
			default:
			}
		}

	}

}
