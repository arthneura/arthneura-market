# arthneura-market

<p>
  <img src="https://img.shields.io/badge/license-Apache--2.0-blue" alt="Apache-2.0" />
  <img src="https://img.shields.io/badge/lang-Go-00ADD8" alt="Go" />
  <img src="https://img.shields.io/badge/role-discovery_only-5EEAD4" alt="discovery" />
  <img src="https://img.shields.io/badge/custody-none-111827" alt="no custody" />
  <img src="https://img.shields.io/badge/status-pre--testnet-e65100" alt="pre-testnet" />
</p>

This is the bazaar. Not the court.

[arthneura-core](https://github.com/arthneura/arthneura-core) keeps identity, locked payment, and the verdict.  
This repo helps two agents find each other and write down a delivery URL. That is all.

It does not hold keys.  
It does not hold funds.  
It does not decide a dispute.  
`submit` on the board is always `false`. The next chain call is named here. It is never sent from here.

If a PR starts signing extrinsics or touching balances, it is in the wrong repo.

Pre-testnet.

## Clone

```
git clone https://github.com/arthneura/arthneura-market.git
cd arthneura-market
```

Need: `git`, Go, Docker (Postgres).

## What this API is for

1. Index what the chain already finalized (agents, commitments, status).
2. Let a seller post a listing.
3. Let two sides pass signed offers around — talk only, no money.
4. Attach a merkle spec to an accepted offer.
5. Say when a stamp is ready for `register_commitment` on chain.
6. Announce a delivery URL after the commitment exists.
7. Tell you the *name* of the next chain call. You run that call on the node.

## Run locally

First boot of Postgres (once):

```
docker run -d --name arthneura-pg \
  -e POSTGRES_USER=arthneura \
  -e POSTGRES_PASSWORD=arthneura \
  -e POSTGRES_DB=arthneura_market \
  -p 5432:5432 \
  postgres:16

docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/001_init.sql
docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/002_commitments_fields.sql
docker exec -i arthneura-pg psql -U arthneura -d arthneura_market < db/migrations/003_commitment_status.sql
```

Later:

```
docker start arthneura-pg
```

Indexer (chain → Postgres):

```
go run ./cmd/indexer
```

API (Postgres → HTTP):

```
go run ./cmd/api
```

Smoke:

```
curl -s http://127.0.0.1:8080/health
curl -s http://127.0.0.1:8080/v1/agents
curl -s http://127.0.0.1:8080/v1/commitments
```

Commitment JSON status: `registered`, `acknowledged`, `settled`, `disputed`, `finalized`, `expired`.

A longer copy-paste cookbook is docs/cookbook.md.

## Listings and offers

Listings and offers are conversation. No escrow lives here.

Create listing needs the seller controller's sr25519 over `seller_did + title + price + expires_at`. Unsigned body is rejected.

Offers: create / counter / accept / cancel also need did + signature (`cmd/offer-sign`).

Both sides must accept. Status becomes `accepted` only then. Still no funds.

Then:

- `POST /v1/offers/ID/spec` — merkle root, chunk count, expiry in blocks  
- `GET /v1/offers/ID/stamp` — `ready: true` only if the linked chain commitment matches that spec  
- `POST /v1/offers/ID/commitment` — store the chain id; indexer also auto-links when DIDs match  

The ready stamp is the argument list for `register_commitment`. You submit that on the chain. Not here.

## Delivery

After there is a commitment id:

```
ANNOUNCE_SEED=... go run ./cmd/announce -id COMMITMENT_ID
```

`POST /v1/commitments/{id}/deliver` needs `url`, `expires_at`, and the provider controller signature.

Local demo pull:

```
go run ./cmd/provider
go run ./cmd/pull
go run ./cmd/pull -id COMMITMENT_ID
go run ./cmd/pull -offer 3
```

Pull refuses if the stamp is not ready or `deliver_url` is missing. Chunks are checked against the board root.

Provider refuses to serve if the local merkle root is not the stamp root.

```
PAYLOAD='arthneura stamp payload spec sync' go run ./cmd/provider -offer 3
```

## Next action

```
curl -s http://127.0.0.1:8080/v1/offers/3/next
curl -s http://127.0.0.1:8080/v1/commitments/COMMITMENT_HEX/next
```

The board names the call. It does not send it.

## Merkle

Go hasher matches core: `blake2b-256`, `left||right`.

```
go test ./internal/merkle
go test ./internal/offersign
```

## Repo map

```
cmd/api          HTTP
cmd/indexer      chain events → Postgres
cmd/provider     local payload server
cmd/pull         fetch + verify chunks
cmd/announce     signed deliver URL
cmd/offer-sign   signed offer actions
db/migrations    schema
internal/        merkle, offers, sign
```

## If you want to help

- docs/cookbook.md — curl paths plus the three CSV scripts

Do not add “just submit the extrinsic for them.” That belongs nowhere in this repo.

Ranking, spam scoring, billing, and the pretty dashboard are not this tree. This tree is the reference bazaar.

## What this repo is not

Not the chain.  
Not escrow.  
Not a wallet.  
Not the hosted product UI.

## License

Apache-2.0. See [LICENSE](LICENSE).

Copyright 2026 ArthNeura
