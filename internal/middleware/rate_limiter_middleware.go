package middleware

import (
	"net"
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
)

// RateLimiterMiddleware recebe PolicyUsecase e RateLimiterInstance.
// O Order é: PolicyUsecase resolve strategy -> strategy gera key+rules -> envia msg ao RateLimiter.
func RateLimiterMiddleware(policy *policy_usecase.PolicyUsecase, rl *ratelimiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// monta input
			var input policy_usecase.InputPolicyDTO
			if api := r.Header.Get("API-KEY"); api != "" {
				input.Tolken = api
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			input.IP = ip

			// resolve strategy
			strategy, key := policy.Resolver(input)

			// generate key (token strategy may validate token)
			key, err := strategy.GenerateKey(r.Context(), key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			reply := make(chan error, 1)
			msg := ratelimiter.RateLimitMessage{
				Ctx:       r.Context(),
				Key:       key,
				Limit:     strategy.GetLimit(),
				TTL:       strategy.GetTTL(),
				ReplyChan: reply,
			}

			// send to rate limiter
			select {
			case rl.InputChan <- msg:
				// wait reply
				if err := <-reply; err != nil {
					http.Error(w, err.Error(), http.StatusTooManyRequests)
					return
				}
			case <-r.Context().Done():
				http.Error(w, "request canceled", http.StatusRequestTimeout)
				return
			}

			// passed rate limiter -> forward
			next.ServeHTTP(w, r)
		})
	}
}
