.PHONY: up down build logs ps migrate sync bot worker jobs clean

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

logs:
	docker compose logs -f

ps:
	docker compose ps

migrate:
	docker compose run --rm migrate

sync:
	docker compose run --rm sync

bot:
	docker compose --profile bot up -d --build bot

worker:
	docker compose logs -f worker

# Jobs that exhausted their retries, with the error that killed them.
jobs:
	docker compose exec postgres psql -U "$${POSTGRES_USER:-repin}" -d "$${POSTGRES_DB:-repin}" \
		-c "SELECT id, kind, dedup_key, attempts, last_error FROM jobs WHERE status = 'failed' ORDER BY updated_at DESC;"

clean:
	docker compose down -v
