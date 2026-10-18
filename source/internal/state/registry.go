package state

import (
	"errors"
	"sort"
	"sync"

	"example.edu/6cc545/cw1/internal/model"
)

var ErrDuplicate = errors.New("duplicate peer ID")
var ErrMissing = errors.New("peer not found")

type Registry struct {
	mu    sync.RWMutex
	peers map[string]model.Peer
}

func NewRegistry() *Registry { return &Registry{peers: make(map[string]model.Peer)} }
func (s *Registry) Add(p model.Peer) error {
	// Validate the peer before accessing shared state.
	if err := p.Validate(); err != nil {
		return err
	}

	// Acquire exclusive access to the peer registry.
	s.mu.Lock()
	defer s.mu.Unlock()

	// Duplicate detection and insertion must be atomic.
	if _, exists := s.peers[p.ID]; exists {
		return ErrDuplicate
	}

	// Store the validated peer.
	s.peers[p.ID] = p

	return nil
}

func (s *Registry) Snapshot() []model.Peer {
	// Acquire read access to the registry.
	s.mu.RLock()

	// Create an independent slice of peer records.
	peers := make([]model.Peer, 0, len(s.peers))

	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	// Release the lock after copying the data.
	s.mu.RUnlock()

	// Sort the independent copy by peer ID.
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].ID < peers[j].ID
	})

	return peers
}

func (s *Registry) Get(id string) (model.Peer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.peers[id]
	if !ok {
		return p, ErrMissing
	}
	return p, nil
}

var _ = sort.Slice
