package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/arthneura/arthneura-market/internal/announce"
	"github.com/arthneura/arthneura-market/pkg/client"
)

func main() {
	log.SetFlags(0)
	id := flag.Int64("offer", 0, "offer id")
	price := flag.Int64("price", 0, "offer price")
	exp := flag.Int64("exp", 0, "unix expiry already on the offer")
	flag.Parse()
	if *id <= 0 || *price <= 0 || *exp <= 0 {
		log.Fatal("need -offer -price -exp")
	}
	did := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	seedHex := strings.TrimPrefix(strings.TrimSpace(os.Getenv("CONTROLLER_SEED")), "0x")
	sb, err := hex.DecodeString(seedHex)
	if did == "" || err != nil || len(sb) != 32 {
		log.Fatal("OWNER_DID and CONTROLLER_SEED required")
	}
	var seed [32]byte
	copy(seed[:], sb)
	msg := client.OfferSignMessage("accept", *id, did, *price, *exp)
	sig, _, err := announce.Sign(seed, msg)
	if err != nil {
		log.Fatal(err)
	}
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	body, _ := json.Marshal(map[string]any{
		"did": did, "price": *price, "expires_at": *exp,
		"signature": hex.EncodeToString(sig[:]),
	})
	res, err := http.Post(strings.TrimRight(u, "/")+"/v1/offers/"+fmt.Sprint(*id)+"/accept", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		log.Fatalf("accept %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	fmt.Printf("ACCEPT_OK\n%s\n", strings.TrimSpace(string(b)))
}
