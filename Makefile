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
	docker compose run --rm migrate 

migrate-down:
	docker compose run --rm migrate down 1