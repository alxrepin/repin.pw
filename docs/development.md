# Локальная разработка

Docker нужен только для баз: Postgres и MinIO удобно поднять из общего
компоуза, а сами сервисы запускать нативно

```bash
make up        # или поднять только базы: docker compose up -d postgres minio minio-init
```

## API (Go 1.26)

```bash
cd api
cp .env.example .env      # DATABASE_URL указывает на localhost:5432
make migrate-up
make build && ./bin/http  # API на :8080
```

Команды `api/Makefile`:

| Команда                       | Что делает |
|-------------------------------| --- |
| `setup`                       | ставит golangci-lint, качает зависимости |
| `build`                       | собирает все бинарники в `bin/` (`http`, `sync`, `cli`, `bot`) |
| `test` / `lint`               | `go test ./...` / `golangci-lint` |
| `migrate-up` / `migrate-down` | миграции через `./bin/cli` |
| `migrate-create name=...`     | новая пара up/down-миграций |
| `sync` / `bot`                | импорт истории / демон обновлений |
| `rerender`                    | пересобрать посты из сохранённых сообщений — без похода в Telegram, удобно после правок нормализации текста |

## Фронт (Bun 1.3)

```bash
cd front
cp .env.example .env      # API-адреса указывают на localhost:8080
bun install
bun run dev               # SSR-сервер с HMR на :3000
```

Скрипты `front/package.json`:

| Скрипт | Что делает |
| --- | --- |
| `dev` | SSR + Vite HMR (`bun --watch server.ts`) |
| `build` | клиентский и серверный бандлы в `dist/` |
| `start` | прод-режим из `dist/` |
| `typecheck` | `vue-tsc --noEmit` |
| `lint` / `lint:check` | biome с фиксами / только проверка |

## CI

`.github/workflows/` — два пайплайна с фильтрами по путям, гоняются только
для затронутой части:

- **api.yml** — `golangci-lint` (v2), `go build`, `go test`
- **web.yml** — `biome`, `vue-tsc`, сборка Vite

Перед пушем локально: `make lint test` в `api/`,
`bun run lint:check && bun run typecheck && bun run build` в `front/`
