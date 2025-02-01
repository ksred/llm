package types

import "time"

// MetricsCallbacks defines callback functions for various metrics events
type MetricsCallbacks struct {
	// Request metrics
	OnRequest  func(provider string)                         // Called when a request starts
	OnResponse func(provider string, duration time.Duration) // Called when a request completes successfully
	OnError    func(provider string, err error)              // Called when a request fails
	OnRetry    func(provider string, attempt int, err error) // Called before each retry attempt

	// Pool metrics
	OnPoolGet       func(provider string, waitTime time.Duration) // Called when a connection is retrieved from the pool
	OnPoolRelease   func(provider string)                         // Called when a connection is released back to the pool
	OnPoolExhausted func(provider string)                         // Called when pool is exhausted (all connections in use)
}
