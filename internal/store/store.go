package store

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
    pool *pgxpool.Pool
}

func Open(ctx context.Context, dsn string) (*Store, error) {
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        return nil, fmt.Errorf("postgres: %w", err)
    }
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("postgres ping: %w", err)
    }
    return &Store{pool: pool}, nil
}

func (s *Store) Close() {
    if s != nil && s.pool != nil {
        s.pool.Close()
    }
}

func (s *Store) UpsertAgent(ctx context.Context, did, controller []byte, block uint64) error {
    _, err := s.pool.Exec(ctx, `
        INSERT INTO agents (did, controller, raw_event)
        VALUES ($1, $2, jsonb_build_object('block', $3::bigint))
        ON CONFLICT (did) DO UPDATE SET
            controller = EXCLUDED.controller,
            raw_event = EXCLUDED.raw_event,
            updated_at = now()
    `, did, fmt.Sprintf("%x", controller), block)
    return err
}
