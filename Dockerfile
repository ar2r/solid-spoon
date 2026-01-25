FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Build arg для версии
ARG APP_VERSION=unknown

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.Version=${APP_VERSION}" -o bot ./cmd/bot

FROM alpine:latest

RUN apk --no-cache add ca-certificates python3 curl && \
    curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app

COPY --from=builder /app/bot .

# Версия приложения как переменная окружения
ARG APP_VERSION=unknown
ENV APP_VERSION=${APP_VERSION}

CMD ["./bot"]
