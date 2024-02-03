package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/yes312/metrics/internal/agent"
)

var f agent.Flags

func init() {

	flag.StringVar(&f.A, "a", "127.0.0.1:8080", "IP adress")
	flag.Uint64Var(&f.R, "r", 10, "reportInterval")
	flag.Uint64Var(&f.P, "p", 2, "pollInterval")
	flag.StringVar(&f.K, "k", "", "key")
	flag.Uint64Var(&f.L, "l", 2, "rate limit")
}

func main() {

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancel()
	}()

	config, err := agent.NewAgentConfig(f)
	if err != nil {
		log.Fatal(err)
	}
	a := agent.New(config)

	err = a.Start(ctx, &wg)
	if err != nil {
		log.Fatal(err)
	}

}
