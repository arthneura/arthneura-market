package client

import (
	"encoding/hex"
	"fmt"

	"github.com/arthneura/arthneura-market/internal/offersign"
	"github.com/arthneura/arthneura-market/internal/schema"
)

type Listing struct {
	ID        int64  `json:"id"`
	SellerDid string `json:"seller_did"`
	Title     string `json:"title"`
	Schema    string `json:"schema"`
	Price     int64  `json:"price"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

type CreateListingInput struct {
	SellerDid string
	Title     string
	Schema    string
	Price     int64
	ExpiresAt int64
	Signature string // 64-byte sr25519 hex, same bytes offer-sign prints
}

func ListingSignMessage(in CreateListingInput) ([]byte, error) {
	id, err := schema.Canonical(in.Schema)
	if err != nil {
		return nil, err
	}
	if in.SellerDid == "" || in.Title == "" {
		return nil, fmt.Errorf("seller_did and title required")
	}
	return offersign.ListingMessage(in.SellerDid, in.Title, in.Price, in.ExpiresAt, id), nil
}

func (c *Client) CreateListing(in CreateListingInput) (Listing, error) {
	id, err := schema.Canonical(in.Schema)
	if err != nil {
		return Listing{}, err
	}
	if _, err := hex.DecodeString(in.Signature); err != nil || in.Signature == "" {
		return Listing{}, fmt.Errorf("signature must be hex")
	}
	var out Listing
	err = c.doJSON("POST", "/v1/listings", map[string]any{
		"seller_did": in.SellerDid,
		"title":      in.Title,
		"schema":     id,
		"price":      in.Price,
		"expires_at": in.ExpiresAt,
		"signature":  in.Signature,
	}, &out)
	return out, err
}
