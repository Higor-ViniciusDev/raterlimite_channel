package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/rest_err"
	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/usecase/tolken_usecase"
)

type TolkenController struct {
	TolkenUsecase tolken_usecase.TolkenUsecaseInterface
}

func NewTolkenController(tolkenusecase tolken_usecase.TolkenUsecaseInterface) *TolkenController {
	return &TolkenController{
		TolkenUsecase: tolkenusecase,
	}
}

func (tc *TolkenController) CreateTolken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dtoTolkenOutput, err := tc.TolkenUsecase.CreateTolken(context.Background())
	if err != nil {
		restErro := rest_err.ConvertInternalErrorToRestError(err)

		w.WriteHeader(restErro.Code)
		json.NewEncoder(w).Encode(restErro)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(dtoTolkenOutput); err != nil {
		restError := rest_err.NewInternalServerError("failed to encode response")

		w.WriteHeader(restError.Code)
		json.NewEncoder(w).Encode(restError)
		return
	}
}
