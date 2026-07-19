# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/http ./cmd/http && \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/sync ./cmd/sync && \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/cli  ./cmd/cli && \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/bot  ./cmd/bot && \
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/worker ./cmd/worker

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/http /out/sync /out/cli /out/bot /out/worker ./bin/
COPY migrations ./migrations

EXPOSE 8080

CMD ["./bin/http"]
