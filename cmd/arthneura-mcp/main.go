package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arthneura/arthneura-market/pkg/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func marketURL() string {
	u := os.Getenv("MARKET_URL")
	if u == "" {
		u = "http://127.0.0.1:8080"
	}
	return u
}

func market() *client.Client { return client.New(marketURL()) }

type emptyIn struct{}

type healthOut struct {
	Ok     bool   `json:"ok"`
	Market string `json:"market"`
	Error  string `json:"error,omitempty"`
}

func health(_ context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, healthOut, error) {
	u := marketURL()
	if err := market().Health(); err != nil {
		return nil, healthOut{Ok: false, Market: u, Error: err.Error()}, nil
	}
	return nil, healthOut{Ok: true, Market: u}, nil
}

type stampIn struct {
	OfferID int64 `json:"offer_id" jsonschema:"accepted offer id"`
}

type stampOut struct {
	OfferID      int64  `json:"offer_id"`
	Ready        bool   `json:"ready"`
	Reason       string `json:"reason"`
	Schema       string `json:"schema"`
	CommitmentID string `json:"commitment_id"`
	Error        string `json:"error,omitempty"`
}

func stamp(_ context.Context, _ *mcp.CallToolRequest, in stampIn) (*mcp.CallToolResult, stampOut, error) {
	if in.OfferID <= 0 {
		return nil, stampOut{Error: "offer_id must be > 0"}, nil
	}
	st, err := market().Stamp(in.OfferID)
	if err != nil {
		return nil, stampOut{OfferID: in.OfferID, Error: err.Error()}, nil
	}
	return nil, stampOut{
		OfferID: st.OfferID, Ready: st.Ready, Reason: st.Reason,
		Schema: st.Schema, CommitmentID: st.CommitmentID,
	}, nil
}

func main() {
	log.SetOutput(os.Stderr)
	s := mcp.NewServer(&mcp.Implementation{Name: "arthneura", Version: "0.1.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "arthneura_health",
		Description: "Check ArthNeura market API. No keys.",
	}, health)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "arthneura_stamp",
		Description: "Read-only: is this offer ready for chain register_commitment.",
	}, stamp)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
