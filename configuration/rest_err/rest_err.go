package rest_err

import (
	"net/http"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/internal/internal_error"
)

type RestErr struct {
	Message string   `json:"message"`
	Err     string   `json:"err"`
	Code    int      `json:"code"`
	Causes  []Causes `json:"causes"`
}

type Causes struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (r *RestErr) Error() string {
	return r.Message
}

func ConvertInternalErrorToRestError(err *internal_error.InternalError) *RestErr {
	switch err.Err {
	case "bad_request":
		return NewBadRequestError(err.Error())
	case "not_found":
		return NewNotFoundError(err.Error())
	case "many_request":
		return NewManyRequestError(err.Error())
	default:
		return NewInternalServerError(err.Error())
	}
}

func NewInternalServerError(message string) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "internal_server",
		Code:    http.StatusInternalServerError,
		Causes:  nil,
	}
}

func NewBadRequestError(message string, causes ...Causes) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "bad_request",
		Code:    http.StatusBadRequest,
		Causes:  nil,
	}
}

func NewNotFoundError(message string) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "not_found",
		Code:    http.StatusNotFound,
		Causes:  nil,
	}
}

func NewManyRequestError(message string) *RestErr {
	return &RestErr{
		Message: message,
		Err:     "many_request",
		Code:    http.StatusTooManyRequests,
		Causes:  nil,
	}
}
