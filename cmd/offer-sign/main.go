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
    "github.com/vedhavyas/go-subkey/v2"
    "github.com/vedhavyas/go-subkey/v2/sr25519"
)

func main() {
    action := flag.String("action", "create", "create|counter|listing")
    id := flag.Int64("id", 0, "listing id (create) or offer id (counter)")
    did := flag.String("did", "", "signer agent did hex")
    title := flag.String("title", "", "listing title (action=listing)")
	schema := flag.String("schema", "", "listing schema id (action=listing)")
    price := flag.Int64("price", 0, "price")
    exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
    flag.Parse()

    if *action != "listing" && *did == "" {
        log.Fatal("-did required")
    }
    if *action == "listing" && (*did == "" || *title == "") {
        log.Fatal("listing needs -did and -title")
    }

    var msg []byte
    if *action == "listing" {
        msg = offersign.ListingMessage(*did, *title, *price, *exp, *schema)
    } else {
        msg = offersign.Message(*action, *id, *did, *price, *exp)
    }

    sig, pub, err := sign(msg)
    if err != nil {
        log.Fatal(err)
    }

    _ = json.NewEncoder(os.Stdout).Encode(map[string]any{
        "action":     *action,
        "id":         *id,
        "did":        *did,
        "title":      *title,
		"schema":     *schema,
        "price":      *price,
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
        return [64]byte{}, [32]byte{}, errSeed()
    }
    sb, err := hex.DecodeString(seedHex)
    if err != nil || len(sb) != 32 {
        return [64]byte{}, [32]byte{}, errSeed()
    }
    var seed [32]byte
    copy(seed[:], sb)
    return announce.Sign(seed, msg)
}

type seedErr struct{}

func errSeed() error { return seedErr{} }
func (seedErr) Error() string {
    return "set SIGNER=alice|bob or ANNOUNCE_SEED (32-byte hex)"
}
