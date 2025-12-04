package tolken_entity

import (
	"context"
	"math/rand"
	"os"
	"strconv"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/go-chi/jwtauth"
)

type Tolken struct {
	JWTSecret string
	TokenAuth *jwtauth.JWTAuth
	TimeLife  int64
}

func NewTolken() *Tolken {
	secret := os.Getenv("JWT_SECRET")
	tll := os.Getenv("TOLKEN_EXPIRATION")
	tllNumber, _ := strconv.Atoi(tll)

	return &Tolken{
		JWTSecret: secret,
		TokenAuth: jwtauth.New("HS256", []byte(secret), nil),
		TimeLife:  int64(tllNumber),
	}
}

func (t *Tolken) GetTolkenString() (string, *internal_error.InternalError) {
	// Numero aleatorio para o tolken sempre ser diferente
	// Não tem uma key expire nele pois a validade do tolken é controlada pela aplicação
	randomID := rand.Int63()
	_, retorno, err := t.TokenAuth.Encode(map[string]interface{}{
		"randomID": randomID,
	})

	if err != nil {
		logger.Error("error getstring tolken ", err)
		return "", internal_error.NewInternalServerError("Error get tolken string")
	}

	return retorno, nil
}

type TolkenRepositoryInterface interface {
	Save(ctx context.Context, tolken *Tolken) *internal_error.InternalError
	ValidateTolken(ctx context.Context, tolkenID string) bool
	DeleteInfoByTolken(ctx context.Context, tolkenID string) *internal_error.InternalError
}
