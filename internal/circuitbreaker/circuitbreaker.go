package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

// Config holds configuration for the CircuitBreaker.
type Config struct {
	FailureThreshold uint          // number of failures before opening
	SuccessThreshold uint          // successes in half-open before closing
	Timeout          time.Duration // how long to stay open before half-open
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu             sync.Mutex
	cfg            Config
	state          State
	failureCount   uint
	successCount   uint
	lastFailureTime time.Time
}

// New creates a new CircuitBreaker with the given config.
func New(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{cfg: cfg, state: StateClosed}
}

// Allow reports whether a request should be allowed through.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailureTime) >= cb.cfg.Timeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			return nil
		}
		return ErrCircuitOpen
	default:
		return nil
	}
}

// RecordSuccess records a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.cfg.SuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
		return
	}
	cb.failureCount = 0
}

// RecordFailure records a failed call.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()
	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		return
	}
	cb.failureCount++
	if cb.failureCount >= cb.cfg.FailureThreshold {
		cb.state = StateOpen
	}
}

// State returns the current circuit breaker state.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
