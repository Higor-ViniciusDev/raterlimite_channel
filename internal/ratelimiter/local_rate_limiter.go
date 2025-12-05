package ratelimiter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
)

type RateLimitMessage struct {
	Ctx       context.Context
	Key       string
	Limit     int64
	TTL       time.Duration
	Strategy  policy_usecase.RateLimitStrategy
	ReplyChan chan *internal_error.InternalError
}

// Counter guarda o contador em memória.
type Counter struct {
	Count        int64
	BlockedUntil time.Time
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

	// Saída de dados parados em buffer não processados
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for key := range rl.requests {
		delete(rl.requests, key)
	}

	for i := 0; i < rl.workers; i++ {
		rl.InputChan <- RateLimitMessage{}
	}

	close(rl.closed)
	close(rl.InputChan)
}

func (rl *RateLimiter) worker() {
	for msg := range rl.InputChan {
		// se o rate limiter estiver sendo fechado, tenta sair
		select {
		case <-rl.closed:
			select {
			case msg.ReplyChan <- internal_error.NewInternalServerError("server shutdown"):
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

			go func(key string, ttl time.Duration) {
				// timer local para limpeza ttl key
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

		// Verifica se está bloqueado
		if counter.BlockedUntil.After(time.Now()) {
			rl.mu.Unlock()

			// Avisar que o limite foi excedido
			logger.Info(fmt.Sprintf("Rate limit exceeded (blocked) for key: %s", msg.Key))
			select {
			case msg.ReplyChan <- internal_error.NewManyRequestError("you have reached the maximum number of requests or actions allowed within a certain time frame"):
			default:
			}
			continue
		}

		counter.Count++
		// compara com o limit enviado na mensagem
		if counter.Count > msg.Limit {
			//Se bloquear request, adicionar penalidade de timer para o bloqueio temporario
			if counter.BlockedUntil.Before(time.Now()) {
				counter.BlockedUntil = time.Now().Add(1 * time.Minute) // bloqueia por 1 minuto
			}

			logger.Info(fmt.Sprintf("Rate limit exceeded for key: %s", msg.Key))
			rl.mu.Unlock()
			// excedeu
			select {
			case msg.ReplyChan <- internal_error.NewManyRequestError("you have reached the maximum number of requests or actions allowed within a certain time frame"):
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
		if err := msg.Strategy.SaveRequestInfo(context.Background(), msg.Key); err != nil {
			logger.Error("erro ao salvar request info", err)
		}
	}
}
