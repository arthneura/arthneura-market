package main

import (
    "encoding/hex"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/arthneura/arthneura-market/internal/deliver"
)

type rootResp struct {
    Root  string `json:"root"`
    Total int    `json:"total"`
}

func main() {
    base := os.Getenv("PROVIDER_URL")
    if base == "" {
        base = "http://127.0.0.1:8090"
    }

    var rr rootResp
    if err := getJSON(base+"/root", &rr); err != nil {
        log.Fatal(err)
    }
    rootb, err := hex.DecodeString(rr.Root)
    if err != nil || len(rootb) != 32 {
        log.Fatal("bad root")
    }
    var root [32]byte
    copy(root[:], rootb)
    log.Printf("pull from %s total=%d root=%s", base, rr.Total, rr.Root)

    for i := 0; i < rr.Total; i++ {
        var box deliver.ChunkBox
        if err := getJSON(fmt.Sprintf("%s/chunks/%d", base, i), &box); err != nil {
            log.Fatal(err)
        }
        if !deliver.VerifyBox(box, rr.Total, root) {
            log.Fatalf("chunk %d VERIFY FAIL", i)
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
