package commands

import (
	"sort"
	"strings"

	"github.com/anthropics/antigravity-cli/internal/cli/types"
)

var DefaultCommands = []types.SlashCommand{
	{Name: "/help", Description: "Show help and slash command reference"},
	{Name: "/exit", Description: "Exit the current session"},
	{Name: "/clear", Description: "Clear the conversation transcript"},
	{Name: "/resume", Description: "Resume a saved conversation"},
	{Name: "/settings", Description: "Open settings and preferences"},
	{Name: "/model", Description: "Select the active model"},
	{Name: "/diff", Description: "Inspect the pending diff"},
	{Name: "/logout", Description: "Clear saved credentials"},
	{Name: "/agents", Description: "Inspect available agents and subagents"},
	{Name: "/artifacts", Description: "Browse generated artifacts"},
	{Name: "/btw", Description: "Open the internal BTW scratchpad"},
	{Name: "/changelog", Description: "Show recent release notes"},
	{Name: "/config", Description: "Inspect merged configuration"},
	{Name: "/context", Description: "View active context and attachments"},
	{Name: "/copy", Description: "Copy the latest assistant output"},
	{Name: "/fast", Description: "Toggle fast mode for lower-latency responses"},
	{Name: "/feedback", Description: "Send product feedback"},
	{Name: "/fork", Description: "Fork the conversation into a new branch"},
	{Name: "/hooks", Description: "Manage lifecycle hooks"},
	{Name: "/keybindings", Description: "Show keyboard shortcuts"},
	{Name: "/mcp", Description: "Inspect MCP servers and tools"},
	{Name: "/open", Description: "Open a file, URL, or artifact"},
	{Name: "/permissions", Description: "Set tool permission behavior"},
	{Name: "/planning", Description: "Toggle visible planning output"},
	{Name: "/rename", Description: "Rename the current chat title"},
	{Name: "/rewind", Description: "Rewind to an earlier checkpoint"},
	{Name: "/skills", Description: "List installed skills and plugins"},
	{Name: "/statusline", Description: "Customize the statusline"},
	{Name: "/tasks", Description: "Inspect active background tasks"},
	{Name: "/title", Description: "Show or set the session title"},
	{Name: "/usage", Description: "Display usage and quota information"},
	{Name: "/addworkspacedir", Description: "Add another workspace directory"},
}

type Registry struct {
	ordered  []types.SlashCommand
	commands map[string]types.SlashCommand
}

func NewRegistry() *Registry {
	r := &Registry{
		ordered:  append([]types.SlashCommand(nil), DefaultCommands...),
		commands: make(map[string]types.SlashCommand, len(DefaultCommands)),
	}
	for _, cmd := range r.ordered {
		r.commands[Normalize(cmd.Name)] = cmd
		for _, alias := range cmd.Aliases {
			r.commands[Normalize(alias)] = cmd
		}
	}
	sort.Slice(r.ordered, func(i, j int) bool {
		return r.ordered[i].Name < r.ordered[j].Name
	})
	return r
}

func (r *Registry) All() []types.SlashCommand {
	out := make([]types.SlashCommand, len(r.ordered))
	copy(out, r.ordered)
	return out
}

func (r *Registry) Lookup(name string) (types.SlashCommand, bool) {
	cmd, ok := r.commands[Normalize(name)]
	return cmd, ok
}

func Normalize(name string) string {
	return strings.TrimPrefix(strings.ToLower(strings.TrimSpace(name)), "/")
}
