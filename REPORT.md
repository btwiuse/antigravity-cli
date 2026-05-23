# Reverse Engineering Report: Antigravity CLI (`agy`)

## Executive Summary

The `agy` binary is the **Antigravity CLI** — a terminal-first coding agent built on the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework. Internally codenamed **"jetski"**, it is developed inside Google's monorepo (`google3/`) and compiled with a custom Go 1.27 RC04 toolchain. The CLI provides multi-step AI agent capabilities (code editing, tool calling, browser automation, subagents) directly in the terminal, optimized for SSH and keyboard-driven workflows.

This report documents the reverse engineering process and the resulting source tree reconstruction.

---

## 1. Binary Metadata

| Property | Value |
|---|---|
| **File** | `agy` (ELF 64-bit LSB PIE, x86-64) |
| **Size** | 182,743,320 bytes (~174 MB) |
| **Linking** | Dynamically linked (`ld-linux-x86-64.so.2`) |
| **Stripped** | Yes |
| **Go Version** | `go1.27-20260427-RC04 cl/906595525 +5fb2392a6f` |
| **Build Tags** | `fieldtrack,boringcrypto,simd` |
| **Build System** | Google Blaze (google3 monorepo) |

## 2. Reverse Engineering Methodology

### 2.1 String Extraction

The primary technique was analyzing the output of `strings agy` (1,562,750 lines). Go binaries retain rich metadata:

- **Function symbol names** — Go embeds fully qualified function/method names for stack traces and reflection
- **Package paths** — Every imported package's path appears in the binary
- **Type information** — struct field names, interface method sets, and generic type instantiations
- **Embedded string literals** — error messages, help text, command names, format strings
- **Proto field descriptors** — protobuf-generated code includes field names and message types

### 2.2 Package Path Analysis

Go binaries built inside Google's monorepo use `google3/` prefixed import paths. By extracting and deduplicating these paths, the complete dependency graph was reconstructed:

```bash
grep -oP 'google3/[a-zA-Z0-9_/]+' strings-agy.txt | sort -u
```

This yielded **230+ internal packages** and **1,002 third-party dependency paths**.

### 2.3 Method Signature Recovery

Go's runtime type information preserves method signatures in the format:
```
package.(*TypeName).MethodName
```

By extracting these patterns:
```bash
grep -oP '\.\(\*[A-Za-z]+\)\.[A-Za-z]+' strings-agy.txt | sort -u
```

We recovered every exported and many unexported methods for all types in the binary.

### 2.4 Dependency Mapping

Third-party dependencies were identified by the `google3/third_party/golang/` prefix, which maps to standard Go modules:

| google3 Path | Public Module |
|---|---|
| `charm_land/bubbletea/v/v2` | `charm.land/bubbletea/v2` |
| `charm_land/lipgloss/v/v2` | `charm.land/lipgloss/v2` |
| `charm_land/bubbles/v/v2` | `charm.land/bubbles/v2` |
| `golang/goldmark` | `github.com/yuin/goldmark` |
| `golang/goquery` | `github.com/PuerkitoBio/goquery` |
| `golang/gorm` | `gorm.io/gorm` |
| `golang/sqlite3` | `github.com/mattn/go-sqlite3` |
| `golang/gitv5` | `github.com/go-git/go-git/v5` |
| `golang/chromedp` → `cdp` | `github.com/chromedp/cdproto` |
| `golang/regexp2` | `github.com/dlclark/regexp2` |
| `golang/yamlv3` | `gopkg.in/yaml.v3` |
| `golang/oauth2` | `golang.org/x/oauth2` |
| `golang/grpc` | `google.golang.org/grpc` |
| `golang/protobuf` | `google.golang.org/protobuf` |
| `golang/connectrpc_com` | `connectrpc.com/connect` |
| `golang/fsnotify` | `github.com/fsnotify/fsnotify` |
| `golang/pty` | `github.com/creack/pty` |
| `golang/bluemonday` | `github.com/microcosm-cc/bluemonday` |
| `golang/difflib` | `github.com/pmezard/go-difflib` |
| `golang/zap` | `go.uber.org/zap` |

---

