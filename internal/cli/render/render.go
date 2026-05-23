package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/anthropics/antigravity-cli/internal/cli/types"
)

type Theme struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Command  lipgloss.Style
	Dim      lipgloss.Style
}

func DefaultTheme() Theme {
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		Subtitle: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		Command:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	}
}

func (t Theme) RenderHeader(title, subtitle string) string {
	if subtitle == "" {
		return t.Title.Render(title)
	}
	return t.Title.Render(title) + "\n" + t.Subtitle.Render(subtitle)
}

func RenderHelp(info types.VersionInfo, cmds []types.SlashCommand) string {
	theme := DefaultTheme()
	lines := []string{
		theme.RenderHeader(info.Name, fmt.Sprintf("%s • %s • %s", info.Codename, info.Version, info.ModulePath)),
		"",
		theme.Dim.Render("Usage:"),
		"  agy [flags] [prompt]",
		"  agy -version",
		"  agy -tui",
		"",
		theme.Dim.Render("Slash commands:"),
	}
	for _, cmd := range cmds {
		lines = append(lines, fmt.Sprintf("  %-18s %s", theme.Command.Render(cmd.Name), cmd.Description))
	}
	return strings.Join(lines, "\n")
}
