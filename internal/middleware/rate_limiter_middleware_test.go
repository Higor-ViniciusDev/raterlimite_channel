package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/infra/repository"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/middleware"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/strategy_usecase"
	"github.com/alicebob/miniredis"
	"github.com/redis/go-redis/v9"
)

func Test_RateLimiterMiddleware_IP(t *testing.T) {
	os.Setenv("TIME_UNLOCKED_NEW_REQUEST_IP", "1")
	os.Setenv("REQUEST_PER_SECOND_IP", "7")
	// os.Setenv("", "")
	// os.Setenv("", "")

	mr, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	requestTolkenRepository := repository.NewTolkenDB(redisClient)

	// Estratégias específicas
	ipStr := strategy_usecase.NewIPStrategyUsecase()
	tokStr := strategy_usecase.NewTokenStrategyUsecase(requestTolkenRepository)

	// PolicyUsecase
	policy := &policy_usecase.PolicyUsecase{
		TokenStrategy: tokStr,
		IPStrategy:    ipStr,
	}

	rl := ratelimiter.NewRateLimiter(4, 100)

	// --- Handler de teste ---
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	handler := middleware.RateLimiterMiddleware(policy, rl)(finalHandler)

	// Servidor real de teste
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()

	// --- Faz 20 requisições em 2 segundos (10 por segundo) ---
	// Com limite de 5 requisições por segundo, deveria bloquear ~10 requisições
	numRequests := 20
	duration := 2 * time.Second
	intervalPerRequest := duration / time.Duration(numRequests)

	var mu sync.Mutex
	var blocked int32
	var success int32
	var errors int32

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			// Aguarda intervalo de tempo para distribuir requisições uniformemente
			time.Sleep(time.Duration(index) * intervalPerRequest)

			req, _ := http.NewRequest("GET", server.URL, nil)
			req.RemoteAddr = "10.0.0.1:1234"

			resp, err := client.Do(req)
			if err != nil {
				mu.Lock()
				errors++
				mu.Unlock()
				t.Logf("Requisição %d - Erro: %v", index, err)
				return
			}
			defer resp.Body.Close()

			elapsed := time.Since(start)
			mu.Lock()
			if resp.StatusCode == http.StatusTooManyRequests {
				blocked++
				t.Logf("Requisição %d (%.3fs) - BLOQUEADA (429)", index, elapsed.Seconds())
			} else if resp.StatusCode == http.StatusOK {
				success++
				t.Logf("Requisição %d (%.3fs) - AUTORIZADA (200)", index, elapsed.Seconds())
			}
			mu.Unlock()
		}(i)
	}

	// Aguarda todas as goroutines terminarem
	time.Sleep(duration + 500*time.Millisecond)

	t.Logf("\n=== RESULTADO DO TESTE ===")
	t.Logf("Total de requisições: %d", numRequests)
	t.Logf("Requisições autorizadas: %d", success)
	t.Logf("Requisições bloqueadas: %d", blocked)
	t.Logf("Erros: %d", errors)

	// Validações
	if success == 0 {
		t.Fatalf("FALHA: Nenhuma requisição foi autorizada")
	}

	if blocked == 0 {
		t.Fatalf("FALHA: Nenhuma requisição foi bloqueada (deveria bloquear com ~10 req em 2 segundos, limite de 5/seg)")
	}

	if success+blocked != int32(numRequests) {
		t.Logf("AVISO: Nem todas as requisições foram contabilizadas (sucesso+bloqueadas: %d, total: %d)", success+blocked, numRequests)
	}

	// Verifica proporção esperada: ~50% de sucesso, ~50% de bloqueio
	successRate := float64(success) / float64(numRequests)
	blockRate := float64(blocked) / float64(numRequests)

	t.Logf("Taxa de sucesso: %.1f%%", successRate*100)
	t.Logf("Taxa de bloqueio: %.1f%%", blockRate*100)

	if successRate < 0.2 || successRate > 0.8 {
		t.Logf("AVISO: Taxa de sucesso fora do esperado (20-80%%)")
	}
}
