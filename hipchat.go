package main

type HipChat struct {
	Token  string
	Name   string
	RoomID string

	lastMsgID  string
	maxResults int

	Pull *HipChatPull

	MinScore int
	chSender chan *Message

	ParentGroup *Group
}

const (
	hipChatMaxMessages = 100
)

func (h *HipChat) start() {
	h.chSender = make(chan *Message, hipChatMaxMessages)

	go h.sender()
}

func (h *HipChat) sender() {
	for {
		h.Pull.sender(<-h.chSender)
	}
}

func (h *HipChat) getParentGroup() *Group {
	return h.ParentGroup
}

func (h *HipChat) getMinScore() int {
	return h.MinScore
}

func (h *HipChat) getName() string {
	return h.Name
}
