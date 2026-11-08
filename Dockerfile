# Build
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bin/api ./cmd/api

# Final image
FROM alpine:3.20

WORKDIR /app

# Создаем группу и пользователя с UID:GID 1000:1000
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

COPY --from=builder --chown=appuser:appgroup /app/bin/api /app/api

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/api"]