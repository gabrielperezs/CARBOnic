package slack

import (
	"fmt"
	"log"
	"sync"

	"github.com/nlopes/slack"

	"github.com/gabrielperezs/CARBOnic/cmds"
	"github.com/gabrielperezs/CARBOnic/lib"
)

func NewConnection(token string) (*SlackBOT, error) {

	b := &SlackBOT{}

	if err := b.connect(token); err != nil {
		return nil, err
	}

	b.rtm = b.bot.NewRTM()
	go b.rtm.ManageConnection()
	go b.listener()

	return b, nil
}

type SlackBOT struct {
	sync.Mutex
	me  *slack.AuthTestResponse
	rtm *slack.RTM
	bot *slack.Client
	r   []*Slack
}

func (tb *SlackBOT) connect(token string) error {

	var err error

	tb.bot = slack.New(token)

	tb.me, err = tb.bot.AuthTest()
	if err != nil {
		log.Printf("ERROR: Slack AuthTest: %s", err)
		return err
	}

	log.Printf("Slack Authorized on account %s", tb.me.User)

	return nil
}

func (tb *SlackBOT) listener() {

	for msg := range tb.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:

			if ev.Msg.User == "" {
				continue
			}

			if ev.User == tb.me.User {
				log.Printf("D: ignore my self")
				continue
			}

			group := false
			for _, t := range tb.r {
				if t.cfg.Group == ev.Msg.Channel {
					group = true
					if lib.IsDupMessage(fmt.Sprintf("slack_cmd_%s", ev.Channel), ev.Text) {
						continue
					}
					info, err := tb.rtm.GetUserInfo(ev.User)
					if err != nil {
						cmds.Commands(t, "unknown", ev.Text)
					}
					cmds.Commands(t, info.Profile.DisplayName, ev.Text)
				}
			}

			if !group {
				go tb.rtm.PostMessage(ev.Msg.Channel, slack.MsgOptionText("Can you read?", false))
			}

		case *slack.RTMError:
			log.Printf("Slack ERROR: %s", ev.Error())
		}
	}

}

func (tb *SlackBOT) Send(g string, m *lib.Message) error {
	_, _, err := tb.rtm.PostMessage(
		g,
		slack.MsgOptionText(m.Msg, false),
	)
	if err != nil {
		return err
	}
	return nil
}
