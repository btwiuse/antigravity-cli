package messages

type Ready struct{}

type Info struct {
	Text string
}

type Warning struct {
	Text string
}

type Error struct {
	Err error
}

type CommandSelected struct {
	Name string
}

type ConversationStarted struct {
	ID string
}

type StepUpdated struct {
	Title string
	Done  bool
}

func (m Error) Error() string {
	if m.Err == nil {
		return ""
	}
	return m.Err.Error()
}
