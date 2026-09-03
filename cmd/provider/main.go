package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strconv"

    "github.com/arthneura/arthneura-market/internal/deliver"
    "github.com/arthneura/arthneura-market/internal/merkle"
)

func main() {
    addr := os.Getenv("DELIVER_ADDR")
    if addr == "" {
        addr = "127.0.0.1:8090"
    }
    chunks := [][]byte{[]byte("hello"), []byte("world"), []byte("arth")}
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

func hex32(h [32]byte) string {
    const d = "0123456789abcdef"
    var s [64]byte
    for i, b := range h {
        s[i*2] = d[b>>4]
        s[i*2+1] = d[b&0xf]
    }
    return string(s[:])
}
