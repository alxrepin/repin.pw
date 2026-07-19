package httpx

import (
	"context"
	"net/http"
)

type contextKey string

const (
	requestKey       contextKey = "http-request"
	requestWriterKey contextKey = "http-response-writer"
)

func Request(ctx context.Context) (*http.Request, bool) {
	r, ok := ctx.Value(requestKey).(*http.Request)
	return r, ok
}

func ResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(requestWriterKey).(http.ResponseWriter)
	return w, ok
}

func withRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	ctx = context.WithValue(ctx, requestKey, r)
	ctx = context.WithValue(ctx, requestWriterKey, w)

	return ctx
}
