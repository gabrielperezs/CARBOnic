package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	cmdCATCH = "catch"
	cmdPING  = "ping"
)

var (
	trimSpace = regexp.MustCompile(`/ +/`)
)

type Message struct {
	origin interface{}
	score  int
	msg    string
}

type Plugin interface {
	getName() string
	getParentGroup() *Group
	getMinScore() int
}

func commands(plugin Plugin, from, msg string) {
	if !strings.HasPrefix(msg, "/") {
		return
	}

	g := plugin.getParentGroup()

	msg = trimSpace.ReplaceAllString(strings.ToLower(msg[1:]), " ")
	log.Printf("[%s] - %s - %s - Command: %s", g.Name, plugin.getName(), msg)
	parts := strings.Split(msg, " ")

	if parts[0] == cmdCATCH {
		for _, sqs := range g.SQS {
			if int(g.getMsgScore()) >= sqs.Score || (len(parts) > 1 && parts[1] == "all") || plugin.getMinScore() >= sqs.Score {
				log.Printf("[%s] - %d - Purge SQS %s", g.Name, sqs.Score, sqs.Url)
				sqs.purge()
			}
		}
		g.chReciv <- &Message{
			score: int(g.getMsgScore()),
			msg:   fmt.Sprintf("[%s] catched the alarms", from),
		}
		return
	}

	if parts[0] == cmdPING {
		g.chReciv <- &Message{
			score: 10,
			msg:   "Ping (yes, i'm alive)",
		}
		return
	}
}
