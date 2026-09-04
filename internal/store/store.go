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
    Did          string `json:"did"`
    Controller   string `json:"controller"`
    Block        int64  `json:"block"`
    Capabilities int64  `json:"capabilities"`
    Status       string `json:"status"`
    Label        string `json:"label"`
}

type Commitment struct {
    CommitmentID string `json:"commitment_id"`
    Provider     string `json:"provider"`
    Consumer     string `json:"consumer"`
    MerkleRoot   string `json:"merkle_root"`
    TotalChunks  int64  `json:"total_chunks"`
    ExpiresAt    int64  `json:"expires_at"`
    Block        int64  `json:"block"`
    Status       string `json:"status"`
    DeliverURL   string `json:"deliver_url,omitempty"`
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
        INSERT INTO agents (did, controller, raw_event, status)
        VALUES ($1, $2, jsonb_build_object('block', $3::bigint), 'active')
        ON CONFLICT (did) DO UPDATE SET
            controller = EXCLUDED.controller,
            raw_event = EXCLUDED.raw_event,
            updated_at = now()
    `, did, fmt.Sprintf("%x", controller), block)
    return err
}

func (s *Store) SetAgentStatus(ctx context.Context, did []byte, status string) error {
    _, err := s.pool.Exec(ctx, `
        UPDATE agents SET status = $2, updated_at = now() WHERE did = $1
    `, did, status)
    return err
}

func (s *Store) SetAgentProfile(ctx context.Context, did []byte, capabilities int64, label string) error {
    _, err := s.pool.Exec(ctx, `
        UPDATE agents SET capabilities = $2, label = $3, updated_at = now() WHERE did = $1
    `, did, capabilities, label)
    return err
}

func (s *Store) ListAgents(ctx context.Context) ([]Agent, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT did, controller, COALESCE((raw_event->>'block')::bigint, 0),
               COALESCE(capabilities,0), COALESCE(status,'active'), COALESCE(label,'')
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
        SELECT did, controller, COALESCE((raw_event->>'block')::bigint, 0),
               COALESCE(capabilities,0), COALESCE(status,'active'), COALESCE(label,'')
        FROM agents
        WHERE did = $1
    `, raw)
    return scanAgent(row)
}

func (s *Store) UpsertCommitment(ctx context.Context, id, provider, consumer, root []byte, chunks, expires, block uint64) error {
    _, err := s.pool.Exec(ctx, `
        INSERT INTO commitments (
            commitment_id, provider, consumer, merkle_root,
            total_chunks, expires_at, block, status, raw_event
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,'registered', jsonb_build_object('block', $7::bigint))
        ON CONFLICT (commitment_id) DO UPDATE SET
            provider = EXCLUDED.provider,
            consumer = EXCLUDED.consumer,
            merkle_root = EXCLUDED.merkle_root,
            total_chunks = EXCLUDED.total_chunks,
            expires_at = EXCLUDED.expires_at,
            block = EXCLUDED.block,
            raw_event = EXCLUDED.raw_event,
            updated_at = now()
    `, id, provider, consumer, root, int64(chunks), int64(expires), int64(block))
    return err
}

func (s *Store) SetCommitmentStatus(ctx context.Context, id []byte, status string) error {
    _, err := s.pool.Exec(ctx, `
        UPDATE commitments SET status = $2, updated_at = now()
        WHERE commitment_id = $1
    `, id, status)
    return err
}

func (s *Store) SetDeliverURL(ctx context.Context, id []byte, url string) error {
    _, err := s.pool.Exec(ctx, `
        UPDATE commitments SET deliver_url = $2, updated_at = now()
        WHERE commitment_id = $1
    `, id, url)
    return err
}

func (s *Store) ListCommitments(ctx context.Context) ([]Commitment, error) {
    rows, err := s.pool.Query(ctx, `
        SELECT commitment_id, provider, consumer, merkle_root,
               COALESCE(total_chunks,0), COALESCE(expires_at,0), COALESCE(block,0),
               COALESCE(status, 'registered'), COALESCE(deliver_url, '')
        FROM commitments
        ORDER BY updated_at DESC
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var out []Commitment
    for rows.Next() {
        c, err := scanCommitment(rows)
        if err != nil {
            return nil, err
        }
        out = append(out, c)
    }
    return out, rows.Err()
}

func (s *Store) GetCommitment(ctx context.Context, idHex string) (Commitment, error) {
    raw, err := hex.DecodeString(idHex)
    if err != nil {
        return Commitment{}, fmt.Errorf("bad commitment id hex: %w", err)
    }
    row := s.pool.QueryRow(ctx, `
        SELECT commitment_id, provider, consumer, merkle_root,
               COALESCE(total_chunks,0), COALESCE(expires_at,0), COALESCE(block,0),
               COALESCE(status, 'registered'), COALESCE(deliver_url, '')
        FROM commitments
        WHERE commitment_id = $1
    `, raw)
    return scanCommitment(row)
}

type scanner interface {
    Scan(dest ...any) error
}

func scanAgent(row scanner) (Agent, error) {
    var did []byte
    var controller string
    var block, caps int64
    var status, label string
    if err := row.Scan(&did, &controller, &block, &caps, &status, &label); err != nil {
        if err == pgx.ErrNoRows {
            return Agent{}, err
        }
        return Agent{}, err
    }
    return Agent{
        Did:          hex.EncodeToString(did),
        Controller:   controller,
        Block:        block,
        Capabilities: caps,
        Status:       status,
        Label:        label,
    }, nil
}

func scanCommitment(row scanner) (Commitment, error) {
    var id, provider, consumer, root []byte
    var chunks, expires, block int64
    var status, deliver string
    if err := row.Scan(&id, &provider, &consumer, &root, &chunks, &expires, &block, &status, &deliver); err != nil {
        if err == pgx.ErrNoRows {
            return Commitment{}, err
        }
        return Commitment{}, err
    }
    return Commitment{
        CommitmentID: hex.EncodeToString(id),
        Provider:     hex.EncodeToString(provider),
        Consumer:     hex.EncodeToString(consumer),
        MerkleRoot:   hex.EncodeToString(root),
        TotalChunks:  chunks,
        ExpiresAt:    expires,
        Block:        block,
        Status:       status,
        DeliverURL:   deliver,
    }, nil
}
