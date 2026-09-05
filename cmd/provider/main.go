package main

import (
    "encoding/hex"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    "github.com/arthneura/arthneura-market/internal/deliver"
    "github.com/arthneura/arthneura-market/internal/merkle"
)

type stamp struct {
    Ready        bool   `json:"ready"`
    Reason       string `json:"reason"`
    MerkleRoot   string `json:"merkle_root"`
    TotalChunks  int64  `json:"total_chunks"`
    CommitmentID string `json:"commitment_id"`
}

func main() {
    offer := flag.Int64("offer", 0, "serve chunks that match this offer stamp")
    flag.Parse()

    addr := os.Getenv("DELIVER_ADDR")
    if addr == "" {
        addr = "127.0.0.1:8090"
    }

    chunks := [][]byte{[]byte("hello"), []byte("world"), []byte("arth")}
    if *offer > 0 {
        market := os.Getenv("MARKET_URL")
        if market == "" {
            market = "http://127.0.0.1:8080"
        }
        var s stamp
        if err := getJSON(fmt.Sprintf("%s/v1/offers/%d/stamp", market, *offer), &s); err != nil {
            log.Fatal(err)
        }
        if !s.Ready {
            log.Fatalf("stamp not ready: %s", s.Reason)
        }
        payload := []byte(os.Getenv("PAYLOAD"))
        if len(payload) == 0 {
            log.Fatal("PAYLOAD env required when -offer is set")
        }
        n := int(s.TotalChunks)
        if n <= 0 {
            log.Fatal("stamp total_chunks missing")
        }
        chunks = splitPayload(payload, n)
        root := merkle.Root(chunks)
        want, err := hex.DecodeString(s.MerkleRoot)
        if err != nil || len(want) != 32 {
            log.Fatal("bad stamp merkle_root")
        }
        if hex.EncodeToString(root[:]) != s.MerkleRoot {
            log.Fatalf("local payload root=%x does not match stamp root=%s — refusing to serve", root, s.MerkleRoot)
        }
        log.Printf("serving offer=%d commitment=%s", *offer, s.CommitmentID)
    }

    root := merkle.Root(chunks)
    log.Printf("provider on %s root=%x total=%d", addr, root, len(chunks))

    mux := http.NewServeMux()
    mux.HandleFunc("GET /root", func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(map[string]any{
            "root":  hex32(root),
            "total": len(chunks),
        })
    })
    mux.HandleFunc("GET /chunks/{i}", func(w http.ResponseWriter, r *http.Request) {
        i, err := strconv.Atoi(r.PathValue("i"))
        if err != nil || i < 0 || i >= len(chunks) {
            http.Error(w, "chunk not found", http.StatusNotFound)
            return
        }
        box := deliver.EncodeBox(i, chunks[i], merkle.Proof(chunks, i))
        w.Header().Set("Content-Type", "application/json")
        _ = json.NewEncoder(w).Encode(box)
    })
    log.Fatal(http.ListenAndServe(addr, mux))
}

func splitPayload(payload []byte, n int) [][]byte {
    if n == 1 {
        return [][]byte{payload}
    }
    out := make([][]byte, n)
    size := (len(payload) + n - 1) / n
    for i := 0; i < n; i++ {
        start := i * size
        if start >= len(payload) {
            out[i] = []byte{}
            continue
        }
        end := start + size
        if end > len(payload) {
            end = len(payload)
        }
        out[i] = payload[start:end]
    }
    return out
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

func hex32(h [32]byte) string {
    const d = "0123456789abcdef"
    var s [64]byte
    for i, b := range h {
        s[i*2] = d[b>>4]
        s[i*2+1] = d[b&0xf]
    }
    return string(s[:])
}
