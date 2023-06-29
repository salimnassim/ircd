FROM golang:latest as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./ircd ./cmd

FROM debian:stable-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/ircd /app/ircd

CMD ["/app/ircd"]