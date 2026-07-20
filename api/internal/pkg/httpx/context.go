package httpx

import (
	"context"
	"net/http"
)

type contextKey string

const requestWriterKey contextKey = "http-response-writer"

func ResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(requestWriterKey).(http.ResponseWriter)
	return w, ok
}

func withResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, requestWriterKey, w)
}
