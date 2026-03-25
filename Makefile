.PHONY: help up down build clean logs k8s-deploy k8s-status k8s-destroy k8s-port-forward k8s-logs k8s-pods frontend-dev frontend-build init migrate-db test-auth test-coverage build-auth build-api-gateway frontend-install test-frontend k8s-debug

# Docker Compose - Local Development
init: clean build up migrate-db

kubernetes: build up migrate-db k8s-restart k8s-status

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build

clean:
	docker compose down -v

logs:
	docker compose logs -f

# Docker Compose - Database Migrations (only works with Docker Compose, not K8s)
migrate-db:
	docker cp services/auth-service/migrations/001_create_users_table.up.sql auth-db:/migrate.sql
	docker exec auth-db psql -U auth_user -d auth_db -f /migrate.sql
	docker cp services/stream-service/migrations/001_create_streams_tables.up.sql stream-db:/migrate.sql
	docker exec stream-db psql -U stream_user -d stream_db -f /migrate.sql
	docker cp services/user-service/migrations/001_create_profiles.up.sql user-db:/migrate.sql
	docker exec user-db psql -U user_user -d user_db -f /migrate.sql
	docker cp services/chat-service/migrations/001_create_chat_tables.up.sql chat-db:/migrate.sql
	docker exec chat-db psql -U chat_user -d chat_db -f /migrate.sql
	docker cp services/notification-service/migrations/001_create_notifications.up.sql notification-db:/migrate.sql
	docker exec notification-db psql -U notif_user -d notification_db -f /migrate.sql

# Docker Compose - Individual service builds
build-auth:
	docker compose build auth-service

build-api-gateway:
	docker compose build api-gateway

# Frontend - Development (Node.js, npm only - K8s has containerized frontend)
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

test-frontend:
	cd frontend && npm test

# Go Testing (not deployment related - for local development only)
test-auth:
	cd services/auth-service && go test ./... -v

test-coverage:
	cd services/auth-service && go test ./... -coverprofile=coverage.out
	cd services/auth-service && go tool cover -html=coverage.out -o coverage.html

# Kubernetes - Production Deployment
k8s-deploy:
	kubectl apply -f infra/kubernetes/00-namespace.yaml
	kubectl apply -f infra/kubernetes/01-rbac.yaml
	kubectl apply -f infra/kubernetes/02-storageclass.yaml
	kubectl apply -f infra/kubernetes/03-pv-pvc.yaml
	kubectl apply -f infra/kubernetes/secrets/
	kubectl apply -f infra/kubernetes/deployments/

k8s-restart:
	kubectl rollout restart deployment -n livestream

k8s-status:
	kubectl get pods -n livestream
	kubectl get svc -n livestream

k8s-pods:
	kubectl get pods -n livestream -o wide

k8s-logs:
	kubectl logs -f -n livestream deployment/auth-service

k8s-port-forward:
	powershell -Command "Start-Process powershell -ArgumentList 'kubectl port-forward -n livestream svc/api-gateway 8080:80'"
	powershell -Command "Start-Process powershell -ArgumentList 'kubectl port-forward -n livestream svc/frontend 3000:80'"

k8s-destroy:
	kubectl delete namespace livestream --ignore-not-found

# Kubernetes - Debugging (init containers handle migrations K8s doesn't need manual migration)
k8s-debug:
	kubectl get pods -n livestream
	kubectl describe pod -n livestream