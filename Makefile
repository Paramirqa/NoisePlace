.PHONY: test vet fmt rebuild sqlc migrate-up migrate-down

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

rebuild:
	docker compose build --no-cache
	docker compose up -d --force-recreate

sqlc:
	sqlc generate

migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down 1