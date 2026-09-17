package schema

import (
    "fmt"
    "strings"
)

const CSVv1 = "csv.v1"

func Canonical(s string) (string, error) {
    id := strings.TrimSpace(s)
    if id == "" {
        return "", fmt.Errorf("schema missing")
    }
    if id != CSVv1 {
        return "", fmt.Errorf("schema not csv.v1")
    }
    return id, nil
}

func Same(a, b string) error {
    left, err := Canonical(a)
    if err != nil {
        return err
    }
    right, err := Canonical(b)
    if err != nil {
        return err
    }
    if left != right {
        return fmt.Errorf("listing/offer schema mismatch")
    }
    return nil
}
