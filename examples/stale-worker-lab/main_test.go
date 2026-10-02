package main

import (
	"errors"
	"sync"
	"testing"
)

func TestVersionRefreshCanReplayStaleIntent(t *testing.T) {
	value, first, retry := demonstrate(false)
	if !errors.Is(first, ErrConflict) || retry != nil || value != "A: obsolete desired configuration" {
		t.Fatalf("value=%q first=%v retry=%v", value, first, retry)
	}
}

func TestFenceRejectsStaleIntentEvenAfterVersionRefresh(t *testing.T) {
	value, first, retry := demonstrate(true)
	if !errors.Is(first, ErrStale) || !errors.Is(retry, ErrStale) || value != "B: current desired configuration" {
		t.Fatalf("value=%q first=%v retry=%v", value, first, retry)
	}
}

func TestFenceRejectsOldOwnerBeforeNewOwnerWrites(t *testing.T) {
	s := &store{}
	_ = s.activate(1)
	a := s.read()
	_ = s.activate(2)
	if err := s.write(a.version, 1, "old owner", true); !errors.Is(err, ErrStale) {
		t.Fatalf("old owner accepted before successor's first write: %v", err)
	}
	if err := s.write(a.version, 2, "new owner", true); err != nil {
		t.Fatal(err)
	}
}

func TestEpochCannotMoveBackwardOrBeInventedByWriter(t *testing.T) {
	s := &store{}
	_ = s.activate(2)
	for _, epoch := range []uint64{0, 1, 2} {
		if !errors.Is(s.activate(epoch), ErrStale) {
			t.Fatalf("activated epoch %d", epoch)
		}
	}
	for _, epoch := range []uint64{0, 1, 3} {
		if !errors.Is(s.write(0, epoch, "invalid", true), ErrStale) {
			t.Fatalf("accepted epoch %d", epoch)
		}
	}
}

func TestConcurrentCurrentOwnerWritesStillNeedVersionCheck(t *testing.T) {
	s := &store{}
	_ = s.activate(1)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); <-start; results <- s.write(0, 1, "same version", true) }()
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
}
