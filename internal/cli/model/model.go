package model

import (
	"fmt"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anthropics/antigravity-cli/internal/cli/commands"
	"github.com/anthropics/antigravity-cli/internal/cli/layout"
	"github.com/anthropics/antigravity-cli/internal/cli/messages"
	"github.com/anthropics/antigravity-cli/internal/cli/render"
	"github.com/anthropics/antigravity-cli/internal/cli/types"
)

var _ tea.Model = RootModel{}

// RootModel is the top-level Bubble Tea model coordinating every screen.
type RootModel struct {
	Version      types.VersionInfo
	Layout       layout.Shell
	Theme        render.Theme
	Commands     *commands.Registry
	Conversation ConversationModel
	Input        InputModel
	Prompt       PromptModel
	Diff         DiffModel
	Settings     SettingsModel
	Help         HelpModel
	Agents       AgentsModel
	Artifacts    ArtifactsModel
	Statusline   StatuslineModel
	Active       string
	Status       string
}

type ConversationModel struct {
	TitleText string
	Lines     []string
}

type InputModel struct {
	Placeholder string
	Value       string
}

type PromptModel struct {
	Text string
}

type DiffModel struct {
	Files []string
}

type SettingsModel struct {
	Current map[string]string
}

type HelpModel struct {
	Commands []types.SlashCommand
}

type AgentsModel struct {
	Names []string
}

type ArtifactsModel struct {
	Items []types.Artifact
}

type StatuslineModel struct {
	Left  string
	Right string
}

func NewRootModel(info types.VersionInfo, registry *commands.Registry) RootModel {
	cmds := registry.All()
	return RootModel{
		Version:  info,
		Layout:   layout.Compute(layout.Dimensions{Width: 120, Height: 40}),
		Theme:    render.DefaultTheme(),
		Commands: registry,
		Conversation: ConversationModel{
			TitleText: "Conversation",
			Lines: []string{
				"Welcome to Antigravity CLI.",
				"This source tree is a public skeleton of the internal jetski TUI.",
			},
		},
		Input:  InputModel{Placeholder: "Ask Antigravity to inspect, edit, or explain code"},
		Prompt: PromptModel{Text: "Terminal-first coding agent with battlemode, hooks, MCP, and artifacts."},
		Diff:   DiffModel{Files: nil},
		Settings: SettingsModel{Current: map[string]string{
			"theme":       "terminal",
			"permissions": string(types.PermissionManual),
			"model":       "claude-sonnet",
		}},
		Help:       HelpModel{Commands: cmds},
		Agents:     AgentsModel{Names: []string{"task", "explore", "research"}},
		Artifacts:  ArtifactsModel{},
		Statusline: StatuslineModel{Left: info.Codename, Right: info.Version},
		Active:     "conversation",
		Status:     "ready",
	}
}

func (m RootModel) Init() tea.Cmd {
	return func() tea.Msg { return messages.Ready{} }
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.Ready:
		m.Status = "ready"
	case messages.Info:
		m.Status = msg.Text
	case messages.Warning:
		m.Status = msg.Text
	case messages.Error:
		m.Status = msg.Error()
	case tea.WindowSizeMsg:
		m.Layout = layout.Compute(layout.Dimensions{Width: msg.Width, Height: msg.Height})
	}
	return m, nil
}

func (m RootModel) View() tea.View {
	header := m.Theme.RenderHeader(m.Version.Name, m.Statusline.View())
	body := m.Conversation.View() + "\n\n" + m.Input.View()
	switch m.Active {
	case "help":
		body = m.Help.View()
	case "diff":
		body = m.Diff.View()
	case "settings":
		body = m.Settings.View()
	}
	footer := m.Theme.Subtitle.Render(m.Status)
	return tea.NewView(header + "\n\n" + body + "\n\n" + footer)
}

func (m ConversationModel) Title() string { return m.TitleText }

func (m ConversationModel) View() string {
	if len(m.Lines) == 0 {
		return "No conversation yet."
	}
	return strings.Join(m.Lines, "\n")
}

func (m InputModel) Title() string { return "Input" }

func (m InputModel) View() string {
	if m.Value == "" {
		return "> " + m.Placeholder
	}
	return "> " + m.Value
}

func (m PromptModel) Title() string { return "Prompt" }
func (m PromptModel) View() string  { return m.Text }

func (m DiffModel) Title() string { return "Diff" }

func (m DiffModel) View() string {
	if len(m.Files) == 0 {
		return "No diff staged in the stub model."
	}
	return "Changed files:\n- " + strings.Join(m.Files, "\n- ")
}

func (m SettingsModel) Title() string { return "Settings" }

func (m SettingsModel) View() string {
	if len(m.Current) == 0 {
		return "No settings available."
	}
	keys := make([]string, 0, len(m.Current))
	for k := range m.Current {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", k, m.Current[k]))
	}
	return strings.Join(lines, "\n")
}

func (m HelpModel) Title() string { return "Help" }

func (m HelpModel) View() string {
	lines := []string{"Slash commands:"}
	for _, cmd := range m.Commands {
		lines = append(lines, fmt.Sprintf("  %-18s %s", cmd.Name, cmd.Description))
	}
	return strings.Join(lines, "\n")
}

func (m AgentsModel) Title() string { return "Agents" }
func (m AgentsModel) View() string  { return strings.Join(m.Names, ", ") }

func (m ArtifactsModel) Title() string { return "Artifacts" }

func (m ArtifactsModel) View() string {
	if len(m.Items) == 0 {
		return "No artifacts captured."
	}
	lines := make([]string, 0, len(m.Items))
	for _, item := range m.Items {
		lines = append(lines, fmt.Sprintf("%s (%s)", item.Name, item.Path))
	}
	return strings.Join(lines, "\n")
}

func (m StatuslineModel) Title() string { return "Statusline" }
func (m StatuslineModel) View() string  { return m.Left + " • " + m.Right }
