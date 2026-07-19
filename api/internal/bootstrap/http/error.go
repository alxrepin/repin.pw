package http

import (
	"net/http"

	"repin/internal/context/domain"
	"repin/internal/pkg/httpx"
)

const (
	codeNotFound = iota + 101
)

var errMap = map[error]httpx.ErrorCodes{
	domain.ErrPostNotFound:    {Internal: codeNotFound, HTTP: http.StatusNotFound},
	domain.ErrChannelNotFound: {Internal: codeNotFound, HTTP: http.StatusNotFound},
}
