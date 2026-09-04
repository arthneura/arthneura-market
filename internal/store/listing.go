package store

import (
    "context"
    "encoding/hex"
    "fmt"
)

type Listing struct {
    ID        int64  `json:"id"`
    SellerDid string `json:"seller_did"`
    Title     string `json:"title"`
    Price     int64  `json:"price"`
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

func (s *Store) ListListings(ctx context.Context) ([]Listing, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT id, seller_did, title, price
        FROM listings
        ORDER BY id DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Listing
    for rows.Next() {
        var id, price int64
        var did []byte
        var title string
        if err := rows.Scan(&id, &did, &title, &price); err != nil {
            return nil, err
        }
        out = append(out, Listing{
            ID:        id,
            SellerDid: hex.EncodeToString(did),
            Title:     title,
            Price:     price,
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
