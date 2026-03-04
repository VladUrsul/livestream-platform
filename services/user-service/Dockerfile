FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o user-service ./cmd/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/user-service /user-service
COPY --from=builder /app/migrations /migrations
EXPOSE 8082
ENTRYPOINT ["/user-service"]