package ratelimiter

import (
	"errors"
	"sync"
	"time"
)

type RateLimitMessage struct {
	Key       string
	ReplyChan chan error
}

type Counter struct {
	Count int64
}

type RateLimiter struct {
	InputChan chan RateLimitMessage
	TTL       time.Duration
	Limit     int64
	workers   int

	mu       sync.Mutex
	requests map[string]*Counter
}

func NewRateLimiter(workers int, ttl time.Duration, limit int64) *RateLimiter {
	rl := &RateLimiter{
		workers:   workers,
		TTL:       ttl,
		Limit:     limit,
		requests:  make(map[string]*Counter),
		InputChan: make(chan RateLimitMessage, 10000),
	}

	for range workers {
		go rl.worker()
	}

	return rl
}

func (rl *RateLimiter) worker() {
	for msg := range rl.InputChan {
		rl.mu.Lock()

		counter, exists := rl.requests[msg.Key]

		if !exists {
			counter = &Counter{Count: 0}
			rl.requests[msg.Key] = counter

			// TTL automático limpa o contador
			go func(key string) {
				<-time.After(rl.TTL)
				rl.mu.Lock()
				delete(rl.requests, key)
				rl.mu.Unlock()
			}(msg.Key)
		}

		counter.Count++

		if counter.Count > rl.Limit {
			rl.mu.Unlock()
			msg.ReplyChan <- errors.New("rate limit exceeded")
			continue
		}

		rl.mu.Unlock()

		msg.ReplyChan <- nil
	}
}

func (rl *RateLimiter) Allow(key string) error {
	reply := make(chan error)
	rl.InputChan <- RateLimitMessage{
		Key:       key,
		ReplyChan: reply,
	}
	return <-reply
}
