package main

import "testing"

func TestGood(t *testing.T) {
    n, reason, err := CheckFile("../../testdata/csv/good.csv")
    if err != nil || reason != "ok" || n != 2 {
        t.Fatalf("good: n=%d reason=%s err=%v", n, reason, err)
    }
}

func TestBadEmail(t *testing.T) {
    row, reason, err := CheckFile("../../testdata/csv/bad-email.csv")
    if err == nil || row != 1 || reason != "bad_email" {
        t.Fatalf("bad email: row=%d reason=%s err=%v", row, reason, err)
    }
}

func TestBadHeader(t *testing.T) {
    row, reason, err := CheckFile("../../testdata/csv/bad-header.csv")
    if err == nil || row != 0 || reason != "bad_header" {
        t.Fatalf("bad header: row=%d reason=%s err=%v", row, reason, err)
    }
}
