package viewport

import "strings"

type Model struct {
	Width  int
	Height int
	offset int
	lines  []string
}

func New(width, height int) Model {
	return Model{Width: width, Height: height}
}

func (m *Model) SetContent(content string) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if normalized == "" {
		m.lines = nil
		m.offset = 0
		return
	}
	m.lines = strings.Split(normalized, "\n")
	limit := len(m.lines) - m.Height
	if limit < 0 {
		limit = 0
	}
	if m.offset > limit {
		m.offset = limit
	}
}

func (m *Model) ScrollDown(lines int) {
	limit := len(m.lines) - m.Height
	if limit < 0 {
		limit = 0
	}
	m.offset += lines
	if m.offset > limit {
		m.offset = limit
	}
}

func (m *Model) ScrollUp(lines int) {
	m.offset -= lines
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) VisibleLines() []string {
	if len(m.lines) == 0 {
		return nil
	}
	if m.Height <= 0 || m.Height >= len(m.lines) {
		return append([]string(nil), m.lines...)
	}
	end := m.offset + m.Height
	if end > len(m.lines) {
		end = len(m.lines)
	}
	return append([]string(nil), m.lines[m.offset:end]...)
}

func (m Model) View() string {
	return strings.Join(m.VisibleLines(), "\n")
}
