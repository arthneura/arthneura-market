package schema

import (
	"fmt"
	"strings"
)

const CSVv1 = "csv.v1"
const BytesV1 = "bytes.v1"

func Canonical(s string) (string, error) {
	id := strings.TrimSpace(s)
	if id == "" {
		return "", fmt.Errorf("schema missing")
	}
	switch id {
	case CSVv1, BytesV1:
		return id, nil
	default:
		return "", fmt.Errorf("unknown schema %q", id)
	}
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
