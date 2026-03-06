.PHONY: help init up down build logs build-auth build-api-gateway clean migrate-db test-auth test-coverage frontend-install frontend-dev frontend-build test-frontend

init: clean build up migrate-db

up:
	docker compose up -d
	@echo "All services started. Frontend: http://localhost:3000"

down:
	docker compose down

build:
	docker compose build

clean:
	docker compose down -v

migrate-db:
	@echo "Running database migrations..."
	docker cp services/auth-service/migrations/001_create_users_table.up.sql auth-db:/migrate.sql
	docker exec auth-db psql -U auth_user -d auth_db -f /migrate.sql
	docker cp services/stream-service/migrations/001_create_streams_tables.up.sql stream-db:/migrate.sql
	docker exec stream-db psql -U stream_user -d stream_db -f /migrate.sql
	docker cp services/user-service/migrations/001_create_profiles.up.sql user-db:/migrate.sql
	docker exec user-db psql -U user_user -d user_db -f /migrate.sql
	docker cp services/chat-service/migrations/001_create_chat_tables.up.sql chat-db:/migrate.sql
	docker exec chat-db psql -U chat_user -d chat_db -f /migrate.sql
	@echo "Database migrations completed"

logs:
	docker compose logs -f

test-auth:
	cd services/auth-service && go test ./... -v

test-coverage:
	cd services/auth-service && go test ./... -coverprofile=coverage.out
	cd services/auth-service && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at services/auth-service/coverage.html"

build-auth:
	docker compose build auth-service

build-api-gateway:
	docker compose build api-gateway

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

test-frontend:
	cd frontend && npm test