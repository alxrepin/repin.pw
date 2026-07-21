<div align="center">

# 📨 Персональный блог из Telegram-канала

Пишете пост в канал и он на сайте: с медиа, SEO-метаданными, фидами и Instant View

[![api](https://github.com/alxrepin/repin.pw/actions/workflows/api.yml/badge.svg)](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/api.yml)
[![web](https://github.com/alxrepin/repin.pw/actions/workflows/web.yml/badge.svg)](https://github.com/alxrepin/TGChannel2Blog/actions/workflows/web.yml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vuedotjs&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

- Бот **отслеживает** публикации, правки и удаления в канале
- SSR, canonical, Open Graph, JSON-LD, sitemap, RSS и `llms.txt`
- Поддержка [Instant View](docs/instant-view.md) внутри Telegram
- **SEO-friendly:** семантика, SEO/GEO-оптимизация, заголовки и описания генерирует LLM через OpenRouter

## Быстрый старт

```bash
cp .env.example .env     # прописать TELEGRAM_BOT_TOKEN (бот-админ канала)
make up                  # собрать и поднять весь стек
make sync                # опционально: импортировать историю канала
```

Сайт — <http://localhost:3000>, API — <http://localhost:8080/api/v1/posts>.
Подробно, шаг за шагом и с объяснениями — в **[docs/setup.md](docs/setup.md)**.

## Структура репозитория

```
.
├── api/                  # Go: HTTP API, импорт, бот, воркер, cli
├── front/                # Vue 3 + Vite: SSR-сайт на Bun
├── docs/                 # документация
├── .docker/
│   ├── container/        # Dockerfile'ы api и web
│   └── data/             # данные контейнеров (bind-mount, в .gitignore)
├── .github/workflows/    # CI: линт, тесты, сборка
├── compose.yml           # весь стек
└── Makefile              # обёртки над docker compose
```

Внутреннее устройство `api/` (слои) и `front/` (модули) разобрано в
[docs/architecture.md](docs/architecture.md)

## Документация

| Документ | Комментарий                                    |
| --- |------------------------------------------------|
| [setup.md](docs/setup.md) | установка шаг за шагом, переменные, продакшен  |
| [architecture.md](docs/architecture.md) | сервисы, слои API, структура фронта, эндпоинты |
| [sync.md](docs/sync.md) | импорт истории канала                          |
| [bot.md](docs/bot.md) | живые обновления: посты, канал, прокси         |
| [worker.md](docs/worker.md) | очередь задач: медиа, SEO                      |
| [instant-view.md](docs/instant-view.md) | как посты открываются в Telegram статьями      |
| [development.md](docs/development.md) | локальная разработка и CI                      |

## Лицензия

[MIT](LICENSE) © 2026 Alexander Repin
