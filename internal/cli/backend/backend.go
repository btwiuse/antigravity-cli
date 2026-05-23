package backend

import (
	"context"
	"errors"

	"github.com/anthropics/antigravity-cli/internal/cli/backend/auth"
	"github.com/anthropics/antigravity-cli/internal/cli/types"
	"github.com/anthropics/antigravity-cli/internal/cortex"
)

type Service interface {
	Status(context.Context) (types.BackendStatus, error)
	EnsureAuthenticated(context.Context) (auth.Session, error)
	DetectSSH() bool
	Engine() cortex.Engine
}

type Config struct {
	Endpoint    string
	OAuthURL    string
	ProductName string
}

type Client struct {
	cfg       Config
	providers []auth.Provider
	ssh       auth.SSHDetector
	engine    cortex.Engine
}

func New(cfg Config, engine cortex.Engine) *Client {
	return &Client{
		cfg: cfg,
		providers: []auth.Provider{
			&auth.KeyringProvider{Service: cfg.ProductName},
			auth.OAuthProvider{AuthorizeURL: cfg.OAuthURL},
		},
		ssh:    auth.EnvSSHDetector{},
		engine: engine,
	}
}

func (c *Client) Status(ctx context.Context) (types.BackendStatus, error) {
	status := types.BackendStatus{Connected: true, SSH: c.DetectSSH()}
	for _, provider := range c.providers {
		session, err := provider.Load(ctx)
		if err == nil {
			status.Auth = types.AuthState{
				Authenticated: true,
				Method:        session.Method,
				Account:       session.Account,
			}
			break
		}
	}
	return status, nil
}

func (c *Client) EnsureAuthenticated(ctx context.Context) (auth.Session, error) {
	for _, provider := range c.providers {
		session, err := provider.Load(ctx)
		if err == nil {
			return session, nil
		}
	}
	for _, provider := range c.providers {
		session, err := provider.Login(ctx)
		if err == nil {
			return session, nil
		}
	}
	return auth.Session{}, errors.New("unable to authenticate with any provider")
}

func (c *Client) DetectSSH() bool {
	if c.ssh == nil {
		return false
	}
	return c.ssh.Remote()
}

func (c *Client) Engine() cortex.Engine {
	return c.engine
}
