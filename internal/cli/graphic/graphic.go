package graphic

type Glyphs struct {
	Success  string
	Error    string
	Info     string
	Agent    string
	Artifact string
}

func DefaultGlyphs() Glyphs {
	return Glyphs{
		Success:  "✓",
		Error:    "✗",
		Info:     "•",
		Agent:    "◆",
		Artifact: "⬢",
	}
}

func (g Glyphs) Status(ok bool) string {
	if ok {
		return g.Success
	}
	return g.Error
}
