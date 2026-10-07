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
	"path/filepath"
	"strings"
	"time"

	"github.com/arthneura/arthneura-market/internal/announce"
	"github.com/arthneura/arthneura-market/pkg/client"
)

type row struct {
	ID        int64  `json:"id"`
	Price     int64  `json:"price"`
	ExpiresAt string `json:"expires_at"`
	Status    string `json:"status"`
}

func main() {
	id := flag.Int64("offer", 0, "offer id")
	flag.Parse()
	if *id == 0 {
		log.Fatal("need -offer")
	}
	base := strings.TrimRight(os.Getenv("MARKET_URL"), "/")
	if base == "" {
		base = "https://api.arthneura.com"
	}
	dir := strings.TrimSpace(os.Getenv("OWNER_DIR"))
	if dir == "" {
		home, _ := os.UserHomeDir()
		b, err := os.ReadFile(filepath.Join(home, ".arthneura", "owner.dir"))
		if err != nil {
			log.Fatal("OWNER_DIR unset and no ~/.arthneura/owner.dir")
		}
		dir = strings.TrimSpace(string(b))
	}
	did := strings.TrimPrefix(strings.TrimSpace(os.Getenv("OWNER_DID")), "0x")
	if did == "" {
		b, err := os.ReadFile(filepath.Join(dir, "owner.did"))
		if err != nil {
			log.Fatal(err)
		}
		did = strings.TrimPrefix(strings.TrimSpace(string(b)), "0x")
	}
	seedHex := strings.TrimSpace(os.Getenv("CONTROLLER_SEED"))
	if seedHex == "" {
		b, err := os.ReadFile(filepath.Join(dir, "controller.seed"))
		if err != nil {
			log.Fatal(err)
		}
		seedHex = strings.TrimSpace(string(b))
	}
	seedb, err := hex.DecodeString(seedHex)
	if err != nil || len(seedb) != 32 {
		log.Fatal("controller seed must be 32-byte hex")
	}
	var seed [32]byte
	copy(seed[:], seedb)

	res, err := http.Get(base + "/v1/offers")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var wrap struct {
		Offers []row `json:"offers"`
	}
	if err := json.NewDecoder(res.Body).Decode(&wrap); err != nil {
		log.Fatal(err)
	}
	var hit *row
	for i := range wrap.Offers {
		if wrap.Offers[i].ID == *id {
			hit = &wrap.Offers[i]
			break
		}
	}
	if hit == nil {
		log.Fatal("offer not in list")
	}
	exp, err := time.Parse(time.RFC3339, hit.ExpiresAt)
	if err != nil {
		log.Fatal(err)
	}
	msg := client.OfferSignMessage("accept", *id, did, hit.Price, exp.Unix())
	sig, _, err := announce.Sign(seed, msg)
	if err != nil {
		log.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{
		"did":        did,
		"price":      hit.Price,
		"expires_at": exp.Unix(),
		"signature":  hex.EncodeToString(sig[:]),
	})
	post, err := http.Post(fmt.Sprintf("%s/v1/offers/%d/accept", base, *id), "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}
	defer post.Body.Close()
	b, _ := io.ReadAll(post.Body)
	if post.StatusCode != 200 {
		log.Fatalf("accept %d: %s", post.StatusCode, strings.TrimSpace(string(b)))
	}
	fmt.Printf("ACCEPT offer=%d status_body=%s\n", *id, strings.TrimSpace(string(b)))
}