## 3. Architecture Overview

### 3.1 High-Level Architecture

```
┌────────────────────────────────────────────────────────────┐
│                        main.go                             │
│  (flags: -version, -help, -tui, -prompt, -resume)          │
└────────────────┬───────────────────────────────────────────┘
                 │
┌────────────────▼───────────────────────────────────────────┐
│              cli/entrypoints                                │
│  MaybeLaunchCLI · launchCLI · loadTrajectoryFromPath       │
└────────┬─────────────┬───────────────────┬─────────────────┘
         │             │                   │
┌────────▼──┐  ┌───────▼───────┐  ┌───────▼───────┐
│ cli/model │  │ cli/backend   │  │ cli/commands  │
│ (TUI)     │  │ (ServerBackend│  │ (slash cmds)  │
│           │  │  + Auth)      │  │               │
└─────┬─────┘  └───────┬───────┘  └───────────────┘
      │                │
      │        ┌───────▼───────┐
      │        │    cortex     │
      │        │ (Agent Engine)│
      │        └───────┬───────┘
      │                │
      │    ┌───────────┼───────────┐
      │    │           │           │
      ▼    ▼           ▼           ▼
  bubbletea v2    gRPC/Connect   Protobuf
  lipgloss v2     (API layer)    (state)
  bubbles v2
```

### 3.2 Package Tree (Internal)

