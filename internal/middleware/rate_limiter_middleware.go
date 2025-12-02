package middleware

import (
	"net"
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
)

func RateLimiterMiddleware(
	policy *policy_usecase.PolicyUsecase,
	rl *ratelimiter.RateLimiter,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var input policy_usecase.InputPolicyDTO

			if apiKey := r.Header.Get("API-KEY"); apiKey != "" {
				input.Tolken = apiKey
			}

			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			input.IP = ip

			strategy, key := policy.Resolver(input)

			key, err := strategy.GenerateKey(r.Context(), key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			reply := make(chan error)
			rl.InputChan <- ratelimiter.RateLimitMessage{
				Key:       key,
				ReplyChan: reply,
			}

			if err := <-reply; err != nil {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
