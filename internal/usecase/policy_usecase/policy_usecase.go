package policy_usecase

import (
	"context"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/strategy_usecase"
)

type PolicyUsecase struct {
	TokenStrategy RateLimitStrategy
	IPStrategy    RateLimitStrategy
}

type PolicyUsecaseInterface interface {
	Resolver(input InputPolicyDTO) (RateLimitStrategy, string)
}

type RateLimitStrategy interface {
	GenerateKey(ctx context.Context, key string) (string, *internal_error.InternalError)
	GetLimit() int64
	GetTTL() time.Duration
	GetPenaltyDuration() time.Duration
	SaveRequestInfo(ctx context.Context, key string) *internal_error.InternalError
	GetInfoType() string
}

type InputPolicyDTO struct {
	IP     string
	Tolken string
}

func NewPolicyUsecase(ip *strategy_usecase.IPStrategyUsecase, tok *strategy_usecase.TokenStrategyUsecase) *PolicyUsecase {
	return &PolicyUsecase{
		IPStrategy:    ip,
		TokenStrategy: tok,
	}
}

// Decide qual strategy usar
func (p *PolicyUsecase) Resolver(input InputPolicyDTO) (RateLimitStrategy, string) {
	if input.Tolken != "" {
		return p.TokenStrategy, input.Tolken
	}
	return p.IPStrategy, input.IP
}