```
google3/third_party/jetski/
├── cli/                           # Terminal UI layer
│   ├── analytics/                 # Event recording & telemetry
│   ├── backend/                   # ServerBackend (core business logic bridge)
│   │   └── auth/                  # OAuth, Keyring, SSH, WSL auth chain
│   ├── changelog/                 # Version changelog rendering
│   ├── commands/                  # 32+ slash commands (/help, /exit, /model, etc.)
│   ├── editing/                   # Text editing utilities
│   ├── entrypoints/               # CLI launch, config, resume
│   ├── graphic/                   # Braille-encoded logo, color schemes
│   ├── installer/                 # Self-update mechanism
│   ├── keybindings/               # Configurable key bindings
│   ├── launchparams/              # CLI flag parsing & env detection
│   ├── layout/                    # Terminal layout calculation
│   ├── logfile/                   # Log file management
│   ├── logo/                      # ASCII art logo
│   ├── mermaid/                   # Mermaid diagram rendering
│   ├── messages/                  # Bubble Tea message types
│   ├── model/                     # 30+ Bubble Tea TUI models
│   ├── modelresolver/             # AI model selection logic
│   ├── onboarding/                # First-run onboarding flow
│   ├── plugins/                   # Plugin system
│   │   ├── claude/                # Claude model plugin
│   │   ├── common/                # Shared plugin utilities
│   │   ├── gemini/                # Gemini model plugin
│   │   └── marketplace/           # Plugin marketplace
│   ├── printmode/                 # Non-interactive print mode
│   ├── project/                   # Project detection
│   ├── render/                    # Markdown & diff rendering
│   ├── steps/                     # Agent step visualization
│   ├── store/                     # Persistent state (SQLite via GORM)
│   ├── types/                     # Shared type definitions
│   ├── updater/                   # Auto-update checker
│   └── viewport/                  # Custom viewport component
│
├── cortex/                        # Agent execution engine
│   ├── agent_state_component/     # Agent lifecycle & state management
│   ├── agentapi/                  # Agent API handlers
│   ├── artifacts/                 # File artifact management
│   │   └── knowledge/             # Knowledge base integration
│   ├── battlemode/                # "Battle mode" autonomous execution
│   ├── cascade_run_state/         # Multi-step cascade state
│   ├── chatconverters/            # Chat format converters
│   │   ├── browser/               # Browser action converter
│   │   └── code/                  # Code action converter
│   ├── command/                   # Shell command execution
│   │   ├── exebox/                # Executable sandbox
│   │   ├── sandboxproxy/          # Sandbox proxy
│   │   ├── sbox/                  # Sandbox primitives
│   │   └── urpc/                  # Unix RPC for sandbox
│   ├── config/                    # Agent configuration
│   ├── customization/             # Agent customization hooks
│   │   └── hooks/                 # Lifecycle hook system
│   ├── customizations/            # Built-in customizations
│   │   └── builtin/teamwork/      # Multi-agent teamwork
│   ├── executors/                 # Step executors with hook pipeline
│   │   ├── posthooks/
│   │   ├── posttoolhooks/
│   │   ├── prehooks/
│   │   ├── pretoolhooks/
│   │   └── stophooks/
│   ├── fastapply/                 # Fast file apply (streaming edits)
│   ├── handlers/                  # Tool handlers
│   │   ├── browser/               # Browser automation (Chrome DevTools)
│   │   ├── ephemeral/             # Ephemeral tool execution
│   │   └── imagegen/              # Image generation
│   ├── implicit/                  # Implicit context injection
│   ├── messages/                  # Agent message types
│   ├── mixins/                    # Composable agent capabilities
│   │   ├── browser/               # Browser mixin
│   │   ├── knowledge/             # Knowledge mixin
│   │   └── types/                 # Mixin type definitions
│   ├── permissions/               # Tool permission management
│   ├── sdk/                       # SDK integration
│   │   └── sdkprocess/            # SDK process management
│   ├── shared/                    # Shared cortex utilities
│   ├── sharedenv/                 # Shared environment config
│   ├── sharedfragment/            # Shared fragment handling
│   ├── sidecars/                  # Sidecar process management
│   ├── slashcommands/             # Agent-level slash commands
│   ├── snapshot_recorder/         # State snapshot recording
│   ├── state/                     # Agent state machine
│   ├── subagent/                  # Sub-agent spawning & management
│   ├── tokens/                    # Token counting & management
│   ├── tools/                     # Tool implementations
│   │   ├── browser/               # Browser tool
│   │   ├── code/                  # Code editing tool
│   │   ├── knowledge/             # Knowledge retrieval tool
│   │   ├── notebook/              # Notebook tool
│   │   ├── subagent/              # Sub-agent tool
│   │   └── tooloverrides/         # Tool behavior overrides
│   ├── traj/                      # Trajectory management
│   ├── trajectory/                # Conversation trajectory
│   │   ├── dbtrajectory/          # Database-backed trajectory
│   │   └── internal/              # Internal trajectory helpers
│   ├── trajectory_store/          # Trajectory persistence
│   └── utils/                     # Cortex utilities
│       ├── chrome_devtools_client/# CDP client
│       ├── commandutils/          # Command helpers
│       ├── go_sed/                # Streaming sed-like editor
│       ├── mcp/                   # Model Context Protocol client
│       │   └── mcpcore/           # MCP core primitives
│       ├── overrides/             # Runtime overrides
│       ├── scroll_utils/          # Scroll position utilities
│       ├── urlutils/              # URL parsing
│       ├── webm/                  # WebM video handling
│       └── workflowparse/         # Workflow file parsing
│
├── language_server/               # LSP-like language intelligence
│   ├── auth_client/               # Auth client
│   ├── browser/                   # Browser integration
│   ├── chat/                      # Chat provider
│   ├── code_assist_client/        # Code assist API client
│   ├── codesearch/                # Code search
│   ├── completion_provider/       # Code completion
│   ├── documentmanager/           # Document tracking
│   ├── fragments/                 # Code fragment management
│   ├── language_utils/            # Language detection
│   ├── lsp/                       # LSP protocol
│   ├── metadata_provider/         # File metadata
│   ├── modelapigemini/            # Gemini model API
│   ├── projects/                  # Project management
│   ├── sdk_executor/              # SDK command executor
│   ├── state/                     # Server state
│   ├── state_sync/                # State synchronization
│   ├── streaming/                 # Streaming responses
│   ├── telemetry/                 # Usage telemetry
│   └── worktree/                  # Git worktree tracking
│
├── prompt/                        # Prompt engineering
│   ├── template_provider/         # Go template-based prompts
│   │   └── templates/             # System prompts
│   │       ├── system_prompts/    # Identity, guidelines, skills, etc.
│   │       ├── step_strings/      # Step rendering templates
│   │       └── helpers/           # Template helper functions
│   └── cumulative_prompt_handler/ # Cumulative prompt building
│
├── fs/                            # Filesystem abstraction
│   ├── file_watcher/              # File change watcher
│   ├── local/                     # Local filesystem
│   └── workspace_manager/         # Workspace management
│
├── git_utils/                     # Git operations
│   └── git_cci/                   # Git CCI integration
├── gitignore/                     # .gitignore parsing
├── vcs/                           # Version control abstraction
├── scm_utils/                     # SCM utilities
│
├── transport/                     # Network transport
├── remoting/                      # Remote session support
│   ├── reversetunnel/             # Reverse tunnel for remote access
│   └── ssh/                       # SSH connection handling
│
├── experiments/                   # Feature flag experiments
├── unleash/                       # Feature toggle (Unleash)
├── tokenizer/                     # Token counting
├── radix/                         # Radix tree
├── document/                      # Document model
├── web_scraping/                  # Web content extraction
├── terminal/                      # Terminal capabilities
├── log/                           # Logging
│
└── *_pb/                          # Protobuf definitions (20+ packages)
    ├── analytics_pb/
    ├── api_server_pb/
    ├── browser_pb/
    ├── cascade_plugins_pb/
    ├── chat_pb/
    ├── chat_client_server_pb/
    ├── code_edit_pb/
    ├── codeium_common_pb/
    ├── config_pb/
    ├── context_module_pb/
    ├── cortex_pb/
    ├── dev_pb/
    ├── diff_action_pb/
    ├── eval_pb/
    ├── extension_server_pb/
    ├── hooks_pb/
    ├── index_pb/
    ├── jetbox_state_pb/
    ├── jetbox_summaries_pb/
    ├── jetski_cortex_pb/
    ├── language_server_pb/
    ├── model_management_pb/
    ├── opensearch_clients_pb/
    ├── project_pb/
    ├── prompt_pb/
    ├── reactive_component_pb/
    ├── remoting_pb/
    ├── seat_management_pb/
    ├── trainer_pb/
    └── unified_state_sync_pb/
```

