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

func (tb *SlackBOT) connect(token string) (err error) {
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
				continue
			}
			for _, t := range tb.r {
				if t.cfg.Group != ev.Msg.Channel {
					// Ignore message that are not comming from the configured channels/groups
					continue
				}
				if lib.IsDupMessage(fmt.Sprintf("slack_cmd_%s", ev.Channel), ev.Text) {
					// Ignore duplicate messages
					continue
				}
				info, err := tb.rtm.GetUserInfo(ev.User)
				if err != nil {
					// If is an unknown user just define as "unknown" an continue
					// because this is consider as an error
					cmds.Commands(t, fmt.Sprintf("[unknown:%s]", err), ev.Text)
					continue
				}
				name := info.Profile.DisplayName
				if name == "" {
					name = info.Profile.RealName
				}
				cmds.Commands(t, name, ev.Text)
			}
		case *slack.RTMError:
			log.Printf("Slack ERROR: %s", ev.Error())
		}
	}
}

func (tb *SlackBOT) Send(g string, message *lib.Message) error {
	if message.Score > 5 {
		_, _, err := tb.rtm.PostMessage(
			g,
			slack.MsgOptionText(message.Msg, false),
			slack.MsgOptionAttachments(
				slack.Attachment{
					Color: "#FB4444",
					Text:  ":bell: <!channel>",
				},
			),
		)
		return err
	}
	_, _, err := tb.rtm.PostMessage(g, slack.MsgOptionText(message.Msg, false))
	return err
}
