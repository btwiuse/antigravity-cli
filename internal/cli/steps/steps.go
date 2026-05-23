package steps

import (
	"fmt"

	"github.com/anthropics/antigravity-cli/internal/cortex"
)

type Kind string

const (
	StepThought  Kind = "thought"
	StepTask     Kind = "task"
	StepToolCall Kind = "tool_call"
	StepArtifact Kind = "artifact"
	StepDiff     Kind = "diff"
	StepMessage  Kind = "message"
)

type Step struct {
	ID      string
	Kind    Kind
	Title   string
	Summary string
	Done    bool
}

func (s Step) String() string {
	status := "todo"
	if s.Done {
		status = "done"
	}
	return fmt.Sprintf("[%s] %s: %s", status, s.Kind, s.Title)
}

func PlanFromConversation(conv cortex.Conversation) []Step {
	out := make([]Step, 0, len(conv.Steps))
	for i, title := range conv.Steps {
		out = append(out, Step{
			ID:      fmt.Sprintf("%s-%d", conv.ID, i+1),
			Kind:    StepTask,
			Title:   title,
			Summary: title,
			Done:    i < len(conv.Steps)-1,
		})
	}
	return out
}
