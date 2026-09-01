package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/arthneura/arthneura-market/internal/indexer"
)

func main() {
    ws := os.Getenv("CHAIN_WS")
    if ws == "" {
        ws = "ws://127.0.0.1:9944"
    }
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://arthneura:arthneura@127.0.0.1:5432/arthneura_market?sslmode=disable"
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if err := indexer.Run(ctx, ws, dsn); err != nil && err != context.Canceled {
        log.Fatal(err)
    }
    log.Printf("indexer stop")
}
