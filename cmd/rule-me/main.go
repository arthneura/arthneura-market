package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type rule struct {
	Floor int64  `json:"floor"`
	Pick  string `json:"pick"`
}

func main() {
	log.SetFlags(0)
	lid := flag.Int64("listing", 0, "listing id")
	floor := flag.Int64("floor", 0, "reject below this price")
	flag.Parse()
	if *lid <= 0 || *floor <= 0 {
		log.Fatal("need -listing and -floor")
	}
	dir := strings.TrimSpace(os.Getenv("OWNER_DIR"))
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		b, err := os.ReadFile(filepath.Join(home, ".arthneura", "owner.dir"))
		if err != nil {
			log.Fatal("OWNER_DIR required")
		}
		dir = strings.TrimSpace(string(b))
	}
	if dir == "" {
		log.Fatal("OWNER_DIR required")
	}
	path := filepath.Join(dir, "rules.json")
	rules := map[string]rule{}
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &rules)
	}
	rules[fmt.Sprint(*lid)] = rule{Floor: *floor, Pick: "highest"}
	b, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0600); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("RULE listing=%d floor=%d pick=highest\n", *lid, *floor)
	fmt.Printf("FILE=%s\n", path)
}
