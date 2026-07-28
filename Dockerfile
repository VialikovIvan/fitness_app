FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o myapp ./cmd/app

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/myapp .
COPY --from=builder /app/migrations ./migrations
CMD ["./myapp"]