package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
)

type RateLimitMessage struct {
	Ctx       context.Context
	Key       string
	Limit     int64
	TTL       time.Duration
	Strategy  interface{}
	ReplyChan chan error
}

// Counter guarda o contador em memória.
type Counter struct {
	Count int64
}

type RateLimiter struct {
	InputChan chan RateLimitMessage
	workers   int

	mu       sync.Mutex
	requests map[string]*Counter
	closed   chan struct{}
}

func NewRateLimiter(workers int, queueSize int) *RateLimiter {
	if workers <= 0 {
		workers = 2
	}

	if queueSize <= 0 {
		queueSize = 1000
	}

	rl := &RateLimiter{
		workers:   workers,
		InputChan: make(chan RateLimitMessage, queueSize),
		requests:  make(map[string]*Counter),
		closed:    make(chan struct{}),
	}

	for range rl.workers {
		go rl.worker()
	}

	return rl
}

// Stop fecha o rate limiter (fecha o canal de entrada).
func (rl *RateLimiter) Stop() {
	logger.Info("Canal encerrando")
	close(rl.closed)
	close(rl.InputChan)
}

func (rl *RateLimiter) worker() {
	for msg := range rl.InputChan {
		// se o rate limiter estiver sendo fechado, tenta sair
		select {
		case <-rl.closed:
			// devolve erro para quem enviou, se aplicável
			select {
			case msg.ReplyChan <- errors.New("rate limiter shutting down"):
			default:
			}
			return
		default:
		}

		rl.mu.Lock()

		counter, exists := rl.requests[msg.Key]
		if !exists {
			counter = &Counter{
				Count: 0,
			}
			rl.requests[msg.Key] = counter

			// TTL: ao criar o contador, agenda limpeza após msg.TTL
			go func(key string, ttl time.Duration) {
				// timer local para limpeza
				timer := time.NewTimer(ttl)
				select {
				case <-timer.C:
					rl.mu.Lock()
					delete(rl.requests, key)
					rl.mu.Unlock()
				case <-rl.closed:
					timer.Stop()
				}
			}(msg.Key, msg.TTL)
		}

		counter.Count++
		// compara com o limit enviado na mensagem
		if counter.Count > msg.Limit {
			rl.mu.Unlock()
			// excedeu
			select {
			case msg.ReplyChan <- errors.New("rate limit exceeded"):
			default:
			}
			continue
		}

		rl.mu.Unlock()

		// Tudo ok
		select {
		case msg.ReplyChan <- nil:
		default:
		}
	}
}
