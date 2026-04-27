.PHONY: dev test lint migrate-up migrate-down sqlc mock build stop

stop:
	-@kill -9 $$(lsof -t -i:8080) 2>/dev/null || true
	@echo "Server stopped."

dev: stop
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path internal/infra/postgres/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/infra/postgres/migrations -database "$$DATABASE_URL" down 1

sqlc:
	cd internal/infra/postgres && sqlc generate

mock:
	mockery --all --dir=internal/port --output=tests/mock --outpkg=mock

tidy:
	go mod tidy
