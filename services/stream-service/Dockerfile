FROM golang:1.25.4-alpine AS builder
RUN apk add --no-cache ffmpeg
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o stream-service ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ffmpeg ca-certificates
COPY --from=builder /app/stream-service /stream-service
COPY --from=builder /app/migrations /migrations
EXPOSE 8083 1935
ENTRYPOINT ["/stream-service"]