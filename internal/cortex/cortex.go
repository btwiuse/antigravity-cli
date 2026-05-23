package cortex

import (
	"context"
	"fmt"
	"time"
)

type Request struct {
	Prompt    string
	Workspace string
}

type Conversation struct {
	ID        string
	CreatedAt time.Time
	Steps     []string
}

type Capabilities struct {
	BattleMode     bool
	Subagents      bool
	Artifacts      bool
	Knowledge      bool
	Hooks          bool
	MCP            bool
	SDK            bool
	LanguageServer bool
}

type KnowledgeBase interface {
	Search(context.Context, string) ([]string, error)
}

type HookRunner interface {
	BeforeTool(context.Context, string) error
	AfterTool(context.Context, string, error) error
}

type SDK interface {
	RegisterTool(name string, fn any) error
}

type LanguageServer interface {
	IndexWorkspace(context.Context, string) error
	Symbols(context.Context, string) ([]string, error)
}

type Engine interface {
	Name() string
	StartConversation(context.Context, Request) (Conversation, error)
	Capabilities() Capabilities
	ListSubagents(context.Context) []string
}

// AgentEngine is a public stub of the internal cortex agent engine.
type AgentEngine struct {
	Codename string
	battle   bool
	lsp      LanguageServer
	kb       KnowledgeBase
	hooks    HookRunner
	sdk      SDK
}

func NewEngine() *AgentEngine {
	return &AgentEngine{Codename: "jetski", battle: true}
}

func (e *AgentEngine) Name() string {
	return "cortex"
}

func (e *AgentEngine) StartConversation(_ context.Context, req Request) (Conversation, error) {
	prompt := req.Prompt
	if prompt == "" {
		prompt = "help"
	}
	return Conversation{
		ID:        fmt.Sprintf("conv-%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
		Steps: []string{
			"analyze request",
			"plan tool usage",
			"respond to prompt: " + prompt,
		},
	}, nil
}

func (e *AgentEngine) Capabilities() Capabilities {
	return Capabilities{
		BattleMode:     e.battle,
		Subagents:      true,
		Artifacts:      true,
		Knowledge:      true,
		Hooks:          true,
		MCP:            true,
		SDK:            true,
		LanguageServer: true,
	}
}

func (e *AgentEngine) ListSubagents(context.Context) []string {
	return []string{"task", "explore", "research"}
}
