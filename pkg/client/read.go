package client

import "fmt"

type Agent struct {
	Did          string `json:"did"`
	Controller   string `json:"controller"`
	Status       string `json:"status"`
	Label        string `json:"label"`
	Block        int64  `json:"block"`
	Capabilities int64  `json:"capabilities"`
}

type Commitment struct {
	CommitmentID string `json:"commitment_id"`
	Provider     string `json:"provider"`
	Consumer     string `json:"consumer"`
	MerkleRoot   string `json:"merkle_root"`
	TotalChunks  int64  `json:"total_chunks"`
	ExpiresAt    int64  `json:"expires_at"`
	Block        int64  `json:"block"`
	Status       string `json:"status"`
	DeliverURL   string `json:"deliver_url"`
	Metadata     string `json:"metadata"`
}

func (c *Client) ListListings() ([]Listing, error) {
	var wrap struct {
		Listings []Listing `json:"listings"`
	}
	if err := c.doJSON("GET", "/v1/listings", nil, &wrap); err != nil {
		return nil, err
	}
	if wrap.Listings == nil {
		return []Listing{}, nil
	}
	return wrap.Listings, nil
}

func (c *Client) ListOffers() ([]Offer, error) {
	var wrap struct {
		Offers []Offer `json:"offers"`
	}
	if err := c.doJSON("GET", "/v1/offers", nil, &wrap); err != nil {
		return nil, err
	}
	if wrap.Offers == nil {
		return []Offer{}, nil
	}
	return wrap.Offers, nil
}

func (c *Client) GetCommitment(id string) (Commitment, error) {
	if id == "" {
		return Commitment{}, fmt.Errorf("commitment id required")
	}
	var out Commitment
	err := c.doJSON("GET", "/v1/commitments/"+id, nil, &out)
	return out, err
}

func (c *Client) ListCommitments() ([]Commitment, error) {
	var wrap struct {
		Commitments []Commitment `json:"commitments"`
	}
	if err := c.doJSON("GET", "/v1/commitments", nil, &wrap); err != nil {
		return nil, err
	}
	if wrap.Commitments == nil {
		return []Commitment{}, nil
	}
	return wrap.Commitments, nil
}

func (c *Client) GetAgent(did string) (Agent, error) {
	if did == "" {
		return Agent{}, fmt.Errorf("did required")
	}
	var out Agent
	err := c.doJSON("GET", "/v1/agents/"+did, nil, &out)
	return out, err
}
