# arthneura-market

Off-chain marketplace for ArthNeura.

arthneura-core is the court. This repo is the bazaar.
Does not custody keys or funds. Does not decide disputes.

Status: pre-testnet.

## Run locally

Postgres:

    docker start arthneura-pg

Indexer (writes chain events and commitment status to Postgres):

    go run ./cmd/indexer

Discovery API (reads Postgres):

    go run ./cmd/api
    curl -s http://127.0.0.1:8080/health
    curl -s http://127.0.0.1:8080/v1/agents
    curl -s http://127.0.0.1:8080/v1/commitments

Commitment JSON includes status: registered, acknowledged, settled, disputed, finalized, expired.

First boot of Postgres (only once):

    docker run -d --name arthneura-pg \
      -e POSTGRES_USER=arthneura \
      -e POSTGRES_PASSWORD=arthneura \
      -e POSTGRES_DB=arthneura_market \
      -p 5432:5432 \
      postgres:16
    docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/001_init.sql
    docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/002_commitments_fields.sql
    docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/003_commitment_status.sql

## Merkle

Go hasher matches arthneura-core (`blake2b-256`, left||right).

    go test ./internal/merkle

## Local chunk pull (demo)

    go run ./cmd/provider
    go run ./cmd/pull

## Announce a delivery URL

    curl -s -X POST http://127.0.0.1:8080/v1/commitments/COMMITMENT_ID/deliver \
      -H 'Content-Type: application/json' \
      -d '{"url":"http://127.0.0.1:8090"}'

## Pull from the board

    go run ./cmd/api
    go run ./cmd/provider
    go run ./cmd/pull -id COMMITMENT_ID

## Signed deliver announce

POST /v1/commitments/{id}/deliver now needs url, expires_at, and an sr25519
signature by the provider agent's controller.

    ANNOUNCE_SEED=... go run ./cmd/announce -id COMMITMENT_ID

## Agent profile fields

GET /v1/agents includes capabilities, status, label.
Status updates from AgentStatusChanged.

## Listings

    curl -s http://127.0.0.1:8080/v1/listings
    curl -s -X POST http://127.0.0.1:8080/v1/listings \
      -H 'Content-Type: application/json' \
      -d '{"seller_did":"DID_HEX","title":"street photos v1","price":80}'
