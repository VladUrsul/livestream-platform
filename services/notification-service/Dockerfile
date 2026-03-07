FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o notification-service ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/notification-service /notification-service
EXPOSE 8085
ENTRYPOINT ["/notification-service"]