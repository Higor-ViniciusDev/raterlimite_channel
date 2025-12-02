package repository

import (
	"context"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/redis/go-redis/v9"
)

type TolkenDB struct {
	RediDB *redis.Client
}

func NewTolkenDB(redisCli *redis.Client) *TolkenDB {
	return &TolkenDB{
		RediDB: redisCli,
	}
}

func (tb *TolkenDB) Save(ctx context.Context, tolken *tolken_entity.Tolken) *internal_error.InternalError {
	tolkenID, _ := tolken.GetTolkenString()
	// Converter struct para map (Redis não aceita struct)
	data := map[string]interface{}{
		"ttl":      tolken.TimeLife,
		"start_at": time.Now().Unix(),
	}

	cmd := tb.RediDB.HSet(ctx, tolkenID, data)
	if cmd.Err() != nil {
		logger.Error("error save tolken redis", cmd.Err())
		return internal_error.NewInternalServerError("error save tolken redis")
	}

	return nil
}

func (tb *TolkenDB) ValidateTolken(ctx context.Context, tolkenID string) bool {
	exists, err := tb.RediDB.Exists(ctx, tolkenID).Result()

	if err != nil {
		logger.Error("error checking block status redis", err)
	}

	if exists > 0 {
		return true
	}

	return false
}

func (tb *TolkenDB) DeleteInfoByTolken(ctx context.Context, tolkenID string) *internal_error.InternalError {
	err := tb.RediDB.Del(ctx, tolkenID).Err()
	if err != nil {
		logger.Error("Error try delete hkey in redis", err)
		return internal_error.NewInternalServerError("Error try delete key")
	}

	return nil
}
