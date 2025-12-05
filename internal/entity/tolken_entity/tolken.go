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
	JWTSecret    string
	TokenAuth    *jwtauth.JWTAuth
	TimeLife     int
	TolkenString string
}

func NewTolken() *Tolken {
	secret := os.Getenv("JWT_SECRET")
	tll := os.Getenv("TOLKEN_EXPIRATION")
	tllNumber, _ := strconv.Atoi(tll)

	tolkenEntity := &Tolken{
		JWTSecret: secret,
		TokenAuth: jwtauth.New("HS256", []byte(secret), nil),
		TimeLife:  tllNumber,
	}
	tolkenEntity.SetTolkenString()
	return tolkenEntity
}

func (t *Tolken) GetTolkenString() string {
	return t.TolkenString
}

func (t *Tolken) SetTolkenString() {
	// Numero aleatorio para o tolken sempre ser diferente
	// Não tem uma key expire nele pois a validade do tolken é controlada pela aplicação
	randomID := rand.Int63()
	_, retorno, err := t.TokenAuth.Encode(map[string]interface{}{
		"randomID": randomID,
	})

	if err != nil {
		logger.Error("error getstring tolken ", err)
		return
	}

	t.TolkenString = retorno
}

type TolkenRepositoryInterface interface {
	Save(ctx context.Context, tolken *Tolken) *internal_error.InternalError
	ValidateTolken(ctx context.Context, tolkenID string) bool
	DeleteInfoByTolken(ctx context.Context, tolkenID string) *internal_error.InternalError
}