### 3.3 TUI Model Hierarchy

The Bubble Tea TUI uses a component-based architecture. All models implement a common interface:

```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
    Layout() Layout
    SetLayout(Layout)
    Store() *Store
    UIPhase() UIPhase
}
```

**30+ identified TUI models:**

| Model | Purpose |
|---|---|
| `RootModel` | Top-level model, routes between views |
| `ConversationModel` | Main chat view with scrollable messages |
| `InputModel` | Multi-line text input with history |
| `PromptModel` | Prompt composition with @-mentions |
| `DiffModel` | Side-by-side/unified diff viewer |
| `SettingsModel` | Settings panel |
| `HelpModel` | Help reference view |
| `AuthModel` | Authentication flow |
| `ConversationPickerModel` | Session resume picker |
| `ArtifactViewModel` | Artifact browser |
| `ArtifactDetailModel` | Single artifact detail view |
| `ArtifactReviewModel` | Artifact review workflow |
| `AgentsModel` | Agent list and inspector |
| `SkillsModel` | Skills/plugin browser |
| `MCPModel` | MCP server inspector |
| `TasksModel` | Background task monitor |
| `PermissionsModel` | Tool permission editor |
| `SelectorModel` | Generic selection list |
| `RewindModel` | Checkpoint rewind picker |
| `FeedbackModel` | Feedback submission |
| `ProactiveFeedbackModel` | Auto-triggered feedback |
| `ChangelogModel` | Changelog viewer |
| `ContextModel` | Context inspector |
| `HooksModel` | Hook editor |
| `UsageModel` | Usage/quota display |
| `ToolConfirmationModel` | Tool approval dialog |
| `AskQuestionModel` | Agent question dialog |
| `WorkspaceTrustModel` | Workspace trust prompt |
| `StatusLineModel` | Custom status line |
| `SubagentDetailModel` | Sub-agent detail view |
| `SuggestionModel` | Command autocomplete |
| `AtSuggestionModel` | @-mention autocomplete |
| `BugReportModel` | Bug report form |
| `McpAuthModel` | MCP authentication |
| `SettingsErrorModel` | Settings error display |

