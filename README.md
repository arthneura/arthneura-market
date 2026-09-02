# arthneura-market

Off-chain marketplace for ArthNeura.

arthneura-core is the court. This repo is the bazaar.
Does not custody keys or funds. Does not decide disputes.

Status: pre-testnet.

## Run locally

Postgres:

    docker start arthneura-pg

Indexer (writes chain events to Postgres):

    go run ./cmd/indexer

Discovery API (reads Postgres):

    go run ./cmd/api
    curl -s http://127.0.0.1:8080/health
    curl -s http://127.0.0.1:8080/v1/agents

First boot of Postgres (only once):

    docker run -d --name arthneura-pg \
      -e POSTGRES_USER=arthneura \
      -e POSTGRES_PASSWORD=arthneura \
      -e POSTGRES_DB=arthneura_market \
      -p 5432:5432 \
      postgres:16
    docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/001_init.sql
