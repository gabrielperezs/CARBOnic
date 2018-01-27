package telegram

import (
	"log"
	"sync"
	"time"

	"github.com/gabrielperezs/CARBOnic/cmds"
	"github.com/gabrielperezs/CARBOnic/lib"
	"gopkg.in/telegram-bot-api.v4"
)

func NewConnection(token string) (*TelegramBOT, error) {

	b := &TelegramBOT{}

	if err := b.connect(token); err != nil {
		return nil, err
	}

	go b.listener()

	return b, nil
}

type TelegramBOT struct {
	sync.Mutex
	bot *tgbotapi.BotAPI
	r   []*Telegram
}

func (tb *TelegramBOT) connect(token string) error {

	var err error
	tb.bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	log.Printf("Telegram Authorized on account %s", tb.bot.Self.UserName)

	return nil
}

func (tb *TelegramBOT) listener() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := tb.bot.GetUpdatesChan(u)
	if err != nil {
		log.Printf("ERROR: Telegram, can't read from channel - %s", err)
		time.Sleep(15 * time.Second)
		return
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		group := false

		for _, t := range tb.r {
			if t.cfg.Group == update.Message.Chat.ID {
				cmds.Commands(t, update.Message.From.String(), update.Message.Text)
				group = true
			}
		}

		if !group {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello, i'm alive. Please use the groups to send commands.")
			tb.bot.Send(msg)
		}
	}
}

func (tb *TelegramBOT) Send(g int64, m *lib.Message) error {
	msg := tgbotapi.NewMessage(g, m.Msg)
	if _, err := tb.bot.Send(msg); err != nil {
		return err
	}
	return nil
}