### 3.4 Slash Commands

32 slash commands were identified:

| Command | Type | Description |
|---|---|---|
| `/help` | `helpCommand` | Show help and command reference |
| `/exit` | `exitCommand` | Exit session (aliases: `/quit`) |
| `/clear` | `clearCommand` | Clear transcript (aliases: `/new`) |
| `/resume` | `resumeCommand` | Resume a saved conversation |
| `/settings` | `configCommand` | Open settings (aliases: `/config`) |
| `/model` | `modelCommand` | Select active model |
| `/diff` | `diffCommand` | Inspect pending diff |
| `/logout` | `logoutCommand` | Clear credentials |
| `/agents` | `agentsCommand` | Inspect agents |
| `/artifacts` | `artifactCommand` | Browse artifacts |
| `/btw` | `btwCommand` | BTW scratchpad |
| `/changelog` | `changelogCommand` | Release notes |
| `/context` | `contextCommand` | View context |
| `/copy` | `copyCommand` | Copy assistant output |
| `/fast` | `fastCommand` | Toggle fast mode |
| `/feedback` | `feedbackCommand` | Send feedback (aliases: `/bug`) |
| `/fork` | `forkCommand` | Fork conversation (aliases: `/branch`) |
| `/hooks` | `hooksCommand` | Manage hooks |
| `/keybindings` | `keybindingsCommand` | Show shortcuts |
| `/mcp` | `mcpCommand` | MCP servers |
| `/open` | `openCommand` | Open file/URL |
| `/permissions` | `permissionsCommand` | Tool permissions |
| `/planning` | `planningCommand` | Toggle planning |
| `/rename` | `renameCommand` | Rename chat |
| `/rewind` | `rewindCommand` | Rewind to checkpoint |
| `/skills` | `skillsCommand` | List skills |
| `/statusline` | `statuslineCommand` | Customize statusline |
| `/tasks` | `tasksCommand` | Background tasks |
| `/title` | `titleCommand` | Show/set title |
| `/usage` | `usageCommand` | Usage & quota |
| `/addworkspacedir` | `addWorkspaceDirCommand` | Add workspace dir |

---

## 4. Authentication Architecture

The auth system uses a **chained authentication** pattern:

1. **Keyring Auth** (`keyringAuth`) — System keyring (gnome-keyring, macOS Keychain, Windows Credential Manager)
2. **OAuth Auth** (`oauthMethod`) — Google OAuth 2.0 with PKCE
   - Local: Opens browser automatically
   - SSH: Prints authorization URL for manual completion
3. **External Corp Login** (`externalCorpLoginAuth`) — Enterprise SSO

Environment detectors:
- `sshDetector` — Detects SSH sessions via `$SSH_CONNECTION`
- `wslDetector` — Detects Windows Subsystem for Linux
- `containerDetector` — Detects container environments

Token storage:
- `cliTokenStorage` — In-memory token cache
- `cliFileTokenStorage` — File-based token persistence

---

## 5. Agent Engine (Cortex)

The **Cortex** is the core agent execution engine shared between the CLI and the full GUI. Key components:

### 5.1 Agent State Machine
- `AgentState` manages lifecycle with states: idle → executing → generating → tool_calling → idle
- `CostAggregator` tracks token costs per turn
- `SubagentUpdateForwarder` propagates state between parent and child agents

### 5.2 Execution Pipeline
```
prehooks → executor → tool call → pretoolhooks → tool handler →
posttoolhooks → result → posthooks → stophooks
```

### 5.3 Tool System
Tools discovered in the binary:
- **Code tools**: file read/write, search, apply diff, go_sed
- **Browser tools**: Chrome DevTools Protocol automation (navigate, click, screenshot, scrape)
- **Knowledge tools**: context retrieval, code search, fragment management
- **Notebook tools**: interactive notebook execution
- **Subagent tools**: spawn specialized sub-agents
- **Command tools**: shell execution with sandbox isolation

