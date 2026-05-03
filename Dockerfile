FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/stock-market ./cmd/api/main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/stock-market .

ENV GIN_MODE=release

CMD ["./stock-market"]