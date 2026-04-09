package circuitbreaker

import (
	"testing"
	"time"
)

func defaultConfig() Config {
	return Config{
		FailureThreshold: 3,
		SuccessThreshold: 2,
		Timeout:          100 * time.Millisecond,
	}
}

func TestInitialStateClosed(t *testing.T) {
	cb := New(defaultConfig())
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed, got %v", cb.CurrentState())
	}
}

func TestOpensAfterFailureThreshold(t *testing.T) {
	cb := New(defaultConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after %d failures", 3)
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestTransitionsToHalfOpenAfterTimeout(t *testing.T) {
	cb := New(defaultConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(150 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Fatalf("expected nil after timeout, got %v", err)
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Fatalf("expected StateHalfOpen, got %v", cb.CurrentState())
	}
}

func TestClosesAfterSuccessThresholdInHalfOpen(t *testing.T) {
	cb := New(defaultConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(150 * time.Millisecond)
	_ = cb.Allow() // transition to half-open
	cb.RecordSuccess()
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed after successes in half-open, got %v", cb.CurrentState())
	}
}

func TestReopensOnFailureInHalfOpen(t *testing.T) {
	cb := New(defaultConfig())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	time.Sleep(150 * time.Millisecond)
	_ = cb.Allow()
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected StateOpen after failure in half-open, got %v", cb.CurrentState())
	}
}

func TestSuccessResetFailureCount(t *testing.T) {
	cb := New(defaultConfig())
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected StateClosed after success reset, got %v", cb.CurrentState())
	}
	if cb.failureCount != 0 {
		t.Fatalf("expected failureCount=0, got %d", cb.failureCount)
	}
}