### 5.4 Prompt System
Template-based prompt engineering with Go `text/template`:
- `system_prompts/identity` — Agent identity and capabilities
- `system_prompts/guidelines` — Behavioral guidelines
- `system_prompts/function_call` — Tool calling format
- `system_prompts/skills` — Available skills
- `system_prompts/plugins` — Plugin capabilities
- `system_prompts/artifacts` — Artifact handling
- `system_prompts/knowledge_items` — Knowledge base context
- `system_prompts/planning_mode` — Planning mode behavior

### 5.5 MCP (Model Context Protocol)
- `mcpcore` — Core MCP primitives
- `mcp` — MCP client for connecting to external tool servers

### 5.6 Sandbox
Terminal commands run in a sandboxed environment:
- `sbox` — Sandbox primitives
- `sandboxproxy` — Proxy for sandbox communication
- `exebox` — Executable sandbox wrapper
- `urpc` — Unix RPC for sandbox IPC

---

## 6. Third-Party Dependencies

### 6.1 TUI Framework
| Library | Purpose |
|---|---|
| `charm.land/bubbletea/v2` | Elm-architecture TUI framework |
| `charm.land/lipgloss/v2` | Declarative terminal styling |
| `charm.land/bubbles/v2` | Pre-built TUI components (textarea, spinner, viewport) |

### 6.2 Data & Storage
| Library | Purpose |
|---|---|
| `gorm.io/gorm` | ORM for conversation/trajectory persistence |
| `github.com/mattn/go-sqlite3` | SQLite driver |

### 6.3 Git & VCS
| Library | Purpose |
|---|---|
| `github.com/go-git/go-git/v5` | Pure-Go git implementation |
| `github.com/go-git/go-billy/v5` | Filesystem abstraction for git |

### 6.4 Network & API
| Library | Purpose |
|---|---|
| `google.golang.org/grpc` | gRPC client/server |
| `connectrpc.com/connect` | Connect RPC (gRPC-compatible HTTP) |
| `golang.org/x/oauth2` | OAuth 2.0 client |

### 6.5 Content Processing
| Library | Purpose |
|---|---|
| `github.com/yuin/goldmark` | Markdown rendering |
| `github.com/PuerkitoBio/goquery` | HTML parsing/scraping |
| `github.com/microcosm-cc/bluemonday` | HTML sanitization |
| `github.com/chromedp/cdproto` | Chrome DevTools Protocol |

### 6.6 Other Notable Dependencies
| Library | Purpose |
|---|---|
| `go.uber.org/zap` | Structured logging |
| `github.com/fsnotify/fsnotify` | Filesystem notifications |
| `github.com/creack/pty` | PTY for terminal command execution |
| `github.com/dlclark/regexp2` | .NET-compatible regex |
| `github.com/go-enry/go-enry` | Programming language detection |
| `github.com/nfnt/resize` | Image resizing |
| `github.com/fogleman/gg` | 2D graphics |

---

## 7. Protobuf Schema Overview

20+ protobuf packages define the data model:

| Proto Package | Key Messages |
|---|---|
| `cortex_pb` | `CascadeConfig`, `TrajectoryConversionConfig` |
| `chat_pb` | Chat messages and conversation state |
| `analytics_pb` | Telemetry events, completion recording |
| `language_server_pb` | LSP-like requests and responses |
| `config_pb` | Application configuration |
| `hooks_pb` | Lifecycle hook definitions |
| `prompt_pb` | Prompt template parameters |
| `remoting_pb` | Remote session protocol |
| `code_edit_pb` | Code edit operations |
| `diff_action_pb` | Diff action descriptors |
| `project_pb` | Project configuration |
| `browser_pb` | Browser automation commands |

---

## 8. Notable Internal Features

### 8.1 "Battle Mode"
An autonomous execution mode (`cortex/battlemode/`) where the agent can proceed without human confirmation for each step.

### 8.2 Subagent System
The agent can spawn specialized sub-agents (`cortex/subagent/`) for parallel task execution, with state forwarding and cost aggregation.

### 8.3 Artifact Review Mode
A review workflow (`ArtifactReviewModel`) where users can approve/reject file changes one by one.

### 8.4 Terminal Sandbox
Commands execute in an isolated sandbox (`cortex/command/sbox/`) with configurable domain allowlists and a sentinel key system.

