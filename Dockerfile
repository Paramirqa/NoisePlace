# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS build

WORKDIR /src

# Copy dependency manifests
COPY go.mod go.sum ./

# Download dependencies with persistent BuildKit cache
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy application source
COPY . .

# Build application
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server ./cmd/server


FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=build /out/server /server

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/server"]