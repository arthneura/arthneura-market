package store

import (
    "context"
    "encoding/hex"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
    pool *pgxpool.Pool
}

type Agent struct {
    Did        string `json:"did"`
    Controller string `json:"controller"`
    Block      int64  `json:"block"`
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

func (s *Store) ListAgents(ctx context.Context) ([]Agent, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT did, controller, COALESCE((raw_event->>'block')::bigint, 0)
        FROM agents
        ORDER BY updated_at DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Agent
    for rows.Next() {
        a, err := scanAgent(rows)
        if err != nil {
            return nil, err
        }
        out = append(out, a)
    }
    return out, rows.Err()
}

func (s *Store) GetAgent(ctx context.Context, didHex string) (Agent, error) {
    raw, err := hex.DecodeString(didHex)
    if err != nil {
        return Agent{}, fmt.Errorf("bad did hex: %w", err)
    }
    row := s.pool.QueryRow(ctx, `
        SELECT did, controller, COALESCE((raw_event->>'block')::bigint, 0)
        FROM agents
        WHERE did = $1
    `, raw)
    return scanAgent(row)
}

type scanner interface {
    Scan(dest ...any) error
}

func scanAgent(row scanner) (Agent, error) {
    var did []byte
    var controller string
    var block int64
    if err := row.Scan(&did, &controller, &block); err != nil {
        if err == pgx.ErrNoRows {
            return Agent{}, err
        }
        return Agent{}, err
    }
    return Agent{
        Did:        hex.EncodeToString(did),
        Controller: controller,
        Block:      block,
    }, nil
}
