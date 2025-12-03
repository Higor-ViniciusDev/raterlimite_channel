package strategy_usecase

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
)

type IPStrategyUsecase struct {
	limitIP int64
	window  time.Duration
}

func NewIPStrategyUsecase() *IPStrategyUsecase {

	limitStr := os.Getenv("REQUEST_PER_SECOND_IP")
	ttlStr := os.Getenv("TIME_UNLOCKED_NEW_REQUEST_IP")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	ttl, _ := strconv.Atoi(ttlStr)

	return &IPStrategyUsecase{
		limitIP: limit,
		window:  time.Duration(ttl) * time.Second,
	}
}

func (s *IPStrategyUsecase) GenerateKey(ctx context.Context, key string) (string, error) {

	if key == "" {
		return "", internal_error.NewBadRequestError("key invalid")
	}

	return "ip:" + key, nil
}

func (s *IPStrategyUsecase) GetLimit() int64 {
	return s.limitIP
}

func (s *IPStrategyUsecase) GetTTL() time.Duration {
	return s.window
}
