run:
	go run ./cmd/api

vet:
	go vet ./...

test:
	go test ./...

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

db-logs:
	docker compose logs -f postgres
