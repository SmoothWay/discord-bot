package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SmoothWay/discord-bot/internal/config"
	"github.com/getsentry/sentry-go"
)

func main() {

	cfg := config.MustLoad()

	// bot.Start()

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		TracesSampleRate: 1,
	})

	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGTERM, os.Interrupt)
	fmt.Println("Bot started...")
	<-sc

	// bot.Stop()
}
