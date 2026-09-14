package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "log"
    "os"
    "time"

    "github.com/arthneura/arthneura-market/internal/announce"
    "github.com/vedhavyas/go-subkey/v2"
    "github.com/vedhavyas/go-subkey/v2/sr25519"
)

func main() {
    id := flag.String("id", "", "commitment id hex")
    url := flag.String("url", "http://127.0.0.1:8090", "deliver url")
    exp := flag.Int64("exp", time.Now().Unix()+3600, "unix expiry")
    flag.Parse()
    if *id == "" {
        log.Fatal("-id required")
    }
    msg := announce.Message(*id, *url, *exp)
    sig, pub, err := sign(msg)
    if err != nil {
        log.Fatal(err)
    }
    _ = json.NewEncoder(os.Stdout).Encode(map[string]any{
        "url":        *url,
        "expires_at": *exp,
        "signature":  hex.EncodeToString(sig[:]),
        "public":     hex.EncodeToString(pub[:]),
    })
}

func sign(msg []byte) ([64]byte, [32]byte, error) {
    who := os.Getenv("SIGNER")
    if who == "alice" || who == "bob" {
        uri := "//Alice"
        if who == "bob" {
            uri = "//Bob"
        }
        kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, uri)
        if err != nil {
            return [64]byte{}, [32]byte{}, err
        }
        sigb, err := kp.Sign(msg)
        if err != nil {
            return [64]byte{}, [32]byte{}, err
        }
        var sig [64]byte
        var pub [32]byte
        copy(sig[:], sigb)
        copy(pub[:], kp.Public())
        return sig, pub, nil
    }
    seedHex := os.Getenv("ANNOUNCE_SEED")
    if seedHex == "" {
        return [64]byte{}, [32]byte{}, errSeed{}
    }
    sb, err := hex.DecodeString(seedHex)
    if err != nil || len(sb) != 32 {
        return [64]byte{}, [32]byte{}, errSeed{}
    }
    var seed [32]byte
    copy(seed[:], sb)
    return announce.Sign(seed, msg)
}

type errSeed struct{}

func (errSeed) Error() string { return "set SIGNER=alice|bob or ANNOUNCE_SEED" }
