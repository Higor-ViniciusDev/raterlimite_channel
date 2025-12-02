package request_info_entity

import (
	"context"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
)

type RequestInfo struct {
	Key          string
	TimeRequered int64
	Status       RequestCondition
	FontePolicy  string
}

type RequestCondition int64

const (
	Active RequestCondition = iota
	Bloqued
)

const (
	FONTE_IP     = "IP"
	FONTE_TOLKEN = "TOLKEN"
)

type RequestRepository interface {
	Save(ctx context.Context, key string, status RequestCondition, tipoRequest string) *internal_error.InternalError
}
