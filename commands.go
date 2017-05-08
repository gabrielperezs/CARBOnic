package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

const (
	cmdCATCH = "catch"
	cmdPURGE = "purge"
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

	log.Printf("[%s] - Command: %s - %s", g.Name, plugin.getName(), msg)
	parts := strings.Split(msg, " ")

	maxLevelAlarms := plugin.getMinScore()

	switch parts[0] {
	case cmdCATCH:

		for _, sqs := range g.SQS {
			fmt.Println(sqs.Url)
			if sqs.hasAlarms() {
				if sqs.Score > maxLevelAlarms {
					maxLevelAlarms = sqs.Score
				}
				log.Printf("[%s] - %d - Clean SQS %s", g.Name, sqs.Score, sqs.Url)
				go sqs.clean()
			}
		}

		g.chReciv <- &Message{
			score: maxLevelAlarms,
			msg:   fmt.Sprintf("[%s] caught last alarms", from),
		}
		return

	case cmdPURGE:

		for _, sqs := range g.SQS {
			if sqs.hasAlarms() {
				if sqs.Score > maxLevelAlarms {
					maxLevelAlarms = sqs.Score
				}
				log.Printf("[%s] - %d - PURGE SQS %s", g.Name, sqs.Score, sqs.Url)
				go sqs.purge()
			}
		}

		g.chReciv <- &Message{
			score: maxLevelAlarms,
			msg:   fmt.Sprintf("[%s] purged all the alarms", from),
		}
		return

	case cmdPING:

		g.chReciv <- &Message{
			score: 10,
			msg:   "Ping (yes, i'm alive)",
		}
		return
	}

}
