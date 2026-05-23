package auth

import (
	"context"
	"errors"
	"os"
	"time"
)

var ErrNoSession = errors.New("no saved auth session")

type Session struct {
	AccessToken string
	Account     string
	Method      string
	ExpiresAt   time.Time
}

type Provider interface {
	Name() string
	Load(context.Context) (Session, error)
	Login(context.Context) (Session, error)
	Logout(context.Context) error
}

type OAuthProvider struct {
	AuthorizeURL  string
	BrowserOpener func(string) error
}

type KeyringProvider struct {
	Service string
	Session Session
}

type SSHDetector interface {
	Remote() bool
}

type EnvSSHDetector struct{}

func (p OAuthProvider) Name() string { return "oauth" }

func (p OAuthProvider) Load(context.Context) (Session, error) {
	return Session{}, ErrNoSession
}

func (p OAuthProvider) Login(context.Context) (Session, error) {
	if p.BrowserOpener != nil && p.AuthorizeURL != "" {
		_ = p.BrowserOpener(p.AuthorizeURL)
	}
	return Session{
		AccessToken: "oauth-token",
		Account:     "user@example.com",
		Method:      p.Name(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}, nil
}

func (p OAuthProvider) Logout(context.Context) error { return nil }

func (p *KeyringProvider) Name() string { return "keyring" }

func (p *KeyringProvider) Load(context.Context) (Session, error) {
	if p.Session.AccessToken == "" {
		return Session{}, ErrNoSession
	}
	return p.Session, nil
}

func (p *KeyringProvider) Login(context.Context) (Session, error) {
	if p.Session.AccessToken == "" {
		p.Session = Session{
			AccessToken: "keyring-token",
			Account:     "cached@example.com",
			Method:      p.Name(),
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		}
	}
	return p.Session, nil
}

func (p *KeyringProvider) Logout(context.Context) error {
	p.Session = Session{}
	return nil
}

func (EnvSSHDetector) Remote() bool {
	return os.Getenv("SSH_CONNECTION") != "" || os.Getenv("SSH_TTY") != "" || os.Getenv("SSH_CLIENT") != ""
}
