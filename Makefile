.PHONY: test test-integration vet fmt rebuild sqlc migrate-up migrate-down auth
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
	docker compose build
	docker compose up -d --force-recreate

sqlc:
	sqlc generate

migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker compose run --rm migrate down 1

migrate-status:
	docker compose exec postgres psql -U music -d music_platform -c "SELECT version, dirty FROM schema_migrations;"

auth:
	@set -e; \
	printf "Email: "; read EMAIL; \
	printf "Password: "; stty -echo; read PASSWORD; stty echo; echo; \
	RESPONSE=$$(curl -sS -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d "{\"email\":\"$$EMAIL\",\"password\":\"$$PASSWORD\"}"); \
	TOKEN=$$(echo "$$RESPONSE" | jq -er '.access_token') || { \
		echo "Authentication failed:"; \
		echo "$$RESPONSE" | jq .; \
		exit 1; \
	}; \
	printf "%s" "$$TOKEN" > .token; \
	chmod 600 .token; \
	echo "Authentication successful. Token saved to .token"