package keybindings

import "charm.land/bubbles/v2/key"

type KeyMap struct {
	Help     key.Binding
	Exit     key.Binding
	Clear    key.Binding
	Settings key.Binding
	Diff     key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Help:     key.NewBinding(key.WithKeys("f1", "?"), key.WithHelp("f1/?", "help")),
		Exit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
		Clear:    key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "clear")),
		Settings: key.NewBinding(key.WithKeys("ctrl+,"), key.WithHelp("ctrl+,", "settings")),
		Diff:     key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "diff")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Exit, k.Clear}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Help, k.Exit, k.Clear}, {k.Settings, k.Diff}}
}
