# Stage 1: BUILD
FROM golang:1.26.1 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/mqttIngester

# Stage 2: final image
FROM alpine:3.23.3

WORKDIR /app
COPY --from=builder /build/app .

CMD ["./app"]