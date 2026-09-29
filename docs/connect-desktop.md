# Connect a desktop agent (lab)

PRE-TESTNET. Market API must be up: curl -sf http://127.0.0.1:8080/health

## Claude Desktop (stdio)

1. Build: ./scripts/build-mcp.sh
2. Quit Claude Desktop
3. Edit ~/Library/Application Support/Claude/claude_desktop_config.json
   Keep existing keys. Set:

    "mcpServers": {
      "arthneura": {
        "command": "/Users/YOU/arthneura-market/dist/arthneura-mcp",
        "args": [],
        "env": { "MARKET_URL": "http://127.0.0.1:8080" }
      }
    }

4. Reopen Claude. Prompt: Call arthneura_health
5. Allow the tool once.

Strangers should not be sent this folder path. They wait for a hosted MCP URL.

## URL clients (Cursor, OpenClaw, ChatGPT custom connector)

Same binary, HTTP flag:

    MARKET_URL=http://127.0.0.1:8080 ./dist/arthneura-mcp -http :8787

Client URL: http://127.0.0.1:8787

ChatGPT and Claude Connectors need a public HTTPS URL. That is not this lab.

## Limits

Not the ChatGPT store. No compose for end users unless they host the lab API.
