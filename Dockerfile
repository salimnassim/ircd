FROM golang:latest as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./ircd ./cmd

FROM scratch
COPY --from=builder /app/ircd /app/ircd
CMD ["/app/ircd"]