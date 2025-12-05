package strategy_usecase

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/request_info_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
)

type TokenStrategyUsecase struct {
	TokenRepository   tolken_entity.TolkenRepositoryInterface
	RequestRepository request_info_entity.RequestRepository

	limitTok      int64
	window        time.Duration
	PenalityBlock time.Duration
}

func NewTokenStrategyUsecase(tokenRepo tolken_entity.TolkenRepositoryInterface, requestRepository request_info_entity.RequestRepository) *TokenStrategyUsecase {
	limitStr := os.Getenv("REQUEST_PER_SECOND_TOLKEN")
	ttlStr := os.Getenv("TLL_KEY_TOLKEN")
	penaltyStr := os.Getenv("TIME_UNLOCKED_NEW_REQUEST_TOLKEN")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	ttl, _ := strconv.Atoi(ttlStr)
	penalty, _ := strconv.Atoi(penaltyStr)

	return &TokenStrategyUsecase{
		TokenRepository:   tokenRepo,
		RequestRepository: requestRepository,
		limitTok:          limit,
		window:            time.Duration(ttl) * time.Second,
		PenalityBlock:     time.Duration(penalty) * time.Second,
	}
}

func (s *TokenStrategyUsecase) GenerateKey(ctx context.Context, key string) (string, *internal_error.InternalError) {

	if key == "" {
		return "", internal_error.NewBadRequestError("tolken invalid")
	}

	// valida se token existe no redis (sem regras complexas)
	if !s.TokenRepository.ValidateTolken(ctx, key) {
		return "", internal_error.NewBadRequestError("tolken not found")
	}

	return "token:" + key, nil
}

func (s *TokenStrategyUsecase) GetLimit() int64 {
	return s.limitTok
}

func (s *TokenStrategyUsecase) GetTTL() time.Duration {
	return s.window
}

func (s *TokenStrategyUsecase) GetPenaltyDuration() time.Duration {
	return s.PenalityBlock
}

func (s *TokenStrategyUsecase) GetInfoType() string {
	return "TOLKEN"
}

func (s *TokenStrategyUsecase) SaveRequestInfo(ctx context.Context, key string) *internal_error.InternalError {
	newRequest := request_info_entity.NewRequestInfo(key, request_info_entity.Active, s.GetInfoType())
	return s.RequestRepository.Save(ctx, key, newRequest.Status, newRequest.FontePolicy)
}
