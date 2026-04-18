# ─── Build stage ─────────────────────────────────────
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies separately for better layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /dentdesk ./cmd/dentdesk

# ─── Runtime stage ────────────────────────────────────
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /dentdesk .

EXPOSE 8080

ENTRYPOINT ["./dentdesk"]
