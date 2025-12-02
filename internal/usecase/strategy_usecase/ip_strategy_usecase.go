package strategy_usecase

import (
	"context"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/request_info_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/expire_usecase"
)

type IPStrategyUsecase struct {
	Expirer     expire_usecase.ExpirerInterface
	RequestInfo request_info_entity.RequestRepository
}

func NewIPStrategyUsecase(expire expire_usecase.ExpirerInterface, requestRepo request_info_entity.RequestRepository) *IPStrategyUsecase {
	return &IPStrategyUsecase{
		Expirer:     expire,
		RequestInfo: requestRepo,
	}
}

// Validate implementa a interface Strategy para validação por IP
func (ts *IPStrategyUsecase) Validate(ctx context.Context, key string) *internal_error.InternalError {
	// Verifica se o IP foi bloqueado
	blocked := ts.Expirer.IsExpired(key)

	if blocked {
		return internal_error.NewInternalServerError("IP address is rate limited")
	}

	// Salva a informação da requisição
	if err := ts.RequestInfo.Save(ctx, key, request_info_entity.Active, request_info_entity.FONTE_IP); err != nil {
		return err
	}

	return nil
}
