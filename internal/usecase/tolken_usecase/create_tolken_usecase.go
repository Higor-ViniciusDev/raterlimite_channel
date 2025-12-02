package tolken_usecase

import (
	"context"
	"time"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/entity/tolken_entity"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/expire_usecase"
)

type TolkenUsecase struct {
	TolkenRepository tolken_entity.TolkenRepositoryInterface
	Expirer          expire_usecase.ExpirerInterface
}

type TolkenOutputDTO struct {
	Tolken string `json:"tolken"`
}

func NewTolkenUsecase(tolkenRepository tolken_entity.TolkenRepositoryInterface, expire expire_usecase.ExpirerInterface) *TolkenUsecase {
	return &TolkenUsecase{
		TolkenRepository: tolkenRepository,
		Expirer:          expire,
	}
}

type TolkenUsecaseInterface interface {
	CreateTolken(ctx context.Context) (*TolkenOutputDTO, *internal_error.InternalError)
}

func (tl *TolkenUsecase) CreateTolken(ctx context.Context) (*TolkenOutputDTO, *internal_error.InternalError) {
	tolkenEntity := tolken_entity.NewTolken()
	tolkenString, err := tolkenEntity.GetTolkenString()

	if err != nil {
		return nil, err
	}

	err = tl.TolkenRepository.Save(ctx, tolkenEntity)

	if err != nil {
		return nil, err
	}

	tl.Expirer.SetExpiration(
		tolkenString,
		time.Duration(tolkenEntity.TimeLife)*time.Second,
		func() {
			tl.TolkenRepository.DeleteInfoByTolken(context.Background(), tolkenString)
		},
	)

	return &TolkenOutputDTO{
		Tolken: tolkenString,
	}, nil
}
