package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/arthneura/arthneura-market/internal/announce"
	"github.com/arthneura/arthneura-market/pkg/client"
)

type rule struct {
	Floor int64  `json:"floor"`
	Pick  string `json:"pick"`
}

type offer struct {
	ID        int64  `json:"id"`
	ListingID int64  `json:"listing_id"`
	ToDid     string `json:"to_did"`
	Price     int64  `json:"price"`
	ExpiresAt string `json:"expires_at"`
	Status    string `json:"status"`
}

func main() {
	log.SetFlags(0)
	dir := os.Getenv("OWNER_DIR")
	did := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	seedHex := strings.TrimPrefix(strings.TrimSpace(os.Getenv("CONTROLLER_SEED")), "0x")
	sb, err := hex.DecodeString(seedHex)
	if dir == "" || did == "" || err != nil || len(sb) != 32 {
		log.Fatal("OWNER_DIR, OWNER_DID, CONTROLLER_SEED required")
	}
	var seed [32]byte
	copy(seed[:], sb)
	raw, err := os.ReadFile(dir + "/rules.json")
	if err != nil {
		log.Fatal(err)
	}
	rules := map[string]rule{}
	if err := json.Unmarshal(raw, &rules); err != nil {
		log.Fatal(err)
	}
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	u = strings.TrimRight(u, "/")
	res, err := http.Get(u + "/v1/offers")
	if err != nil {
		log.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	var wrap struct {
		Offers []offer `json:"offers"`
	}
	if err := json.Unmarshal(body, &wrap); err != nil {
		log.Fatal(err)
	}
	picked := 0
	seen := map[int64]bool{}
	for lid, r := range rules {
		var best *offer
		for i := range wrap.Offers {
			o := &wrap.Offers[i]
			if fmt.Sprint(o.ListingID) != lid || o.Status != "open" || o.Price < r.Floor {
				continue
			}
			if !strings.EqualFold(o.ToDid, did) {
				continue
			}
			if best == nil || o.Price > best.Price {
				best = o
			}
		}
		if best == nil || seen[best.ID] {
			continue
		}
		seen[best.ID] = true
		exp, err := time.Parse(time.RFC3339, best.ExpiresAt)
		if err != nil {
			log.Fatal(err)
		}
		msg := client.OfferSignMessage("accept", best.ID, did, best.Price, exp.Unix())
		sig, _, err := announce.Sign(seed, msg)
		if err != nil {
			log.Fatal(err)
		}
		payload, _ := json.Marshal(map[string]any{
			"did": did, "price": best.Price, "expires_at": exp.Unix(),
			"signature": hex.EncodeToString(sig[:]),
		})
		post, err := http.Post(fmt.Sprintf("%s/v1/offers/%d/accept", u, best.ID), "application/json", bytes.NewReader(payload))
		if err != nil {
			log.Fatal(err)
		}
		b, _ := io.ReadAll(post.Body)
		post.Body.Close()
		if post.StatusCode != 200 {
			log.Fatalf("accept %d: %s", post.StatusCode, strings.TrimSpace(string(b)))
		}
		fmt.Printf("DECIDE listing=%d offer=%d price=%d\n", best.ListingID, best.ID, best.Price)
		picked++
	}
	if picked == 0 {
		fmt.Println("DECIDE none")
	}
}
