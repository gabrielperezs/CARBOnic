package cmds

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/gabrielperezs/CARBOnic/lib"
)

const (
	cmdCATCH = "catch"
	cmdPURGE = "purge"
	cmdPING  = "ping"
)

var (
	trimSpace = regexp.MustCompile(`/ +/`)
)

func Commands(plugin lib.Plugin, from, Msg string) {
	if !strings.HasPrefix(Msg, "/") {
		return
	}

	g := plugin.Group()

	Msg = trimSpace.ReplaceAllString(strings.ToLower(Msg[1:]), " ")

	parts := strings.Split(Msg, " ")

	maxLevelAlarms := plugin.MinScore()

	switch parts[0] {
	case cmdCATCH:

		for _, v := range g.GetInputs() {
			input := v.(lib.Input)
			fmt.Println(input.GetLabel())
			if input.HasAlarms() {
				if input.GetScore() > maxLevelAlarms {
					maxLevelAlarms = input.GetScore()
				}
				log.Printf("[%s] - %d - Clean SQS %s", g.GetName(), input.GetScore(), input.GetLabel())
				go input.Clean()
			}
		}

		g.Chan() <- &lib.Message{
			Score: maxLevelAlarms,
			Msg:   fmt.Sprintf("[%s] caught last alarms", from),
		}
		return

	case cmdPURGE:

		for _, v := range g.GetInputs() {
			input := v.(lib.Input)
			if input.HasAlarms() {
				if input.GetScore() > maxLevelAlarms {
					maxLevelAlarms = input.GetScore()
				}
				log.Printf("[%s] - %d - PURGE SQS %s", g.GetName(), input.GetScore(), input.GetLabel())
				go input.Purge()
			}
		}

		g.Chan() <- &lib.Message{
			Score: maxLevelAlarms,
			Msg:   fmt.Sprintf("[%s] purged all the alarms", from),
		}
		return

	case cmdPING:

		g.Chan() <- &lib.Message{
			Score: 10,
			Msg:   "Ping (yes, i'm alive)",
		}
		return
	}

}
