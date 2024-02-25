# ircd

Somewhat complex IRC server which implements a subset of RFC1459.

By default the server will run on port `6667` and prometheus metrics are exposed on port `2112`. See `docker-compose.yml` for example environment variable configuration.

## Features

### IRC Commands

- [X] TLS
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
- [X] OPER (pertial, no commands)
- [ ] LIST
- [ ] INVITE
- [X] VERSION (partial, local server only)
- [ ] ADMIN
- [X] MODE (partial)
- [X] AWAY
- [ ] LINK
- [ ] IRCv3

## Environment variables

- SERVER_NAME (string)
- SERVER_VERSION (string)
- PORT (int)
- PORT_TLS (int)
- PROMETHEUS (unset is false)
- TLS (unset is false)
- TLS_CERTIFICATE (path)
- TLS_KEY (path)

## Installation

### Generate TLS Key Pair

```
mkdir tls && cd tls
openssl genrsa -out servercakey.pem
openssl req -new -x509 -key servercakey.pem -out serverca.crt
openssl genrsa -out server.key
openssl req -new -key server.key -out server_reqout.txt
openssl x509 -req -in server_reqout.txt -days 3650 -sha256 \
 -CAcreateserial -CA serverca.crt -CAkey servercakey.pem -out server.crt
```

### Local

1. Run `go mod download && go build -v -o ./dist/ircd ./cmd`.
2. The binary can be found under the `dist` directory.
3. Run with `SERVER_NAME=foo SERVER_VERSION=0.1 PORT=6667 ./dist/ircd`

### Docker

Note: in order for clients to discover their IP address and get the real remote IP address, the server needs to run using the host network driver.
If host networking is not enabled all clients will use the Docker gateway address which might lead to interesting situations. 

1. Configure the environment variables in in `docker-compose.yml`.
2. Run `docker compose up`.