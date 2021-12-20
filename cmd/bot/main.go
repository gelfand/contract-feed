package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/gelfand/contract-feed/core"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	chatID, err := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
	if err != nil {
		log.Fatalf("Could not parse $TELEGRAM_CHAT_ID env variable: %v", err)
	}

	cfg := core.Config{
		TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
		TelegramChatID: chatID,
		RpcAddress:     os.Getenv("RPC_ADDRESS"),
	}

	c, err := core.NewCoordinator(ctx, cfg)
	if err != nil {
		log.Fatalf("Could not create new service coordinator: %v", err)
	}

	if err = c.Run(ctx); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}
