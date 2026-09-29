# ArthNeura MCP

Two transports. Same tools. Same MARKET_URL.

## Tools (read only)

- arthneura_health
- arthneura_stamp
- arthneura_list_listings
- arthneura_list_offers
- arthneura_get_commitment

No listing create. No register_commitment. No billing.

## Stdio (Claude Desktop lab)

    ./scripts/build-mcp.sh
    MARKET_URL=http://127.0.0.1:8080

claude_desktop_config.json command = path to dist/arthneura-mcp
See connect-desktop.md. This is the founder lab path, not the stranger path.

## HTTP (URL clients)

    MARKET_URL=http://127.0.0.1:8080 ./dist/arthneura-mcp -http :8787

Then point the client at http://127.0.0.1:8787

Cursor example:

    {
      "mcpServers": {
        "arthneura": {
          "url": "http://127.0.0.1:8787"
        }
      }
    }

Initialize must return serverInfo.name=arthneura.

Hosted later: MARKET_URL=https://api.arthneura.com and
MCP URL https://mcp.arthneura.com — not live.

No auth on -http. Do not publish :8787 or :9944 to the internet.
