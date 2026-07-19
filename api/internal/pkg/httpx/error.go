package httpx

import (
	"errors"
	"net/http"
)

type ErrorCodes struct {
	Internal int
	HTTP     int
}

const (
	internalError = 100
)

var (
	ErrInternalServer = errors.New("internal server error")
	ErrRequestDecode  = errors.New("error decoding request")
)

var defaultErrMap = map[error]ErrorCodes{
	ErrInternalServer: {Internal: internalError, HTTP: http.StatusInternalServerError},
	ErrRequestDecode:  {Internal: internalError, HTTP: http.StatusBadRequest},
}
