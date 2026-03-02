.PHONY: up down test-auth test-coverage migrate-auth frontend-install frontend-dev

up:
	docker compose up -d

down:
	docker compose down

test-auth:
	cd services/auth-service && go test ./... -v

test-coverage:
	cd services/auth-service && go test ./... -coverprofile=coverage.out
	cd services/auth-service && go tool cover -html=coverage.out -o coverage.html

migrate-auth:
	migrate -path ./services/auth-service/migrations \
	        -database "postgres://auth_user:auth_pass@localhost:5432/auth_db?sslmode=disable" up

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

test-frontend:
	cd frontend && npm test