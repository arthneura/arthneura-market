package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/arthneura/arthneura-market/internal/api"
    "github.com/arthneura/arthneura-market/internal/store"
)

func main() {
    addr := os.Getenv("HTTP_ADDR")
    if addr == "" {
        addr = ":8080"
    }
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://arthneura:arthneura@127.0.0.1:5432/arthneura_market?sslmode=disable"
    }

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    db, err := store.Open(ctx, dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    srv := &http.Server{
        Addr:              addr,
        Handler:           api.NewMux(db),
        ReadHeaderTimeout: 5 * time.Second,
    }

    go func() {
        log.Printf("discovery api on %s", addr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    <-ctx.Done()
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(shutdownCtx)
    log.Printf("api stop")
}
