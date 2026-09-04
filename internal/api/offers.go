package api

import (
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
}
