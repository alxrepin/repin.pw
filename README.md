<div align="center">

# 📨 repin.pw

**Персональный блог, который сам собирается из Telegram-канала**

Импортирует канал, нормализует сообщения в посты и отдаёт их по HTTP API;
фронт на Vue 3 с SSR рендерит блог для читателей и поисковых систем.

[![api](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/api.yml/badge.svg)](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/api.yml)
[![web](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/web.yml/badge.svg)](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/web.yml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

Монорепозиторий:

- **`api/`** — сервис на Go. Импортирует канал, нормализует сообщения в посты
  и отдаёт их по HTTP JSON API.
- **`front/`** — сайт на Vue 3 с SSR (Vite + Bun): рендерит блог для читателей
  и поисковых систем.

## Структура репозитория

```
.
├── .docker/              # всё, что связано с Docker
│   ├── container/        #   Dockerfile'ы сервисов
│   │   ├── api/          #     образ api (http + sync + cli + bot + worker)
│   │   └── web/          #     образ web (SSR)
│   └── data/             #   данные контейнеров (bind-mount, в .gitignore)
│       ├── postgres/     #     база
│       ├── minio/        #     объектное хранилище (медиа)
│       ├── telegram/     #     сессия Telegram
│       └── favicon/      #     аватар канала, отдаётся фронтом как favicon
├── .github/workflows/    # CI: линт и тесты
├── docs/                 # инструкции (напр. docs/sync.md)
├── api/                  # Go-сервис
├── front/                # Vue-сайт
├── compose.yml           # весь стек одной командой
└── Makefile              # обёртки над docker compose
```

Всё Docker-хозяйство собрано в корневом `.docker/`. Данные контейнеров лежат
явными папками в `.docker/data/` — их удобно смотреть и бэкапить; содержимое
игнорируется git, а сами папки создаются при первом запуске.

## Архитектура

```
Telegram-канал ──(cmd/sync, cmd/bot)──▶ Postgres ◀──(cmd/http)── HTTP API ◀── Vue SSR (front) ◀── браузер
                                  │        ▲                                    │
                            очередь ▼      │ медиа, SEO                   /api/v1/* (CORS)
                              jobs ──▶ cmd/worker ──▶ MinIO / OpenRouter
```

- API разбит на слои (domain / application / infrastructure / presentation),
  поверх небольшого встроенного фреймворка в `api/internal/pkg` (конфиг, http,
  db, логгер, миграции) — без внешних фреймворков.
- Импорт канала — разовая команда (`cmd/sync`), а не событийный конвейер.
- Долгие операции импорт не ждёт: скачивание медиа и генерация SEO-метаданных
  ставятся в очередь `jobs` в той же транзакции, что и сам пост, и выполняются
  демоном `cmd/worker`.
- Живые обновления — отдельный демон (`cmd/bot`): бот-администратор канала
  ловит публикации, правки и удаления постов и периодически обновляет
  информацию о канале.
- Фронт ходит в API напрямую (API отдаёт разрешающий CORS), рендерит страницу
  на сервере ради SEO и гидрирует её в браузере.

## Требования

- Docker и Docker Compose — чтобы поднять стек.
- Для локальной разработки: Go 1.26, Bun 1.3.

## Быстрый старт

```bash
cp .env.example .env          # заполнить переменные (см. ниже)
make up                       # собрать и поднять postgres, minio, api, web
```

- Сайт: <http://localhost:3000>
- API: <http://localhost:8080/api/v1/posts>
- Консоль MinIO: <http://localhost:9001>

`make down` останавливает стек. `make clean` дополнительно удаляет тома, но
данные в `.docker/data/` остаются на диске — их чистят вручную.

## Наполнение из Telegram

Шаг необязательный: стек поднимается и с пустым блогом. Чтобы залить посты из
своего канала, нужно зарегистрировать приложение Telegram и запустить импорт:

```bash
make sync                     # первый вход в Telegram — интерактивный (код + 2FA)
```

Пошаговая инструкция (регистрация приложения, `api_id`/`api_hash`, авторизация,
пересборка постов) — в **[docs/sync.md](docs/sync.md)**.

## Бот: живые обновления канала

Чтобы блог обновлялся сам, без запусков `make sync`, есть демон `cmd/bot`.
Он работает через MTProto от имени бота и делает две вещи:

- следит за постами: публикация и правка импортируются сразу, удалённый
  в канале пост удаляется и из блога (Bot API таких событий не отдаёт,
  поэтому бот работает через MTProto);
- раз в `TELEGRAM_CHANNEL_REFRESH_INTERVAL` (по умолчанию `6h`) обновляет
  информацию о канале: название, описание, аватар, число подписчиков.

Состояние очереди обновлений хранится в `.docker/data/telegram/bot/` — после
рестарта бот добирает пропущенные события.

Запуск:

```bash
# 1. Создать бота у @BotFather и добавить его администратором канала
#    (боту хватает права «Изменение публикаций»).
# 2. Прописать TELEGRAM_BOT_TOKEN в .env.
make bot                      # docker compose --profile bot up -d
```

Если Telegram недоступен напрямую, задайте `PROXY_URL` — например
`http://user:password@host:3128`. Прокси используется всем исходящим трафиком:
`make sync`, ботом, воркером и запросами к OpenRouter. MTProto — сырой TCP,
поэтому он идёт через CONNECT-туннель: прокси должен разрешать `CONNECT` на
порт 443. Схема `socks5://` тоже принимается.

## Воркер: медиа и SEO

`cmd/worker` разбирает очередь фоновых задач. Импорт в неё только пишет, так
что публикация поста не ждёт ни скачивания видео, ни ответа модели.

| Задача           | Что делает                                                     |
| ---------------- | -------------------------------------------------------------- |
| `media.download` | скачивает медиа сообщения из Telegram и кладёт в MinIO          |
| `post.seo`       | заполняет `seo_title`, `seo_description`, `seo_keywords`        |

Очередь — таблица `jobs` в Postgres, задачи разбираются через
`FOR UPDATE SKIP LOCKED`. Ключевые свойства:

- **Ничего не теряется.** Задача ставится в одной транзакции с постом, так что
  «пост сохранён, а задача пропала» невозможно физически.
- **Повторы.** Неудачная попытка возвращается в очередь с экспоненциальной
  задержкой (30с → 30м). После `max_attempts` задача помечается `failed` и
  остаётся в таблице как dead letter: `make jobs` покажет её и последнюю ошибку.
- **Переживает падения.** Задачу, чей воркер умер, забирает обратно жнец аренды
  (`WORKER_JOB_LEASE`, по умолчанию 30 минут — должно превышать самое долгое
  скачивание).
- **Схлопывание дублей.** Серия правок одного поста оставляет одну задачу, а не
  очередь одинаковых.

Задача на медиа хранит только идентификаторы: ссылки Telegram на файлы
протухают за часы, поэтому воркер перезапрашивает сообщение в момент
выполнения — заодно он видит актуальное состояние, если пост успели
отредактировать или удалить.

SEO-метаданные генерируются через [OpenRouter](https://openrouter.ai): сначала
`OPENROUTER_MODEL`, при неудаче — повторы, затем `OPENROUTER_FALLBACK_MODEL`.
Модель вызывается для нового поста и при правке, но только если изменился сам
текст: Telegram шлёт событие правки и на смену медиа или реакций. Без
`OPENROUTER_API_KEY` воркер просто не берёт эти задачи, остальное работает.

Воркеру нужна сессия Telegram — либо `TELEGRAM_BOT_TOKEN` (тогда он логинится
сам), либо пользовательская сессия, оставшаяся после `make sync`.

```bash
make worker                   # логи воркера
make jobs                     # задачи, исчерпавшие попытки
```

## Локальная разработка

API:

```bash
cd api
cp .env.example .env
make migrate-up               # через ./bin/cli migrations up
make build && ./bin/http
```

Полезные цели `api/Makefile`: `build`, `test`, `lint`, `migrate-up`,
`migrate-create name=…`, `sync`, `bot`, `rerender`.

Фронт:

```bash
cd front
cp .env.example .env
bun install
bun run dev                   # SSR-сервер с HMR на :3000
```

Скрипты `front/package.json`: `dev`, `build`, `start`, `typecheck`, `lint`.

## Переменные окружения

Три файла `.env.example`, каждый копируется в `.env` рядом:

- **корень** — для `compose.yml` (Postgres, MinIO, Telegram, OpenRouter, прокси,
  публичный URL API, счётчик Яндекс.Метрики).
- **`api/.env.example`** — для локального запуска сервиса без Docker.
- **`front/.env.example`** — для локального запуска SSR-сервера.

Секреты Telegram берутся на <https://my.telegram.org>, ключ OpenRouter — на
<https://openrouter.ai/keys>.

## Эндпоинты

| Метод | Путь                   | Описание                  |
| ----- | ---------------------- | ------------------------- |
| GET   | `/api/v1/posts`        | Список постов (пагинация) |
| GET   | `/api/v1/posts/{slug}` | Пост по slug              |
| GET   | `/api/v1/channel`      | Информация о канале       |
| GET   | `/health`              | Health-check              |

## CI

`.github/workflows/` — два пайплайна с фильтрами по путям:

- **api** — `golangci-lint` (v2) и `go build` + `go test`.
- **web** — `biome`, `vue-tsc` (typecheck) и сборка Vite.

## Лицензия

[MIT](LICENSE) © 2026 Alex Repin
