package types

import (
	"fmt"
	"time"
)

type VersionInfo struct {
	Name       string
	Codename   string
	Version    string
	ModulePath string
	BuildGo    string
}

type SlashCommand struct {
	Name        string
	Description string
	Aliases     []string
}

type Session struct {
	ID        string
	Title     string
	Workspace string
	CreatedAt time.Time
}

type Artifact struct {
	ID        string
	Name      string
	Path      string
	CreatedAt time.Time
}

type AuthState struct {
	Authenticated bool
	Method        string
	Account       string
}

type BackendStatus struct {
	Connected bool
	SSH       bool
	Auth      AuthState
}

type PermissionMode string

const (
	PermissionManual           PermissionMode = "manual"
	PermissionProceedInSandbox PermissionMode = "proceed-in-sandbox"
)

type ModelInfo struct {
	Name     string
	Provider string
	Fast     bool
}

func (v VersionInfo) String() string {
	version := v.Version
	if version == "" {
		version = "dev"
	}
	return fmt.Sprintf("%s %s (%s)", v.Name, version, v.Codename)
}
