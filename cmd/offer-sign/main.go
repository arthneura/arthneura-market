package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "log"
    "os"
    "time"

    "github.com/arthneura/arthneura-market/internal/announce"
    "github.com/arthneura/arthneura-market/internal/offersign"
)

func main() {
    action := flag.String("action", "create", "create|counter")
    id := flag.Int64("id", 0, "listing id (create) or offer id (counter)")
    did := flag.String("did", "", "signer agent did hex")
    price := flag.Int64("price", 0, "price")
    exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
    flag.Parse()
    seedHex := os.Getenv("ANNOUNCE_SEED")
    if *did == "" || seedHex == "" {
        log.Fatal("ANNOUNCE_SEED and -did required")
    }
    sb, err := hex.DecodeString(seedHex)
    if err != nil || len(sb) != 32 {
        log.Fatal("ANNOUNCE_SEED must be 32-byte hex")
    }
    var seed [32]byte
    copy(seed[:], sb)
    msg := offersign.Message(*action, *id, *did, *price, *exp)
    sig, pub, err := announce.Sign(seed, msg)
    if err != nil {
        log.Fatal(err)
    }
    _ = json.NewEncoder(os.Stdout).Encode(map[string]any{
        "action":     *action,
        "id":         *id,
        "did":        *did,
        "price":      *price,
        "expires_at": *exp,
        "signature":  hex.EncodeToString(sig[:]),
        "public":     hex.EncodeToString(pub[:]),
    })
}
