FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-api ./cmd/api

FROM scratch
COPY --from=builder /app/payment-api /payment-api
ENTRYPOINT ["/payment-api"]
