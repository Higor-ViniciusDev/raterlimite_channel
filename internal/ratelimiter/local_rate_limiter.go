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
	KeyCreatedAt time.Time
}

type RateLimiter struct {
	InputChan chan RateLimitMessage
	workers   int

	mu       sync.Mutex
	requests map[string]*Counter
	closed   chan struct{}
}

func NewRateLimiter(workers int, queueSize int) *RateLimiter {
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
		now := time.Now()

		// Verifica se existe chafe e se está bloqueado com base no blockedUntil e no tempo atual
		if exists && !counter.BlockedUntil.IsZero() && now.Before(counter.BlockedUntil) {
			rl.mu.Unlock()

			// Avisar que o limite foi excedido
			logger.Info(fmt.Sprintf("Rate limit exceeded (blocked) for key: %s", msg.Key))
			select {
			case msg.ReplyChan <- internal_error.NewManyRequestError("you have reached the maximum number of requests or actions allowed within a certain time frame"):
			default:
			}
			continue
		}

		if !exists {
			counter = &Counter{
				Count: 0,
			}
			rl.requests[msg.Key] = counter

			go func(key string, ttl time.Duration, penality time.Duration) {
				maxTime := ttl
				if penality > maxTime {
					maxTime = penality
				}
				maxTime += 1 * time.Second // Buffer seguranca caso a key ainda esteja em uso após um segundo que foi definido

				timer := time.NewTimer(maxTime)
				select {
				case <-timer.C:
					rl.mu.Lock()
					if c, ok := rl.requests[key]; ok {
						now := time.Now()
						// Verifica se o contador ainda está expirado antes de deletar
						penalityExpired := c.BlockedUntil.IsZero() || now.After(c.BlockedUntil)
						tllExpired := now.Sub(c.KeyCreatedAt) > ttl

						if penalityExpired && tllExpired {
							delete(rl.requests, key)
						}
					}
					rl.mu.Unlock()
				case <-rl.closed:
					timer.Stop()
				}
			}(msg.Key, msg.TTL, msg.Strategy.GetPenaltyDuration())
		} else {
			// Se o contador existir, mas o bloqueio expirou, resetar o contador
			penalityExpired := counter.BlockedUntil.IsZero() || now.After(counter.BlockedUntil)
			tllExpired := now.Sub(counter.KeyCreatedAt) > msg.TTL

			if penalityExpired && tllExpired {
				// Se ambos expiraram, resetar o contador e o bloqueio
				counter.Count = 0
				counter.BlockedUntil = time.Time{}
				counter.KeyCreatedAt = now
			} else if penalityExpired && !tllExpired {
				// Se o bloqueio expirou, mas o TTL não, apenas resetar o bloqueio
				counter.BlockedUntil = time.Time{}
			}
		}

		counter.Count++
		// compara com o limit enviado na mensagem
		if counter.Count >= msg.Limit {
			//Se bloquear request, adicionar penalidade de timer para o bloqueio temporario
			if counter.BlockedUntil.IsZero() {
				counter.BlockedUntil = now.Add(msg.Strategy.GetPenaltyDuration())
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
		//Para fins de info, apenas salvar os que deram sucesso é requisito do projeto
		if err := msg.Strategy.SaveRequestInfo(context.Background(), msg.Key); err != nil {
			logger.Error("erro ao salvar request info", err)
		}
	}
}
