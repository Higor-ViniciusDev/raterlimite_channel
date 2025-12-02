package repository

import (
	"context"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/request_info_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/redis/go-redis/v9"
)

type RequestInfoRepository struct {
	RedisCLI *redis.Client
}

func NewRequestInfoRepository(redisCli *redis.Client) *RequestInfoRepository {
	return &RequestInfoRepository{
		RedisCLI: redisCli,
	}
}

func (rp *RequestInfoRepository) Save(ctx context.Context, key string, status request_info_entity.RequestCondition, tipoRequest string) *internal_error.InternalError {
	data := map[string]interface{}{
		"FontePolicy":  tipoRequest,
		"Status":       status,
		"TimeRequered": time.Now().Unix(),
	}

	cmd := rp.RedisCLI.HSet(ctx, key, data)
	if cmd.Err() != nil {
		logger.Error("error save info request", cmd.Err())
		return internal_error.NewInternalServerError("error save tolken redis")
	}

	return nil
}
