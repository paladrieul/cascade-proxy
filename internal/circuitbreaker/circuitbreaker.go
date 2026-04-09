package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in the open state.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failing; reject requests
	StateHalfOpen              // probe if upstream recovered
)

// Config holds circuit breaker tuning parameters.
type Config struct {
	FailureThreshold int
	SuccessThreshold int
	Timeout          time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          10 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu             sync.Mutex
	cfg            Config
	state          State
	failureCount   int
	successCount   int
	lastFailureTime time.Time
}

// New creates a CircuitBreaker with the given Config.
func New(cfg Config) *CircuitBreaker {
	return &CircuitBreaker{cfg: cfg, state: StateClosed}
}

// State returns the current state (thread-safe).
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.currentState()
}

// currentState checks whether the open timeout has elapsed and transitions
// to half-open if so. Must be called with cb.mu held.
func (cb *CircuitBreaker) currentState() State {
	if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.cfg.Timeout {
		cb.state = StateHalfOpen
		cb.successCount = 0
	}
	return cb.state
}

// Execute runs fn if the circuit allows it, recording success or failure.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	state := cb.currentState()
	if state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	return err
}

// onFailure records a failure and opens the circuit if the threshold is reached.
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()
	if cb.state == StateHalfOpen || cb.failureCount >= cb.cfg.FailureThreshold {
		cb.state = StateOpen
		cb.successCount = 0
	}
}

// onSuccess records a success and closes the circuit when threshold is met.
func (cb *CircuitBreaker) onSuccess() {
	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.cfg.SuccessThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	} else {
		cb.failureCount = 0
	}
}
