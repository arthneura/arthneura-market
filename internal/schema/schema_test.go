package schema

import "testing"

func TestCanonical(t *testing.T) {
    ok, err := Canonical("csv.v1")
    if err != nil || ok != CSVv1 {
        t.Fatalf("csv.v1 should pass: %q %v", ok, err)
    }
    if _, err := Canonical(" csv.v1 "); err != nil {
        t.Fatalf("trim should pass: %v", err)
    }
    if _, err := Canonical(""); err == nil {
        t.Fatal("empty should fail")
    }
    if _, err := Canonical("CSV.V1"); err == nil {
        t.Fatal("case should fail")
    }
    if _, err := Canonical("csv.v2"); err == nil {
        t.Fatal("v2 should fail")
    }
}

func TestSame(t *testing.T) {
    if err := Same("csv.v1", "csv.v1"); err != nil {
        t.Fatal(err)
    }
    if err := Same("csv.v1", ""); err == nil {
        t.Fatal("empty right should fail")
    }
}
