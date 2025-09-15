package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func envVar(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		if defaultValue == "" {
			log.Fatalf("%s is not set", name)
		}
		return defaultValue
	}
	return value
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	refresh, err := time.ParseDuration(envVar("REFRESH", "15s"))
	if err != nil {
		log.Fatalf("Failed to parse refresh duration: %s", err)
	}

	twitchAPI := NewTwitchAPI(
		envVar("CLIENT_ID", ""),
		envVar("CLIENT_SECRET", ""),
	)
	recorder := NewRecorder(
		twitchAPI,
		envVar("USERNAME", ""),
		envVar("QUALITY", "best"),
		envVar("TEMP_PATH", ""),
		envVar("FINAL_PATH", ""),
		envVar("WEB_API_TOKEN", ""),
		refresh,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	err = recorder.Run(ctx)
	if err != nil {
		log.Printf("Failed to run recorder: %s", err)
		return
	}
}
