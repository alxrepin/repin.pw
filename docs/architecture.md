# Архитектура

Два сервиса, у каждого одна ответственность: Go-API владеет данными, Vue-фронт — рендерингом, SSR. Общаются только по HTTP

Принципы:

- **Импорт — команда, а не конвейер.** История канала заливается разовым
  `cmd/sync`; живые обновления ловит отдельный демон `cmd/bot` через MTProto
- **Публикация ничего не ждёт.** Скачивание медиа и генерация SEO ставятся в
  очередь `jobs` в одной транзакции с постом и выполняются `cmd/worker`
- **Фронт ходит в API напрямую.** Браузер — на публичный адрес API (CORS
  разрешён), SSR-сервер — на внутренний (`API_INTERNAL_URL`)

## Сервисы (compose.yml)

| Сервис | Команда | Роль |
| --- | --- | --- |
| `postgres` | — | посты, канал, очередь задач |
| `minio` (+`minio-init`) | — | медиа постов, публичное чтение |
| `migrate` | `./bin/cli migrations up` | миграции, разовый |
| `api` | `./bin/http` | HTTP JSON API + фиды |
| `web` | `bun server.ts` | SSR-фронт |
| `bot` | `./bin/bot` | живые обновления канала → [bot.md](bot.md) |
| `worker` | `./bin/worker` | очередь: медиа, SEO → [worker.md](worker.md) |
| `sync` | `./bin/sync` | импорт истории, профиль `tools` → [sync.md](sync.md) |

## API (`api/`)

Go без внешних фреймворков: маленький встроенный «фреймворк» в
`internal/pkg`, поверх него — слои domain / application / infrastructure /
presentation. Зависимости направлены внутрь: domain ни о ком не знает

```
api/
├── cmd/                  # точки входа, по бинарнику на процесс
│   ├── http/             #   HTTP API
│   ├── bot/              #   демон живых обновлений
│   ├── worker/           #   демон очереди задач
│   ├── sync/             #   импорт истории канала
│   └── cli/              #   миграции, rerender
├── internal/
│   ├── bootstrap/        # сборка приложений: DI, конфиг, запуск
│   ├── context/
│   │   ├── domain/       # сущности: Post, Channel, Job — чистые типы
│   │   ├── application/
│   │   │   ├── service/  #   сервисы чтения (посты, канал)
│   │   │   └── usecase/  #   сценарии: sync, watch, jobs, rerender, regenseo
│   │   ├── infrastructure/
│   │   │   ├── db/postgres/   # репозитории
│   │   │   ├── storage/minio/ # объектное хранилище
│   │   │   ├── telegram/      # MTProto-клиент (gotd), бот, сессии
│   │   │   ├── openrouter/    # генерация SEO через LLM
│   │   │   └── text/          # нормализация сообщений в HTML постов
│   │   └── presentation/
│   │       ├── http/     #   контроллеры: posts, channel, feeds, media
│   │       └── cli/      #   команды cli
│   └── pkg/              # встроенный фреймворк: config, httpx, db,
│                         #   logger, migration, proxyx, validator
├── migrations/           # SQL-миграции (пара up/down)
└── prompts/              # промпты SEO-генерации, вшиваются в бинарник
```

## Фронт (`front/`)

Vue 3 + Vite, SSR на Express под Bun. Один и тот же код рендерится на
сервере (`entry-server.ts`) и гидрируется в браузере (`entry-client.ts`);
данные, загруженные на сервере, передаются клиенту через
`window.__INITIAL_STATE__` (`shared/ssr/state.ts`)

```
front/
├── server.ts             # Express: SSR, прокси фидов, robots.txt, Instant View
├── server/
│   └── instantview.ts    # страница для краулера Telegram → instant-view.md
├── index.html            # шаблон с плейсхолдерами <!--app-*-->
├── src/
│   ├── app/              # каркас: main.ts, router, глобальные стили
│   ├── layouts/          # обёртка страниц (шапка, подвал)
│   ├── modules/          # фичи, каждая со своими pages/api/components
│   │   ├── home/         #   главная: лента постов
│   │   ├── posts/        #   страница поста + вёрстка контента
│   │   ├── channel/      #   данные канала (шапка, meta)
│   │   └── error/        #   404
│   ├── shared/           # переиспользуемое: api-клиент, composables,
│   │   │                 #   config (env, seo), ui-кит, типы
│   │   └── ssr/          # передача состояния сервер → клиент
│   ├── entry-server.ts   # рендер в строку для SSR
│   └── entry-client.ts   # гидрация в браузере
├── uno.config.ts         # UnoCSS
└── vite.config.ts
```

SEO обеспечивается на сервере: мета-теги и JSON-LD собираются через
`@unhead/vue` при рендере, фиды проксируются из API, для краулера Telegram
есть [отдельный рендер](instant-view.md)

## HTTP-эндпоинты

API (`api.repin.pw` / `localhost:8080`):

| Метод | Путь | Описание |
| --- | --- | --- |
| GET | `/api/v1/posts` | список постов, пагинация `?page&limit` |
| GET | `/api/v1/posts/{slug}` | пост + соседние |
| GET | `/api/v1/channel` | канал: название, описание, аватар |
| GET | `/health` | health-check |
| GET | `/sitemap.xml`, `/rss.xml` | фиды для поисковиков и читалок |
| GET | `/llms.txt`, `/llms-full.txt` | контент для LLM-краулеров |

Сайт (`repin.pw` / `localhost:3000`) отдаёт страницы и дополнительно:

- проксирует фиды API под своим доменом (`/sitemap.xml` и т.д.)
- генерирует `robots.txt`
- для `/posts/{slug}` с User-Agent Telegram отдаёт
  [Instant View-разметку](instant-view.md)
