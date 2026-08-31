package main

import (
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

    go func() {
        if err := indexer.Run(ws); err != nil {
            log.Fatal(err)
        }
    }()

    ch := make(chan os.Signal, 1)
    signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
    <-ch
    log.Printf("indexer stop")
}
