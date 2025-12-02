package strategy_usecase

import (
	"context"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/request_info_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/expire_usecase"
)

type TolkenStrategyUsecase struct {
	Expirer          expire_usecase.ExpirerInterface
	TolkenRepository tolken_entity.TolkenRepositoryInterface
	RequestInfo      request_info_entity.RequestRepository
}

func NewTolkenStrategyUsecase(expire expire_usecase.ExpirerInterface, tolkenRepository tolken_entity.TolkenRepositoryInterface, requestInfo request_info_entity.RequestRepository) *TolkenStrategyUsecase {
	return &TolkenStrategyUsecase{
		Expirer:          expire,
		TolkenRepository: tolkenRepository,
		RequestInfo:      requestInfo,
	}
}

func (ts *TolkenStrategyUsecase) Validate(ctx context.Context, key string) *internal_error.InternalError {
	// Verifica se o token é válido
	isValid := ts.TolkenRepository.ValidateTolken(ctx, key)

	if !isValid {
		return internal_error.NewInternalServerError("Invalid or expired token")
	}

	// Verifica se o token foi bloqueado pelo rate limiter
	blocked := ts.Expirer.IsExpired(key)

	if blocked {
		return internal_error.NewInternalServerError("Token is rate limited")
	}

	// Salva a informação da requisição
	if err := ts.RequestInfo.Save(ctx, key, request_info_entity.Active, request_info_entity.FONTE_TOLKEN); err != nil {
		return err
	}

	return nil
}
