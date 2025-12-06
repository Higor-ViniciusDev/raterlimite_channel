package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/rest_err"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/ratelimiter"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/policy_usecase"
)

// RateLimiterMiddleware recebe PolicyUsecase e RateLimiterInstance.
// O Order é: PolicyUsecase resolve strategy -> strategy gera key+rules -> envia msg ao RateLimiter.
func RateLimiterMiddleware(policy *policy_usecase.PolicyUsecase, rl *ratelimiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
			var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
			var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

			// monta input
			var input policy_usecase.InputPolicyDTO
			if api := r.Header.Get("API-KEY"); api != "" {
				input.Tolken = api
			}

			//Exemplo tirado do middleware https://github.com/go-chi/chi/blob/master/middleware/realip.go
			var ip string

			if tcip := r.Header.Get(trueClientIP); tcip != "" {
				ip = tcip
			} else if xrip := r.Header.Get(xRealIP); xrip != "" {
				ip = xrip
			} else if xff := r.Header.Get(xForwardedFor); xff != "" {
				ip, _, _ = strings.Cut(xff, ",")
			} else if host := r.RemoteAddr; host != "" {
				ip = host
			}

			if strings.Contains(ip, ":") {
				host, _, err := net.SplitHostPort(ip)
				if err == nil {
					ip = host
				}
			}

			if ip == "" || net.ParseIP(ip) == nil {
				input.IP = ""
			} else {
				input.IP = net.ParseIP(ip).String()
			}
			// resolve strategy
			strategy, key := policy.Resolver(input)

			// generate key (token strategy may validate token)
			key, err := strategy.GenerateKey(r.Context(), key)
			if err != nil {
				restError := rest_err.ConvertInternalErrorToRestError(err)
				http.Error(w, restError.Message, restError.Code)
				return
			}

			reply := make(chan *internal_error.InternalError, 1)
			msg := ratelimiter.RateLimitMessage{
				Ctx:       r.Context(),
				Key:       key,
				Limit:     strategy.GetLimit(),
				TTL:       strategy.GetTTL(),
				ReplyChan: reply,
				Strategy:  strategy,
			}

			// send to rate limiter
			select {
			case rl.InputChan <- msg:
				// wait reply
				if err := <-reply; err != nil {
					restError := rest_err.ConvertInternalErrorToRestError(err)
					http.Error(w, restError.Message, restError.Code)
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
