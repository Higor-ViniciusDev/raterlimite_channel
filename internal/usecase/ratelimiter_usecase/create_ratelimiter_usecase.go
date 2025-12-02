package ratelimiter_usecase

import (
	"context"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/request_info_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/expire_usecase"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/strategy_usecase"
)

type PolicyUsecase struct {
	TokenStrategy RateLimitStrategyInterface
	IPStrategy    RateLimitStrategyInterface
}

func NewPolicyUsecase(expire expire_usecase.ExpirerInterface, tolkenRepository tolken_entity.TolkenRepositoryInterface, requestInfo request_info_entity.RequestRepository) *PolicyUsecase {
	return &PolicyUsecase{
		IPStrategy:    strategy_usecase.NewIPStrategyUsecase(expire, requestInfo),
		TokenStrategy: strategy_usecase.NewTolkenStrategyUsecase(expire, tolkenRepository, requestInfo),
	}
}

// Interface usecase para fazer inversão de dependencias
type PolicyUsecaseInterface interface {
	Resolver(ctx context.Context, input InputPolicyDTO) (RateLimitStrategyInterface, string)
}

type RateLimitStrategyInterface interface {
	Validate(ctx context.Context, key string) *internal_error.InternalError
}

type InputPolicyDTO struct {
	IP     string
	Tolken string
}

// Metodo resolver que vai decidir qual regrar usar IP ou Tolken regra
func (p *PolicyUsecase) Resolver(ctx context.Context, input InputPolicyDTO) (RateLimitStrategyInterface, string) {

	if input.Tolken != "" {
		return p.TokenStrategy, input.Tolken
	}

	return p.IPStrategy, input.IP
}
