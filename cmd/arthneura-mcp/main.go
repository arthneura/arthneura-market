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

type listOut struct {
	Error string `json:"error,omitempty"`
	N     int    `json:"n"`
	Items any    `json:"items"`
}

func listListings(_ context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, listOut, error) {
	items, err := market().ListListings()
	if err != nil {
		return nil, listOut{Error: err.Error()}, nil
	}
	return nil, listOut{N: len(items), Items: items}, nil
}

func listOffers(_ context.Context, _ *mcp.CallToolRequest, _ emptyIn) (*mcp.CallToolResult, listOut, error) {
	items, err := market().ListOffers()
	if err != nil {
		return nil, listOut{Error: err.Error()}, nil
	}
	return nil, listOut{N: len(items), Items: items}, nil
}

type commitIn struct {
	ID string `json:"id" jsonschema:"commitment id hex without 0x"`
}

func getCommitment(_ context.Context, _ *mcp.CallToolRequest, in commitIn) (*mcp.CallToolResult, client.Commitment, error) {
	item, err := market().GetCommitment(in.ID)
	if err != nil {
		return nil, client.Commitment{}, err
	}
	return nil, item, nil
}

func main() {
	log.SetOutput(os.Stderr)
	s := mcp.NewServer(&mcp.Implementation{Name: "arthneura", Version: "0.1.1"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "arthneura_health", Description: "Check ArthNeura market API. No keys."}, health)
	mcp.AddTool(s, &mcp.Tool{Name: "arthneura_stamp", Description: "Read-only: offer ready for register_commitment?"}, stamp)
	mcp.AddTool(s, &mcp.Tool{Name: "arthneura_list_listings", Description: "List market listings. Public read. No keys."}, listListings)
	mcp.AddTool(s, &mcp.Tool{Name: "arthneura_list_offers", Description: "List market offers. Public read. No keys."}, listOffers)
	mcp.AddTool(s, &mcp.Tool{Name: "arthneura_get_commitment", Description: "Get one commitment from the market index."}, getCommitment)
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
