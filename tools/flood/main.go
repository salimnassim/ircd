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
	var n = flag.Int("n", 10, "number of clients")

	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		go run()
		defer wg.Done()
	}
	wg.Wait()
}

func run() {
	channels := []string{"#testing1", "#testing2", "#testing3", "#testing4", "#testing5"}
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
			// go func(chs []string) {
			// 	for {
			// 		ch := chs[rand.Intn(len(chs))]
			// 		c.Privmsg(ch, randomString(32, "", ""))
			// 	}
			// }(channels)
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
