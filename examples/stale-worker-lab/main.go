// An in-memory teaching model, not a distributed lock implementation.
package main

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrConflict = errors.New("version conflict")
	ErrStale    = errors.New("stale ownership epoch")
)

type snapshot struct {
	version uint64
	value   string
}

type store struct {
	mu      sync.Mutex
	version uint64
	epoch   uint64
	value   string
}

func (s *store) read() snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return snapshot{s.version, s.value}
}

// activate models a trusted ownership transition installed at the write target.
// Workers must not be allowed to mint or activate their own arbitrary epochs.
func (s *store) activate(epoch uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if epoch <= s.epoch {
		return ErrStale
	}
	s.epoch = epoch
	return nil
}

func (s *store) write(version, epoch uint64, value string, fenced bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Fence validation and mutation share the same critical section.
	if fenced && (epoch == 0 || epoch != s.epoch) {
		return ErrStale
	}
	if version != s.version {
		return ErrConflict
	}
	s.value = value
	s.version++
	return nil
}

func demonstrate(fenced bool) (string, error, error) {
	s := &store{value: "initial"}
	_ = s.activate(1)
	a := s.read()     // A captures an old decision, then pauses.
	_ = s.activate(2) // Lease takeover, modeled explicitly without wall-clock sleeps.
	b := s.read()
	if err := s.write(b.version, 2, "B: current desired configuration", fenced); err != nil {
		panic(err)
	}
	first := s.write(a.version, 1, "A: obsolete desired configuration", fenced)
	latest := s.read()
	// Deliberately wrong: refresh only the version, retaining A's old intent.
	retry := s.write(latest.version, 1, "A: obsolete desired configuration", fenced)
	return s.read().value, first, retry
}

func main() {
	for _, fenced := range []bool{false, true} {
		value, first, retry := demonstrate(fenced)
		fmt.Printf("fenced=%t\n  first attempt: %v\n  naive retry: %v\n  final value: %s\n", fenced, first, retry, value)
	}
}
