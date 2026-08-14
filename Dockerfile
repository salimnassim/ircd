FROM golang:1.26.5 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o ./ircd ./cmd

FROM scratch
COPY --from=builder /app/ircd /app/ircd
CMD ["/app/ircd"]