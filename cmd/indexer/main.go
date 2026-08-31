package main

import (
    "fmt"
    "os"
)

func main() {
    ws := os.Getenv("CHAIN_WS")
    if ws == "" {
        ws = "ws://127.0.0.1:9944"
    }
    fmt.Println("arthneura-market indexer stub, chain=", ws)
}
