package lib

type Plugin interface {
	MinScore() int
	Group() Group
}

type Group interface {
	GetChats() []Chat
	GetInputs() []Input
	GetName() string
	Chan() chan *Message
	Exit()
}

type Chat interface {
	GetLabel() string
	MinScore() int
	SetGroup(g Group)
	Group() Group
	Chan() chan *Message
	Exit()
}

type Input interface {
	StartSession()
	SetGroup(g Group)
	DelGroup(g Group)
	GetScore() int
	HasAlarms() bool
	GetLabel() string
	Clean()
	Purge()
	Exit()
}
