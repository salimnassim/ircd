package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"sync"

	irc "github.com/fluffle/goirc/client"
)

func main() {
	var n int
	var f bool
	var ch string

	flag.IntVar(&n, "n", 50, "number of clients")
	flag.BoolVar(&f, "flood", false, "flood privmsg")
	flag.StringVar(&ch, "chan", "#testing2", "channel to join")
	flag.Parse()

	fmt.Printf("starting %d clients, flood is %t", n, f)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go run(f, ch)
		defer wg.Done()
	}
	wg.Wait()
}

func run(f bool, ch string) {
	channels := []string{ch}
	cfg := irc.NewConfig(randomString(9, "f[", "]"))

	cfg.SSL = true
	cfg.Flood = true
	cfg.SSLConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	cfg.Server = "localhost:6697"
	cfg.NewNick = func(n string) string { return randomString(9, "f[", "]") }
	c := irc.Client(cfg)

	c.HandleFunc(irc.CONNECTED,
		func(conn *irc.Conn, line *irc.Line) {
			for _, ch := range channels {
				conn.Join(ch)
			}
			if f {
				go func(chs []string) {
					for {
						ch := chs[rand.Intn(len(chs))]
						c.Privmsg(ch, randomString(16, "", ""))
					}
				}(channels)
			}
		})

	quit := make(chan bool)
	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err.Error())
	}

	<-quit
}

func randomString(n int, prefix string, suffix string) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s%s", prefix, string(b), suffix)
}
