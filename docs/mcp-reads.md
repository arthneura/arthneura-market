# MCP reads. Owner signs.

MCP tools are read-only.

They may: health, list listings, list offers, stamp, get commitment.

They may not: create a listing, accept an offer, register_commitment,
lock, raise, close, hold a key.

The owner signs on their machine (`cmd/offer-sign`). The market verifies
the signature. The chain is still the court.

When https://mcp.arthneura.com/mcp is hosted, this rule does not change.
