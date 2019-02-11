package chats

import (
	"errors"
	"fmt"

	"github.com/gabrielperezs/CARBOnic/chats/hipchat"
	"github.com/gabrielperezs/CARBOnic/chats/slack"
	"github.com/gabrielperezs/CARBOnic/chats/telegram"
	"github.com/gabrielperezs/CARBOnic/lib"
)

func Get(cfg interface{}) (lib.Chat, error) {

	c := cfg.(map[string]interface{})
	if _, ok := c["Type"]; !ok {
		return nil, errors.New("Type not defined")
	}

	switch c["Type"] {
	case "telegram":
		return telegram.NewOrGet(c)
	case "hipchat":
		return hipchat.NewOrGet(c)
	case "slack":
		return slack.NewOrGet(c)
	default:
		return nil, fmt.Errorf("Plugin don't exists: %s", c["Type"])
	}
}
