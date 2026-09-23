package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListingSignMessageRejectsUnknownSchema(t *testing.T) {
	_, err := ListingSignMessage(CreateListingInput{
		SellerDid: "aa",
		Title:     "t",
		Schema:    "nope.v1",
	})
	if err == nil {
		t.Fatal("unknown schema must fail")
	}
}

func TestListingSignMessageAllowsJob(t *testing.T) {
	msg, err := ListingSignMessage(CreateListingInput{
		SellerDid: "aa",
		Title:     "t",
		Schema:    "job.v1",
		Price:     1,
		ExpiresAt: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(msg) != "arthneura-listing|aa|t|1|2|job.v1" {
		t.Fatalf("msg=%q", msg)
	}
}

func TestStampHTTP(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/offers/3/stamp" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Stamp{OfferID: 3, Ready: false, Reason: "commitment not linked", Schema: "csv.v1"})
	}))
	defer s.Close()
	st, err := New(s.URL).Stamp(3)
	if err != nil {
		t.Fatal(err)
	}
	if st.Ready || st.Schema != "csv.v1" {
		t.Fatalf("%+v", st)
	}
}
