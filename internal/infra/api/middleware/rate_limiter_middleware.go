package middleware

import (
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
)

func RateLimiterMiddleware(rl *ratelimiter.RateLimiter) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// apiKey := r.Header.Get("API-KEY")
			// input.Tolken = apiKey

			// ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			// input.IP = ip

			// policy, key := policyUsecase.Resolver(r.Context(), input)

			// // ---- envia para o RateLimiter local ----
			// reply := make(chan error)
			// rl.InputChan <- ratelimiter.RateLimitMessage{
			// 	Key:       key,
			// 	ReplyChan: reply,
			// }

			// if err := <-reply; err != nil {
			// 	http.Error(w, err.Error(), http.StatusTooManyRequests)
			// 	return
			// }

			// // ---- fallback: valida regras adicionais (TTL via Redis, token inválido, etc) ----
			// if err := policy.Validate(r.Context(), key); err != nil {
			// 	restError := rest_err.ConvertInternalErrorToRestError(err)
			// 	http.Error(w, restError.Message, restError.Code)
			// 	return
			// }

			next.ServeHTTP(w, r)
		})
	}
}
