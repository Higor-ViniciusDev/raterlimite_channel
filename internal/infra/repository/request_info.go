package repository

import (
	"context"
	"fmt"
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
		"FontePolicy": tipoRequest,
		"Status":      int64(status),
		"TimeRequest": time.Now().Unix(),
	}
	//Mudando a key para não repitir a informação
	randomico := time.Now().UnixNano()
	keyFormatada := fmt.Sprintf("%s_%d", key, randomico)
	cmd := rp.RedisCLI.HSet(ctx, keyFormatada, data)
	if cmd.Err() != nil {
		logger.Error("error save info request in repository", cmd.Err())
		return internal_error.NewInternalServerError("error save tolken redis")
	}

	return nil
}
