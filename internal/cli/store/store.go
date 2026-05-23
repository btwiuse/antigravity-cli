package store

import (
	"sort"
	"sync"

	"github.com/anthropics/antigravity-cli/internal/cli/types"
)

type MemoryStore struct {
	mu        sync.RWMutex
	sessions  map[string]types.Session
	artifacts map[string]types.Artifact
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions:  make(map[string]types.Session),
		artifacts: make(map[string]types.Artifact),
	}
}

func (s *MemoryStore) SaveSession(session types.Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[session.ID] = session
}

func (s *MemoryStore) Session(id string) (types.Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

func (s *MemoryStore) ListSessions() []types.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]types.Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		out = append(out, session)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *MemoryStore) SaveArtifact(artifact types.Artifact) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.artifacts[artifact.ID] = artifact
}

func (s *MemoryStore) ListArtifacts() []types.Artifact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]types.Artifact, 0, len(s.artifacts))
	for _, artifact := range s.artifacts {
		out = append(out, artifact)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
