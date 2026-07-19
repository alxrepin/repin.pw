package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"
)

type Paginate struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type APIResponse[Data, Item any] struct {
	Data     *Data     `json:"data"`
	Items    []Item    `json:"items"`
	Paginate *Paginate `json:"paginate"`
	Meta     any       `json:"meta"`
}

func NewAPIResponse[Data, Item any](data *Data, items []Item, paginate *Paginate, meta any) APIResponse[Data, Item] {
	if items == nil {
		items = []Item{}
	}

	return APIResponse[Data, Item]{Data: data, Items: items, Paginate: paginate, Meta: meta}
}

type APIErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSON(ctx context.Context, status int, body any) {
	w, ok := ResponseWriter(ctx)
	if !ok {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		zerolog.Ctx(ctx).Error().Err(err).Msg("encode response failed")
	}
}

func writeError(ctx context.Context, err error, errMap map[error]ErrorCodes) {
	code, status := internalError, http.StatusInternalServerError

	for target, codes := range errMap {
		if errors.Is(err, target) {
			code, status = codes.Internal, codes.HTTP
			break
		}
	}

	zerolog.Ctx(ctx).Error().Err(err).Int("code", code).Int("status", status).Msg("request failed")
	writeJSON(ctx, status, APIErrorResponse{Code: code, Message: http.StatusText(status)})
}
