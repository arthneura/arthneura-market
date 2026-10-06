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
	lid := flag.Int64("listing", 0, "listing id")
	price := flag.Int64("price", 0, "price")
	exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
	flag.Parse()
	if *lid <= 0 || *price <= 0 {
		log.Fatal("need -listing and -price")
	}
	did := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	if did == "" {
		log.Fatal("OWNER_DID required")
	}
	seedHex := strings.TrimPrefix(strings.TrimSpace(os.Getenv("CONTROLLER_SEED")), "0x")
	sb, err := hex.DecodeString(seedHex)
	if err != nil || len(sb) != 32 {
		log.Fatal("CONTROLLER_SEED must be 32-byte hex")
	}
	var seed [32]byte
	copy(seed[:], sb)
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	c := client.New(u)
	ls, err := c.ListListings()
	if err != nil {
		log.Fatal(err)
	}
	seller := ""
	for _, L := range ls {
		if L.ID == *lid {
			seller = strings.TrimPrefix(L.SellerDid, "0x")
			break
		}
	}
	if seller == "" {
		log.Fatal("listing not found")
	}
	if seller == did {
		log.Fatal("cannot offer on your own listing")
	}
	msg := client.OfferSignMessage("create", *lid, did, *price, *exp)
	sig, _, err := announce.Sign(seed, msg)
	if err != nil {
		log.Fatal(err)
	}
	item, err := c.CreateOffer(client.CreateOfferInput{
		ListingID: *lid,
		FromDid:   did,
		ToDid:     seller,
		Price:     *price,
		ExpiresAt: *exp,
		Signature: hex.EncodeToString(sig[:]),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("OFFER_ID=%d\nLISTING_ID=%d\nFROM_DID=%s\nTO_DID=%s\nPRICE=%d\n", item.ID, item.ListingID, item.FromDid, item.ToDid, item.Price)
}
