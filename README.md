# ircd

Simple IRC server which implements a subset of RFC1459. 

By default the server will run on port `6667` and prometheus metrics are exposed on port `2112`.

## Features

### IRC Commands

- [ ] CAP
- [x] PRIVMSG
- [x] NICK
- [x] USER
- [x] JOIN
- [x] PART
- [x] TOPIC
- [x] WHO
- [x] WHOIS
- [X] LUSERS
- [ ] PASS
- [ ] OPER
- [ ] LIST
- [ ] INVITE
- [ ] VERSION
- [ ] ADMIN
- [/] MODE (partial)
- [ ] AWAY

## Installation
### Local

1. Run `go mod download && go build -v -o ./dist/ircd ./cmd`.
2. The binary can be found under the `dist` directory.
3. Run with `SERVER_NAME=foo SERVER_VERSION=0.1 ./dist/ircd`

### Docker

1. Configure the environment variables in in `docker-compose.yml`.
2. Run `docker compose up`.