### 8.5 Plugin Architecture
Supports plugins for different AI model providers:
- `plugins/gemini/` — Google Gemini
- `plugins/claude/` — Anthropic Claude
- `plugins/marketplace/` — Plugin marketplace

### 8.6 Hooks System
Lifecycle hooks (`cortex/customization/hooks/`) with JSON hook definitions, supporting pre/post tool execution customization.

### 8.7 Fast Apply
Streaming file edits (`cortex/fastapply/`) for applying large diffs without full file rewrites.

### 8.8 Google Cloud AI Platform Integration
Direct integration with Vertex AI prediction services via `google.cloud.aiplatform.v1beta1` and `master` API versions.

---

## 9. Reconstructed Source Tree

The compilable source tree is available in this repository:

```
├── go.mod                              # Go module definition
├── go.sum                              # Dependency checksums
├── main.go                             # CLI entrypoint
└── internal/
    ├── cli/
    │   ├── backend/
    │   │   ├── auth/auth.go            # Authentication chain
    │   │   └── backend.go              # ServerBackend
    │   ├── commands/commands.go         # Slash command registry
    │   ├── entrypoints/entrypoints.go   # App lifecycle
    │   ├── graphic/graphic.go           # Logo and graphics
    │   ├── keybindings/keybindings.go   # Key binding config
    │   ├── layout/layout.go            # Layout calculation
    │   ├── messages/messages.go         # TUI messages
    │   ├── model/model.go              # All TUI models
    │   ├── render/render.go            # Rendering utilities
    │   ├── steps/steps.go              # Step visualization
    │   ├── store/store.go              # Persistent storage
    │   ├── types/types.go              # Shared types
    │   └── viewport/viewport.go        # Custom viewport
    └── cortex/
        └── cortex.go                   # Agent engine core
```

### Build & Run

```bash
go build ./...
./antigravity-cli -version
./antigravity-cli -help
```

---

## 10. Observations & Analysis

### 10.1 Codebase Scale
The binary contains symbols from **200+ internal packages** and **90+ third-party Go modules**, suggesting a codebase of roughly **100,000–200,000 lines of Go** (excluding generated protobuf code).

### 10.2 Internal Codename
The project is internally called **"jetski"** (all internal paths use `google3/third_party/jetski/`). The public branding is "Antigravity CLI."

### 10.3 Codeium Heritage
Several packages reference `codeium_common_pb`, suggesting the project evolved from or shares infrastructure with Codeium (a code completion tool).

### 10.4 Google-Specific Integrations
The binary includes deep Google infrastructure dependencies:
- **Clearcut** — Google's client-side logging system
- **Thinmint/Ubermint** — Identity management
- **LOAS/Zatar** — Internal authentication
- **Borglet** — Borg container integration
- **GCE/GCP** — Cloud platform integration
- **Piper** — Google's internal VCS client

### 10.5 Security Considerations
- The binary uses **BoringCrypto** (Google's FIPS-validated crypto module)
- Terminal sandbox with domain allowlists
- Token storage in system keyring with file fallback
- Safe archive extraction (`security/safearchive`)
- Safe text rendering (`security/safetext`)

### 10.6 Go Version
Built with **Go 1.27 RC04** (an internal pre-release), significantly ahead of the current public Go release. The `go.mod` in the reconstructed source uses Go 1.25 for compatibility with current public Charm library versions.

---

## 11. Limitations of This Analysis

1. **No control flow recovery** — String analysis reveals structure but not logic. Function bodies cannot be reconstructed.
2. **Stripped binary** — Debug symbols are removed; only runtime-essential metadata remains.
3. **Protobuf schemas incomplete** — Only field names and message names could be recovered; full `.proto` files would require disassembly.
4. **Google-internal dependencies** — Many packages (`google3/base/`, `google3/security/`, etc.) have no public equivalents.
5. **Embedded assets** — Prompt templates, HTML templates, and configuration defaults are embedded in the binary but were not fully extracted.

---

*Report generated by reverse engineering analysis of `agy` binary (SHA256 of agy.tar.xz).*
*Analysis date: 2026-05-23*
