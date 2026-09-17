package main

import (
    "encoding/csv"
    "flag"
    "fmt"
    "os"
    "strings"
)

func main() {
    schema := flag.String("schema", "", "schema id, e.g. csv.v1")
    flag.Parse()
    if flag.NArg() < 1 {
        fmt.Fprintln(os.Stderr, "usage: csvcheck [-schema csv.v1] FILE")
        os.Exit(2)
    }
    if *schema != "" && *schema != "csv.v1" {
        fmt.Fprintf(os.Stderr, "ROW=-1 REASON=unknown_schema\n")
        os.Exit(1)
    }
    row, reason, err := CheckFile(flag.Arg(0))
    if err != nil {
        fmt.Fprintf(os.Stderr, "ROW=%d REASON=%s\n", row, reason)
        os.Exit(1)
    }
    fmt.Printf("ROW=-1 REASON=ok rows=%d\n", row)
}

func CheckFile(path string) (int, string, error) {
    f, err := os.Open(path)
    if err != nil {
        return -1, "open", err
    }
    defer f.Close()
    r := csv.NewReader(f)
    r.TrimLeadingSpace = true
    recs, err := r.ReadAll()
    if err != nil {
        return -1, "parse", err
    }
    return CheckRecords(recs)
}

func CheckRecords(recs [][]string) (int, string, error) {
    if len(recs) < 2 {
        return 0, "need_header_and_row", fmt.Errorf("need header and one row")
    }
    h := recs[0]
    if len(h) < 3 || h[0] != "company" || h[1] != "domain" || h[2] != "email" {
        return 0, "bad_header", fmt.Errorf("header must be company,domain,email")
    }
    for i, rec := range recs[1:] {
        row := i + 1
        if len(rec) < 3 {
            return row, "col_count", fmt.Errorf("row %d missing columns", row)
        }
        line := strings.Join(rec, ",")
        if len(line) > 1024 {
            return row, "row_too_long", fmt.Errorf("row %d too long", row)
        }
        company, domain, email := strings.TrimSpace(rec[0]), strings.TrimSpace(rec[1]), strings.TrimSpace(rec[2])
        if company == "" {
            return row, "empty_company", fmt.Errorf("row %d empty company", row)
        }
        if domain == "" || strings.Contains(domain, "://") {
            return row, "bad_domain", fmt.Errorf("row %d bad domain", row)
        }
        if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
            return row, "bad_email", fmt.Errorf("row %d bad email", row)
        }
    }
    return len(recs) - 1, "ok", nil
}
