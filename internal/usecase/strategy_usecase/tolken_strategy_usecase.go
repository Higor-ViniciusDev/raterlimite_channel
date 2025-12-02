package strategy_usecase

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
)

type TokenStrategyUsecase struct {
	tokenRepo tolken_entity.TolkenRepositoryInterface

	limitTok int64
	window   time.Duration
}

func NewTokenStrategyUsecase(tokenRepo tolken_entity.TolkenRepositoryInterface) *TokenStrategyUsecase {
	limitStr := os.Getenv("REQUEST_PER_SECOND_TOLKEN")
	ttlStr := os.Getenv("TIME_UNLOCKED_NEW_REQUEST_TOLKEN")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	ttl, _ := strconv.Atoi(ttlStr)

	return &TokenStrategyUsecase{
		tokenRepo: tokenRepo,
		limitTok:  limit,
		window:    time.Duration(ttl) * time.Second,
	}
}

func (s *TokenStrategyUsecase) GenerateKey(ctx context.Context, key string) (string, error) {

	if key == "" {
		return "", errors.New("token not provided")
	}

	// valida se token existe no redis (sem regras complexas)
	if !s.tokenRepo.ValidateTolken(ctx, key) {
		return "", errors.New("invalid or expired token")
	}

	return "token:" + key, nil
}

func (s *TokenStrategyUsecase) Limit() int64 {
	return s.limitTok
}

func (s *TokenStrategyUsecase) Window() time.Duration {
	return s.window
}
