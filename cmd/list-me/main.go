package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/arthneura/arthneura-market/internal/announce"
	"github.com/arthneura/arthneura-market/pkg/client"
)

func main() {
	log.SetFlags(0)
	title := flag.String("title", "", "what to sell")
	schema := flag.String("schema", "csv.v1", "schema id")
	price := flag.Int64("price", 0, "price")
	exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
	flag.Parse()
	if *title == "" || *price <= 0 {
		log.Fatal("need -title and -price")
	}
	did := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	if did == "" {
		log.Fatal("OWNER_DID required")
	}
	seedHex := strings.TrimPrefix(strings.TrimSpace(os.Getenv("CONTROLLER_SEED")), "0x")
	if seedHex == "" {
		log.Fatal("CONTROLLER_SEED required")
	}
	sb, err := hex.DecodeString(seedHex)
	if err != nil || len(sb) != 32 {
		log.Fatal("CONTROLLER_SEED must be 32-byte hex")
	}
	var seed [32]byte
	copy(seed[:], sb)
	msg, err := client.ListingSignMessage(client.CreateListingInput{
		SellerDid: did, Title: *title, Schema: *schema, Price: *price, ExpiresAt: *exp,
	})
	if err != nil {
		log.Fatal(err)
	}
	sig, _, err := announce.Sign(seed, msg)
	if err != nil {
		log.Fatal(err)
	}
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	item, err := client.New(u).CreateListing(client.CreateListingInput{
		SellerDid: did, Title: *title, Schema: *schema, Price: *price, ExpiresAt: *exp,
		Signature: hex.EncodeToString(sig[:]),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("LISTING_ID=%d\nSCHEMA=%s\nSELLER_DID=%s\nTITLE=%s\nPRICE=%d\n", item.ID, item.Schema, item.SellerDid, item.Title, item.Price)
}
