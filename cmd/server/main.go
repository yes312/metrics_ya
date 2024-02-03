package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yes312/metrics/internal/server/server"
)

var f server.Flags

func init() {

	flag.StringVar(&f.A, "a", "127.0.0.1:8080", "IP adress")
	flag.StringVar(&f.F, "f", "/tmp/metrics-db.json", "File storage path")
	// flag.StringVar(&f.D, "d", "postgres://postgres:12345@localhost:5432/metricsdb", "database uri")
	// flag.StringVar(&f.F, "f", "", "File storage path")
	flag.StringVar(&f.D, "d", "", "database uri")
	flag.IntVar(&f.I, "i", 300, "Store interval")
	flag.BoolVar(&f.R, "r", true, "Is restore?")
	flag.StringVar(&f.K, "k", "", "key")

}

func main() {

	flag.Parse()

	config, err := server.NewConfig(f)
	if err != nil {
		log.Fatal(err)
	}

	s := server.New(config)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-c
		cancel()

		fmt.Println("Программа получила сигнал от пользователя и завершается.")
		os.Exit(0)
	}()
	defer s.Close()
	if err := s.Start(ctx); err != nil {
		log.Fatal(err)
	}

}
