package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/arthneura/arthneura-market/internal/announce"
	"github.com/arthneura/arthneura-market/internal/offersign"
	"github.com/arthneura/arthneura-market/pkg/client"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	os.Args = append(os.Args[:1], os.Args[2:]...)

	switch cmd {
	case "listing":
		listing()
	case "offer":
		offer()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `arthneura-owner — sign on this machine, POST to the market.

MCP does not run this. The agent only reads. You (or a local shell the
agent is allowed to start) run one command.

  arthneura-owner listing -title "csv leads" -schema csv.v1 -price 1000
  arthneura-owner offer -listing 16 -price 1000

Env:
  MARKET_URL   default http://127.0.0.1:8080
  SIGNER       alice | bob   (lab controllers)
  OWNER_DID    agent DID hex, no 0x
`)
}

func market() *client.Client {
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	return client.New(u)
}

func did() string {
	d := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	if d == "" {
		log.Fatal("OWNER_DID required (hex, no 0x)")
	}
	return d
}

func listing() {
	title := flag.String("title", "", "what you are selling")
	schema := flag.String("schema", "csv.v1", "csv.v1|bytes.v1|api.v1|job.v1|meter.v1")
	price := flag.Int64("price", 0, "price")
	exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
	flag.Parse()
	if *title == "" || *price <= 0 {
		log.Fatal("listing needs -title and -price")
	}
	seller := did()
	msg, err := client.ListingSignMessage(client.CreateListingInput{
		SellerDid: seller,
		Title:     *title,
		Schema:    *schema,
		Price:     *price,
		ExpiresAt: *exp,
	})
	if err != nil {
		log.Fatal(err)
	}
	sig, _, err := sign(msg)
	if err != nil {
		log.Fatal(err)
	}
	item, err := market().CreateListing(client.CreateListingInput{
		SellerDid: seller,
		Title:     *title,
		Schema:    *schema,
		Price:     *price,
		ExpiresAt: *exp,
		Signature: hex.EncodeToString(sig[:]),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("LISTING_ID=%d\n", item.ID)
	fmt.Printf("SCHEMA=%s\n", item.Schema)
	fmt.Printf("SELLER_DID=%s\n", item.SellerDid)
	fmt.Printf("TITLE=%s\n", item.Title)
	fmt.Printf("PRICE=%d\n", item.Price)
}

func offer() {
	lid := flag.Int64("listing", 0, "listing id")
	price := flag.Int64("price", 0, "price")
	to := flag.String("to", "", "seller DID hex (optional; loaded from listing)")
	exp := flag.Int64("exp", time.Now().Unix()+1800, "unix expiry")
	flag.Parse()
	if *lid <= 0 || *price <= 0 {
		log.Fatal("offer needs -listing and -price")
	}
	from := did()
	seller := strings.TrimPrefix(*to, "0x")
	if seller == "" {
		ls, err := market().ListListings()
		if err != nil {
			log.Fatal(err)
		}
		for _, L := range ls {
			if L.ID == *lid {
				seller = strings.TrimPrefix(L.SellerDid, "0x")
				if *price == 0 {
					*price = L.Price
				}
				break
			}
		}
	}
	if seller == "" {
		log.Fatal("listing not found; pass -to DID")
	}
	msg := offersign.Message("create", *lid, from, *price, *exp)
	sig, _, err := sign(msg)
	if err != nil {
		log.Fatal(err)
	}
	item, err := market().CreateOffer(client.CreateOfferInput{
		ListingID: *lid,
		FromDid:   from,
		ToDid:     seller,
		Price:     *price,
		ExpiresAt: *exp,
		Signature: hex.EncodeToString(sig[:]),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("OFFER_ID=%d\n", item.ID)
	fmt.Printf("LISTING_ID=%d\n", item.ListingID)
	fmt.Printf("FROM_DID=%s\n", item.FromDid)
	fmt.Printf("TO_DID=%s\n", item.ToDid)
	fmt.Printf("PRICE=%d\n", item.Price)
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
		return [64]byte{}, [32]byte{}, fmt.Errorf("set SIGNER=alice|bob or ANNOUNCE_SEED")
	}
	sb, err := hex.DecodeString(seedHex)
	if err != nil || len(sb) != 32 {
		return [64]byte{}, [32]byte{}, fmt.Errorf("ANNOUNCE_SEED must be 32-byte hex")
	}
	var seed [32]byte
	copy(seed[:], sb)
	return announce.Sign(seed, msg)
}

func init() {
	_ = strconv.IntSize
}
