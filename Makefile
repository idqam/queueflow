.PHONY: run-api run-worker run-autoscaler test lint docker-build local-up local-down seed flood dlq-demo migrate

run-api:
	go run ./cmds/api

run-worker:
	go run ./cmds/worker

run-autoscaler:
	go run ./cmds/autoscaler

test:
	go test ./... -race -cover

lint:
	go vet ./...
	golangci-lint run ./...

docker-build:
	docker build -f Dockerfile.api -t queueflow/api:latest .
	docker build -f Dockerfile.worker -t queueflow/worker:latest .
	docker build -f Dockerfile.autoscaler -t queueflow/autoscaler:latest .

local-up:
	docker compose up -d

local-down:
	docker compose down -v

migrate:
	psql $(DATABASE_URL) -f internal/db/migrations/001_create_jobs.sql
	psql $(DATABASE_URL) -f internal/db/migrations/002_create_job_events.sql
	psql $(DATABASE_URL) -f internal/db/migrations/003_create_workers.sql

seed:
	./scripts/seed_jobs.sh

flood:
	./scripts/flood_queue.sh

dlq-demo:
	./scripts/trigger_dlq.sh
