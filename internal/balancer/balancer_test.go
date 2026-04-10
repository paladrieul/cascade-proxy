package balancer

import (
	"testing"
)

func defaultTargets() []string {
	return []string{
		"http://backend1:8080",
		"http://backend2:8080",
		"http://backend3:8080",
	}
}

func TestNewReturnsErrorOnNoTargets(t *testing.T) {
	_, err := New(Config{Targets: []string{}})
	if err == nil {
		t.Fatal("expected error for empty targets, got nil")
	}
}

func TestNewReturnsErrorOnInvalidURL(t *testing.T) {
	_, err := New(Config{Targets: []string{"://invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestNextRoundRobin(t *testing.T) {
	b, err := New(Config{Targets: defaultTargets()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	seen := make(map[string]int)
	for i := 0; i < 9; i++ {
		u, err := b.Next()
		if err != nil {
			t.Fatalf("Next() error: %v", err)
		}
		seen[u.Host]++
	}

	for _, host := range []string{"backend1:8080", "backend2:8080", "backend3:8080"} {
		if seen[host] != 3 {
			t.Errorf("expected host %s to be selected 3 times, got %d", host, seen[host])
		}
	}
}

func TestNextDistributesSingleTarget(t *testing.T) {
	b, err := New(Config{Targets: []string{"http://only:9090"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		u, err := b.Next()
		if err != nil {
			t.Fatalf("Next() error: %v", err)
		}
		if u.Host != "only:9090" {
			t.Errorf("expected only:9090, got %s", u.Host)
		}
	}
}

func TestLenReturnsTargetCount(t *testing.T) {
	b, err := New(Config{Targets: defaultTargets()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Len() != 3 {
		t.Errorf("expected Len() == 3, got %d", b.Len())
	}
}

func TestNextIsConcurrentlySafe(t *testing.T) {
	b, err := New(Config{Targets: defaultTargets()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			_, _ = b.Next()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
