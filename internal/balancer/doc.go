// Package balancer implements a round-robin load balancer for distributing
// upstream HTTP requests across multiple backend targets.
//
// # Usage
//
// Create a Balancer with a list of target URLs:
//
//	b, err := balancer.New(balancer.Config{
//		Targets: []string{
//			"http://backend1:8080",
//			"http://backend2:8080",
//		},
//	})
//
// Call Next() on each incoming request to obtain the next target:
//
//	target, err := b.Next()
//
// Next() is safe for concurrent use by multiple goroutines.
package balancer
