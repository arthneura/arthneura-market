package store

import (
    "context"
    "encoding/hex"
    "time"
)

type Offer struct {
    ID        int64  `json:"id"`
    ListingID int64  `json:"listing_id"`
    FromDid   string `json:"from_did"`
    ToDid     string `json:"to_did"`
    Price     int64  `json:"price"`
    ExpiresAt string `json:"expires_at"`
    Status    string `json:"status"`
}

func (s *Store) CreateOffer(ctx context.Context, listingID int64, from, to []byte, price int64, exp time.Time) (Offer, error) {
    var id int64
    err := s.pool.QueryRow(ctx, `
        INSERT INTO offers (listing_id, from_did, to_did, price, expires_at, status)
        VALUES ($1,$2,$3,$4,$5,'open')
        RETURNING id
    `, listingID, from, to, price, exp).Scan(&id)
    if err != nil {
        return Offer{}, err
    }
    return Offer{
        ID:        id,
        ListingID: listingID,
        FromDid:   hex.EncodeToString(from),
        ToDid:     hex.EncodeToString(to),
        Price:     price,
        ExpiresAt: exp.UTC().Format(time.RFC3339),
        Status:    "open",
    }, nil
}

func (s *Store) CancelOffer(ctx context.Context, id int64) error {
    _, err := s.pool.Exec(ctx, `
        UPDATE offers SET status = 'cancelled'
        WHERE id = $1 AND status = 'open'
    `, id)
    return err
}

func (s *Store) ListOffers(ctx context.Context) ([]Offer, error) {
    _, _ = s.pool.Exec(ctx, `
        UPDATE offers SET status = 'expired'
        WHERE status = 'open' AND expires_at < now()
    `)
    rows, err := s.pool.Query(ctx, `
        SELECT id, listing_id, from_did, to_did, price, expires_at, status
        FROM offers
        ORDER BY id DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Offer
    for rows.Next() {
        var id, listingID, price int64
        var from, to []byte
        var exp time.Time
        var status string
        if err := rows.Scan(&id, &listingID, &from, &to, &price, &exp, &status); err != nil {
            return nil, err
        }
        out = append(out, Offer{
            ID:        id,
            ListingID: listingID,
            FromDid:   hex.EncodeToString(from),
            ToDid:     hex.EncodeToString(to),
            Price:     price,
            ExpiresAt: exp.UTC().Format(time.RFC3339),
            Status:    status,
        })
    }
    return out, rows.Err()
}
