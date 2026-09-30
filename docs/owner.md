# Owner signs on this machine

The agent may speak any language. It does not hold the key.

You say: list this CSV for 1000.
The agent turns that into one command. You run the command
(or you allow a local shell to run it). The market checks the
signature. The chain is still the court.

    export MARKET_URL=http://127.0.0.1:8080
    export SIGNER=alice
    export OWNER_DID=f6ea766deb6576c32f794075d814a11fda6e26217056dab1fd2ca8e9a9e9c8e5

    go run ./cmd/owner listing -title "csv leads" -schema csv.v1 -price 1000
    go run ./cmd/owner offer -listing 16 -price 1000

MCP stays read-only. See mcp-reads.md.
