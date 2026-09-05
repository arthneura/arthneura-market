package store

import (
    "context"
    "encoding/hex"
    "fmt"
)

type Listing struct {
    ID           int64  `json:"id"`
    SellerDid    string `json:"seller_did"`
    Title        string `json:"title"`
    Price        int64  `json:"price"`
    SellerStatus string `json:"seller_status,omitempty"`
    Capabilities int64  `json:"capabilities,omitempty"`
}

func (s *Store) CreateListing(ctx context.Context, seller []byte, title string, price int64) (Listing, error) {
    var id int64
    err := s.pool.QueryRow(ctx, `
        INSERT INTO listings (seller_did, title, price)
        VALUES ($1, $2, $3)
        RETURNING id
    `, seller, title, price).Scan(&id)
    if err != nil {
        return Listing{}, err
    }
    return Listing{
        ID:        id,
        SellerDid: hex.EncodeToString(seller),
        Title:     title,
        Price:     price,
    }, nil
}

func (s *Store) ListListings(ctx context.Context, status string, cap int64) ([]Listing, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT l.id, l.seller_did, l.title, l.price,
               COALESCE(a.status, 'active'), COALESCE(a.capabilities, 0)
        FROM listings l
        LEFT JOIN agents a ON a.did = l.seller_did
        WHERE ($1 = '' OR COALESCE(a.status, 'active') = $1)
          AND ($2 = 0 OR (COALESCE(a.capabilities, 0) & $2) = $2)
        ORDER BY l.id DESC
    `, status, cap)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Listing
    for rows.Next() {
        var id, price, caps int64
        var did []byte
        var title, st string
        if err := rows.Scan(&id, &did, &title, &price, &st, &caps); err != nil {
            return nil, err
        }
        out = append(out, Listing{
            ID:           id,
            SellerDid:    hex.EncodeToString(did),
            Title:        title,
            Price:        price,
            SellerStatus: st,
            Capabilities: caps,
        })
    }
    return out, rows.Err()
}

func DecodeDid(h string) ([]byte, error) {
    b, err := hex.DecodeString(h)
    if err != nil || len(b) != 32 {
        return nil, fmt.Errorf("bad did hex")
    }
    return b, nil
}

func (s *Store) GetListing(ctx context.Context, id int64) (Listing, error) {
    var seller []byte
    var title string
    var price int64
    err := s.pool.QueryRow(ctx, `
        SELECT id, seller_did, title, price FROM listings WHERE id = $1
    `, id).Scan(&id, &seller, &title, &price)
    if err != nil {
        return Listing{}, err
    }
    return Listing{ID: id, SellerDid: hex.EncodeToString(seller), Title: title, Price: price}, nil
}
