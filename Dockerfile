FROM golang:1.26.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o subscription-service ./cmd/app

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/subscription-service .

EXPOSE 8080

CMD ["./subscription-service"]