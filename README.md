# ircd

An IRC server written in Go implementing a subset of RFC 1459. Clients and channels are actors running on their own goroutines, coordinated through directories and event channels; the server listens on plaintext and TLS ports simultaneously and exposes Prometheus metrics.

## Contents

- `cmd/ircd.go` the binary entrypoint: reads configuration from the environment, starts the plaintext and TLS listeners (generating a self-signed certificate as a fallback when none is configured), and optionally serves Prometheus metrics.
- `server.go`, `serve.go` `Server`/`ServerConfig` (including the RPL_ISUPPORT limits in `ServerConfigParameters`), and the accept loop that spawns a session per connection.
- `session.go`, `session_handlers.go` the per-client session state machine, command dispatch, and handler implementations (`PASS`, `NICK`, `USER`, `JOIN`, `PART`, `TOPIC`, `PRIVMSG`, `WHO`, `WHOIS`, `KICK`, `LUSERS`, `MODE`, `AWAY`, `QUIT`, `OPER`, `VERSION`, `LIST`, `INVITE`, and more).
- `handle_command_kill.go` the operator-only `KILL` command.
- `channel_actor.go`, `channel_events.go` the per-channel actor: membership, modes, topic, and bans, driven by events from sessions.
- `directory_channel.go`, `directory_client.go` concurrency-safe registries used to look up channels and clients by name/ID.
- `store_opers.go` in-memory operator credential store used by `OPER`.
- `command.go` outbound raw IRC line formatters (`PRIVMSG`, `NOTICE`, `MODE`, `JOIN`, `PART`, `KICK`, `INVITE`, `PING`, `ERROR`, ...).
- `rpl.go` numeric reply (`RPL_*`/`ERR_*`) formatters.
- `message.go`, `parse.go` IRC line parsing into a `message{tags, prefix, command, params}` struct.
- `mask.go` wildcard (`*`/`?`) mask matching used for bans and `WHO`/`WHOIS` filtering.
- `modes.go` mode bit flags and letter maps for client, channel, and channel membership modes.
- `metrics.go` Prometheus metrics: `ircd_clients`, `ircd_channels`, `ircd_command{name}`.

## Features

### IRC commands

- [X] TLS (with self-signed certificate fallback)
- [ ] CAP
- [x] PRIVMSG
- [x] NICK
- [x] USER
- [x] JOIN
- [x] PART
- [x] TOPIC
- [x] WHO
- [x] WHOIS
- [x] KICK
- [X] LUSERS
- [X] PASS
- [X] OPER (partial, gates KILL)
- [X] KILL (operator only)
- [X] LIST (partial, no ELIST)
- [X] INVITE
- [X] VERSION (partial, local server only)
- [ ] ADMIN
- [X] MODE (client: iorwtz, channel: ikmspCrORnzt, member: vhoaq)
- [X] AWAY
- [ ] LINK
- [ ] IRCv3

### Other

- Flood protection: token-bucket rate limiting per session, disconnects with `Excess Flood` when exceeded.
- SendQ limit: bounded outbox per session, disconnects with `SendQ exceeded` when full.
- Ping/pong keepalive with a configurable timeout, disconnects with `Timeout` when a client stops responding.
- Hostname cloaking: clients are assigned a vhost (`ipv4-`/`ipv6-`/`tls-<uuid>.vhost`) instead of exposing their real host.

## Install

### Local

1. Run `go mod download && make build` (or `go build -v -o ./dist/ircd ./cmd`).
2. The binary can be found under the `dist` directory.
3. Run with `SERVER_NAME=foo SERVER_VERSION=0.1 PORT=6667 ./dist/ircd`.

### Docker

Note: for clients to see their real remote IP instead of the Docker gateway address, the server needs to run using the host network driver.

1. Configure the environment variables in `docker-compose.yml`.
2. Run `docker compose up --build`.

### TLS certificate

If `TLS` is enabled and `TLS_CERTIFICATE`/`TLS_KEY` are not both set, the server generates an ephemeral self-signed certificate on startup and logs a warning — nothing further is needed to run with TLS. To use a real certificate pair instead:

```
mkdir tls && cd tls
openssl genrsa -out servercakey.pem
openssl req -new -x509 -key servercakey.pem -out serverca.crt
openssl genrsa -out server.key
openssl req -new -key server.key -out server_reqout.txt
openssl x509 -req -in server_reqout.txt -days 3650 -sha256 \
 -CAcreateserial -CA serverca.crt -CAkey servercakey.pem -out server.crt
```

## Usage

The server is configured entirely through environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `SERVER_NAME` | Server name, also used as the CN/SAN for the self-signed TLS fallback | `ircd.network.fqdn` |
| `SERVER_VERSION` | Version string shown in `RPL_VERSION` | `0.1` |
| `SERVER_PASSWORD` | Server connect password, checked during the `PASS`/`NICK`/`USER` handshake | `<empty>` (no password) |
| `NETWORK_NAME` | Network name shown in `RPL_WELCOME` | `Network` |
| `PORT` | Plaintext IRC listen port | `6667` |
| `PORT_TLS` | TLS IRC listen port (only used if `TLS` is set) | `6697` |
| `PROMETHEUS` | Enables the metrics HTTP server on `:2112` (presence enables it, value is ignored) | `true` |
| `TLS` | Enables the TLS listener (presence enables it, value is ignored) | `true` |
| `TLS_CERTIFICATE` | Path to a TLS certificate; if unset (with `TLS` on), a self-signed certificate is generated | `<empty>` (self-signed fallback) |
| `TLS_KEY` | Path to the matching TLS key; same fallback behavior as above | `<empty>` (self-signed fallback) |

```
SERVER_NAME=foo SERVER_VERSION=0.1 NETWORK_NAME=Network SERVER_PASSWORD= \
PORT=6667 PORT_TLS=6697 TLS=true TLS_CERTIFICATE= TLS_KEY= PROMETHEUS=true \
./dist/ircd
```

Metrics (if enabled) are exposed on `http://<host>:2112/metrics`.

## Development

```
make test    # go test -v ./...
make build   # CGO_ENABLED=0 go build -o ./dist/ircd ./cmd
make ci      # gofmt check, build, vet, race tests, mod tidy check, govulncheck
```
