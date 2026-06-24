# Dev-only image for the fake 1С OData server (cmd/mock-1c).
# Mirrors the backend's multi-stage scratch build.

FROM golang:1.26.3-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# mock-1c imports internal/modules/integration/domain/entities, so internal/
# is required at build time.
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o mock-1c ./cmd/mock-1c

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/mock-1c /mock-1c
EXPOSE 9191
ENTRYPOINT ["/mock-1c"]
