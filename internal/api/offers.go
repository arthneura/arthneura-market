package api

import (
	"encoding/hex"
    "encoding/json"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/arthneura/arthneura-market/internal/store"
)

func mountOffers(mux *http.ServeMux, db *store.Store) {
    mux.HandleFunc("GET /v1/offers", func(w http.ResponseWriter, r *http.Request) {
        items, err := db.ListOffers(r.Context())
        if err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        if items == nil {
            items = []store.Offer{}
        }
        writeJSON(w, http.StatusOK, map[string]any{"offers": items})
    })
    mux.HandleFunc("POST /v1/offers", func(w http.ResponseWriter, r *http.Request) {
        var body struct {
            ListingID int64  `json:"listing_id"`
            FromDid   string `json:"from_did"`
            ToDid     string `json:"to_did"`
            Price     int64  `json:"price"`
            ExpiresIn int64  `json:"expires_in_sec"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json"})
            return
        }
        if body.ListingID <= 0 || body.Price < 0 {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "listing_id and price required"})
            return
        }
        from, err := store.DecodeDid(body.FromDid)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad from_did"})
            return
        }
        to, err := store.DecodeDid(body.ToDid)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad to_did"})
            return
        }
        if body.ExpiresIn <= 0 {
            body.ExpiresIn = 1800
        }
        item, err := db.CreateOffer(r.Context(), body.ListingID, from, to, body.Price, time.Now().Add(time.Duration(body.ExpiresIn)*time.Second))
        if err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })
    mux.HandleFunc("POST /v1/offers/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        if err := db.CancelOffer(r.Context(), id); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": "cancelled"})
    })

    mux.HandleFunc("POST /v1/offers/{id}/counter", func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        var body struct {
            Price     int64 `json:"price"`
            ExpiresIn int64 `json:"expires_in_sec"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Price < 0 {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "price required"})
            return
        }
        if body.ExpiresIn <= 0 {
            body.ExpiresIn = 1800
        }
        item, err := db.CounterOffer(r.Context(), id, body.Price, time.Now().Add(time.Duration(body.ExpiresIn)*time.Second))
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })

    mux.HandleFunc("POST /v1/offers/{id}/accept", func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        var body struct {
            Did string `json:"did"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "did required"})
            return
        }
        who, err := store.DecodeDid(body.Did)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad did"})
            return
        }
        item, err := db.AcceptOffer(r.Context(), id, who)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })

    mux.HandleFunc("POST /v1/offers/{id}/spec", func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        var body struct {
            MerkleRoot      string `json:"merkle_root"`
            TotalChunks     int64  `json:"total_chunks"`
            ExpiresInBlocks int64  `json:"expires_in_blocks"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad json"})
            return
        }
        root, err := hex.DecodeString(strings.TrimSpace(body.MerkleRoot))
        if err != nil || len(root) != 32 || body.TotalChunks <= 0 || body.ExpiresInBlocks <= 0 {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "root 32-byte hex, chunks and blocks > 0"})
            return
        }
        if err := db.SetOfferSpec(r.Context(), id, root, body.TotalChunks, body.ExpiresInBlocks); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        item, err := db.GetOffer(r.Context(), id)
        if err != nil {
            writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })

    mux.HandleFunc("GET /v1/offers/{id}/stamp", func(w http.ResponseWriter, r *http.Request) {
        id, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        off, err := db.GetOffer(r.Context(), id)
        if err != nil {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "offer not found"})
            return
        }
        if off.Status != "accepted" {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "offer not accepted"})
            return
        }
        if off.MerkleRoot == "" || off.TotalChunks <= 0 || off.ExpiresInBlocks <= 0 {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "delivery spec missing"})
            return
        }
        listing, err := db.GetListing(r.Context(), off.ListingID)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "listing not found"})
            return
        }
        provider := listing.SellerDid
        consumer := off.ToDid
        if off.FromDid != provider {
            consumer = off.FromDid
        } else {
            consumer = off.ToDid
        }
        writeJSON(w, http.StatusOK, map[string]any{
            "offer_id":           off.ID,
            "provider_did":       provider,
            "consumer_did":       consumer,
            "merkle_root":        off.MerkleRoot,
            "total_chunks":       off.TotalChunks,
            "expires_in_blocks":  off.ExpiresInBlocks,
            "price":              off.Price,
            "ready":              true,
        })
    })
}
