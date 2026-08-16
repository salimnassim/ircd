package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/salimnassim/ircd"
)

// selfSignedCertificate generates an ephemeral ECDSA self-signed certificate,
// used as a fallback when no TLS certificate/key pair is configured. The
// subject and DNS SAN are set to fqdn (the configured server name), falling
// back to "localhost" when it's empty.
func selfSignedCertificate(fqdn string) (tls.Certificate, error) {
	if fqdn == "" {
		fqdn = "localhost"
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: fqdn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{fqdn},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func envIntDefault(key string, def int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		slog.Warn("invalid int env value, using default", "key", key, "value", v, "default", def)
		return def
	}
	return n
}

// loadCloakSecret returns the HMAC key used to derive cloaked hostnames from
// CLOAK_SECRET, generating an ephemeral one when unset. An ephemeral secret
// means cloaks (and any bans matched against them) won't survive a restart.
func loadCloakSecret() []byte {
	if secret, ok := os.LookupEnv("CLOAK_SECRET"); ok && secret != "" {
		return []byte(secret)
	}

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		slog.Error("unable to generate cloak secret", "err", err)
		os.Exit(1)
	}
	slog.Warn("CLOAK_SECRET not set, generated an ephemeral secret: cloaked hostnames and bans against them will not survive a restart")
	return secret
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	_, prometheusEnabled := os.LookupEnv("PROMETHEUS")
	if prometheusEnabled {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		go func() {
			slog.Info("starting http, listening on :2112")
			if err := http.ListenAndServe(":2112", mux); err != nil {
				slog.Error("metrics http server stopped", "err", err)
			}
		}()
	}

	_, tlsEnabled := os.LookupEnv("TLS")

	config := ircd.ServerConfig{
		Name:        os.Getenv("SERVER_NAME"),
		Password:    os.Getenv("SERVER_PASSWORD"),
		CloakSecret: loadCloakSecret(),
		Network:     os.Getenv("NETWORK_NAME"),
		Version:     os.Getenv("SERVER_VERSION"),
		MOTD: []string{
			"4This is the message of the day.",
			"5It contains multiple lines because the lines could be long.",
			"🍩🍫🍡🍦🍬🍮",
		},
		TLS:             tlsEnabled,
		CertificateFile: os.Getenv("TLS_CERTIFICATE"),
		CertificateKey:  os.Getenv("TLS_KEY"),
		PingFrequency:   30,
		PongMaxLatency:  10,

		MaxConnectionsGlobal: envIntDefault("MAX_CONNECTIONS_GLOBAL", 1000),
		MaxConnectionsPerIP:  envIntDefault("MAX_CONNECTIONS_PER_IP", 10),
		MaxConnectRatePerIP:  envIntDefault("MAX_CONNECT_RATE_PER_IP", 6),
		MaxConnectBurstPerIP: envIntDefault("MAX_CONNECT_BURST_PER_IP", 3),
		Parameters: ircd.ServerConfigParameters{
			MaxAwayLength:     128,
			CaseMapping:       "ascii",
			ChannelLimit:      "#&:64",
			ChannelModes:      "b,f,lk,ztSsrOmMiCc",
			MaxChannelLength:  50,
			ChannelTypes:      "&#",
			EList:             "",
			Excepts:           "",
			MaxHostnameLength: 32,
			MaxKickLength:     32,
			MaxList:           "b:16",
			MaxModes:          16,
			Network:           "Network",
			MaxNickLength:     31,
			ChannelPrefixes:   "(qaohv)~&@%+",

			MaxTargets:     "PRIVMSG:3,WHOIS:1,JOIN:3",
			MaxTopicLength: 128,
			MaxUserLength:  20,
		},
	}

	server := ircd.NewServer(config)

	ctx, cancel := context.WithCancel(context.Background())

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		slog.Error("cant listen", "err", err)
		os.Exit(1)
	}
	go func() {
		slog.Info(fmt.Sprintf("starting irc, listening on tcp:%s", os.Getenv("PORT")))
		server.Serve(ctx, listener, false)
	}()

	var tlsListener net.Listener
	if config.TLS {
		var fallbackCert *tls.Certificate
		if config.CertificateFile == "" || config.CertificateKey == "" {
			cert, err := selfSignedCertificate(config.Name)
			if err != nil {
				slog.Error("cant generate self-signed certificate", "err", err)
				os.Exit(1)
			}
			fallbackCert = &cert
			slog.Warn("no TLS certificate/key configured, using a generated self-signed certificate")
		}

		tlsListener, err = tls.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("PORT_TLS")),
			&tls.Config{
				GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
					if fallbackCert != nil {
						return fallbackCert, nil
					}
					cert, err := tls.LoadX509KeyPair(config.CertificateFile, config.CertificateKey)
					if err != nil {
						return nil, err
					}
					return &cert, nil
				},
			})
		if err != nil {
			slog.Error("cant listen tls", "err", err)
			os.Exit(1)
		}
		go func() {
			slog.Info(fmt.Sprintf("starting irc, listening on tcp:%s TLS", os.Getenv("PORT_TLS")))
			server.Serve(ctx, tlsListener, true)
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	slog.Info("shutting down")
	cancel()
	listener.Close()
	if tlsListener != nil {
		tlsListener.Close()
	}
	server.Shutdown(10 * time.Second)
}
