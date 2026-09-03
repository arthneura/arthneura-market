package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "log"
    "os"
    "time"

    "github.com/arthneura/arthneura-market/internal/announce"
)

func main() {
    id := flag.String("id", "", "commitment id hex")
    url := flag.String("url", "http://127.0.0.1:8090", "deliver url")
    exp := flag.Int64("exp", time.Now().Unix()+3600, "unix expiry")
    flag.Parse()
    seedHex := os.Getenv("ANNOUNCE_SEED")
    if *id == "" || seedHex == "" {
        log.Fatal("ANNOUNCE_SEED and -id required")
    }
    sb, err := hex.DecodeString(seedHex)
    if err != nil || len(sb) != 32 {
        log.Fatal("ANNOUNCE_SEED must be 32-byte hex")
    }
    var seed [32]byte
    copy(seed[:], sb)
    msg := announce.Message(*id, *url, *exp)
    sig, pub, err := announce.Sign(seed, msg)
    if err != nil {
        log.Fatal(err)
    }
    body := map[string]any{
        "url":        *url,
        "expires_at": *exp,
        "signature":  hex.EncodeToString(sig[:]),
        "public":     hex.EncodeToString(pub[:]),
    }
    _ = json.NewEncoder(os.Stdout).Encode(body)
}
