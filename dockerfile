# Stage 1: BUILD
FROM --platform=$BUILDPLATFORM golang:1.26.1 AS builder

# mod caching 
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

# copying surce
COPY . .

# cross-compile respecting target platform
ARG TARGETOS
ARG TARGETARCH

# building the app
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/mqttIngester

# Stage 2: final image
FROM alpine:3.23.3

# copying only the app in the final image
WORKDIR /app
COPY --from=builder /build/app .

CMD ["./app"]