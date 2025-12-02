package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
)

// Strategy interface define o contrato para validações de rate limit
type Strategy interface {
	Validate(ctx context.Context, key string) *internal_error.InternalError
}

type RateLimitMessage struct {
	Key       string
	ReplyChan chan error
	Ctx       context.Context
}

type Counter struct {
	Count int64
}

type RateLimiter struct {
	InputChan chan RateLimitMessage
	TTL       time.Duration
	Limit     int64
	workers   int
	Strategy  Strategy

	mu       sync.Mutex
	requests map[string]*Counter
}

func NewRateLimiter(workers int, ttl time.Duration, limit int64, strategy Strategy) *RateLimiter {
	rl := &RateLimiter{
		workers:   workers,
		TTL:       ttl,
		Limit:     limit,
		Strategy:  strategy,
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

			// TTL automático
			go func(key string) {
				t := time.After(rl.TTL)
				<-t
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

		// ---- Executa a validação adicional via Strategy ----
		if err := rl.Strategy.Validate(msg.Ctx, msg.Key); err != nil {
			msg.ReplyChan <- errors.New(err.Message)
			continue
		}

		msg.ReplyChan <- nil
	}
}
