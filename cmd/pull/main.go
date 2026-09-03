package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"

    "github.com/arthneura/arthneura-market/internal/deliver"
)

type commitment struct {
    CommitmentID string `json:"commitment_id"`
    MerkleRoot   string `json:"merkle_root"`
    TotalChunks  int64  `json:"total_chunks"`
    DeliverURL   string `json:"deliver_url"`
    Status       string `json:"status"`
}

func main() {
    id := flag.String("id", os.Getenv("COMMITMENT_ID"), "commitment id hex")
    flag.Parse()
    market := os.Getenv("MARKET_URL")
    if market == "" {
        market = "http://127.0.0.1:8080"
    }
    *id = strings.TrimSpace(*id)
    if *id == "" {
        log.Fatal("pass -id COMMITMENT_HEX or COMMITMENT_ID")
    }

    var c commitment
    if err := getJSON(market+"/v1/commitments/"+*id, &c); err != nil {
        log.Fatal(err)
    }
    if strings.TrimSpace(c.DeliverURL) == "" {
        log.Fatal("no deliver_url on this commitment — provider must announce first")
    }
    rootb, err := hex.DecodeString(c.MerkleRoot)
    if err != nil || len(rootb) != 32 {
        log.Fatal("bad merkle_root on board")
    }
    var root [32]byte
    copy(root[:], rootb)
    total := int(c.TotalChunks)
    if total <= 0 {
        log.Fatal("total_chunks missing on board")
    }
    base := strings.TrimRight(c.DeliverURL, "/")
    log.Printf("id=%s status=%s url=%s total=%d", c.CommitmentID, c.Status, base, total)

    for i := 0; i < total; i++ {
        var box deliver.ChunkBox
        if err := getJSON(fmt.Sprintf("%s/chunks/%d", base, i), &box); err != nil {
            log.Fatal(err)
        }
        if !deliver.VerifyBox(box, total, root) {
            log.Fatalf("chunk %d VERIFY FAIL against board root", i)
        }
        log.Printf("chunk %d ok", i)
    }
    log.Printf("all chunks verified")
}

func getJSON(url string, dest any) error {
    res, err := http.Get(url)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        return fmt.Errorf("%s -> %s", url, res.Status)
    }
    return json.NewDecoder(res.Body).Decode(dest)
}
