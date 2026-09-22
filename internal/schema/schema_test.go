package schema

import "testing"

func TestCanonical(t *testing.T) {
	ok, err := Canonical("csv.v1")
	if err != nil || ok != CSVv1 {
		t.Fatalf("csv.v1 should pass: %q %v", ok, err)
	}
	ok, err = Canonical("bytes.v1")
	if err != nil || ok != BytesV1 {
		t.Fatalf("bytes.v1 should pass: %q %v", ok, err)
	}
	ok, err = Canonical("api.v1")
	if err != nil || ok != APIv1 {
		t.Fatalf("api.v1 should pass: %q %v", ok, err)
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
	if _, err := Canonical("BYTES.V1"); err == nil {
		t.Fatal("bytes case should fail")
	}
}

func TestSame(t *testing.T) {
	if err := Same("csv.v1", "csv.v1"); err != nil {
		t.Fatal(err)
	}
	if err := Same("bytes.v1", "bytes.v1"); err != nil {
		t.Fatal(err)
	}
	if err := Same("api.v1", "api.v1"); err != nil {
		t.Fatal(err)
	}
	if err := Same("api.v1", "bytes.v1"); err == nil {
		t.Fatal("api vs bytes should fail")
	}
	ok, err := Canonical("job.v1")
	if err != nil || ok != JobV1 {
		t.Fatalf("job.v1 should pass: %q %v", ok, err)
	}
	if err := Same("job.v1", "api.v1"); err == nil {
		t.Fatal("job vs api should fail")
	}
	if err := Same("csv.v1", "bytes.v1"); err == nil {
		t.Fatal("csv vs bytes should fail")
	}
	if err := Same("csv.v1", ""); err == nil {
		t.Fatal("empty right should fail")
	}
}
