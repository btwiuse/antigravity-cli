package entrypoints

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/anthropics/antigravity-cli/internal/cli/backend"
	"github.com/anthropics/antigravity-cli/internal/cli/commands"
	"github.com/anthropics/antigravity-cli/internal/cli/model"
	"github.com/anthropics/antigravity-cli/internal/cli/render"
	"github.com/anthropics/antigravity-cli/internal/cli/store"
	"github.com/anthropics/antigravity-cli/internal/cli/types"
	"github.com/anthropics/antigravity-cli/internal/cortex"
)

// App wires the reverse-engineered CLI surface to public Go module paths.
type App struct {
	Info     types.VersionInfo
	Registry *commands.Registry
	Backend  backend.Service
	Store    *store.MemoryStore
}

func NewApp(version string) *App {
	info := types.VersionInfo{
		Name:       "Antigravity CLI",
		Codename:   "jetski",
		Version:    version,
		ModulePath: "github.com/anthropics/antigravity-cli",
		BuildGo:    "go1.27-rc04 (google3 reverse-engineered)",
	}
	engine := cortex.NewEngine()

	return &App{
		Info:     info,
		Registry: commands.NewRegistry(),
		Backend: backend.New(backend.Config{
			Endpoint:    "https://antigravity.invalid/api",
			OAuthURL:    "https://antigravity.invalid/oauth",
			ProductName: info.Name,
		}, engine),
		Store: store.NewMemoryStore(),
	}
}

func (a *App) Version() string {
	return a.Info.String()
}

func (a *App) Help() string {
	return render.RenderHelp(a.Info, a.Registry.All())
}

func (a *App) Run(ctx context.Context, args []string, out io.Writer) error {
	if out == nil {
		out = os.Stdout
	}
	if len(args) == 0 {
		_, err := fmt.Fprintln(out, a.Help())
		return err
	}

	switch strings.ToLower(args[0]) {
	case "help", "/help", "-h", "--help":
		_, err := fmt.Fprintln(out, a.Help())
		return err
	case "version", "--version", "-version":
		_, err := fmt.Fprintln(out, a.Version())
		return err
	case "tui":
		return a.RunTUI()
	default:
		conv, err := a.Backend.Engine().StartConversation(ctx, cortex.Request{
			Prompt:    strings.Join(args, " "),
			Workspace: ".",
		})
		if err != nil {
			return err
		}
		a.Store.SaveSession(types.Session{
			ID:        conv.ID,
			Title:     strings.Join(args, " "),
			Workspace: ".",
			CreatedAt: conv.CreatedAt,
		})
		_, err = fmt.Fprintf(out, "started %s with %d planned steps\n", conv.ID, len(conv.Steps))
		return err
	}
}

func (a *App) RunTUI() error {
	root := model.NewRootModel(a.Info, a.Registry)
	program := tea.NewProgram(root)
	_, err := program.Run()
	return err
}
