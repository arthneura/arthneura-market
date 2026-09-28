# Connect a desktop agent (lab)

This is for owners who can drop a binary and a JSON file on their computer.
Claude Desktop, OpenClaw, Cursor, Claude Code, Grok Build, Muse Code.

Not ChatGPT.com / grok.com chat. Those need a hosted HTTP MCP later.

You do not clone this repo. You do not run docker compose unless you are hosting the lab API yourself.

## 1. Lab API

Local: MARKET_URL=http://127.0.0.1:8080
Hosted lab when it exists: MARKET_URL=https://api.arthneura.com

Health: GET $MARKET_URL/health -> {"ok":true}
The MCP process never talks to chain RPC 9944.

## 2. Binary

./scripts/build-mcp.sh
Writes dist/arthneura-mcp for this OS.
chmod +x dist/arthneura-mcp

## 3. Claude Desktop

macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
Windows: %APPDATA%\\Claude\\claude_desktop_config.json

command = ABSOLUTE path to dist/arthneura-mcp
env MARKET_URL = http://127.0.0.1:8080

Quit Claude Desktop fully. Reopen. Ask: Call arthneura_health.

## 4. OpenClaw

Settings -> MCP -> Add server -> Stdio.
Same command path. Env MARKET_URL.

openclaw mcp add arthneura --command /ABSOLUTE/PATH/dist/arthneura-mcp

## 5. Cursor / Claude Code / Grok Build / Muse Code

One stdio server. command = binary. env MARKET_URL.
Do not point these tools at ws://127.0.0.1:9944.

## 6. Tools v1 (read only)

arthneura_health
arthneura_stamp
arthneura_list_listings
arthneura_list_offers
arthneura_get_commitment

Listing create and register_commitment stay with the owner.

## 7. Limits

PRE-TESTNET lab. No ChatGPT store in this doc.
