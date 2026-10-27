.PHONY: test test-integration vet fmt rebuild sqlc migrate-up migrate-down
test:
	go test ./...


test-integration:
	@set -eu; \
	trap 'docker compose -p noiseplace-test -f compose.test.yaml rm -sf postgres-test >/dev/null 2>&1 || true' EXIT; \
	docker compose -p noiseplace-test -f compose.test.yaml rm -sf postgres-test >/dev/null 2>&1 || true; \
	docker compose -p noiseplace-test -f compose.test.yaml up -d postgres-test; \
	docker compose -p noiseplace-test -f compose.test.yaml run --rm migrate-test; \
	TEST_DATABASE_URL='postgres://music_test:noiseplace_test_only@127.0.0.1:5433/noiseplace_test?sslmode=disable' \
		go test -count=1 -v ./tests/integration
		
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