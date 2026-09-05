package api

import (
    "net/http"
    "strconv"
    "strings"

    "github.com/arthneura/arthneura-market/internal/store"
)

func mountNext(mux *http.ServeMux, db *store.Store) {
    mux.HandleFunc("GET /v1/commitments/{id}/next", func(w http.ResponseWriter, r *http.Request) {
        id := strings.TrimSpace(r.PathValue("id"))
        c, err := db.GetCommitment(r.Context(), id)
        if err != nil {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "commitment not found"})
            return
        }
        writeJSON(w, http.StatusOK, nextAction(c, 0))
    })
    mux.HandleFunc("GET /v1/offers/{id}/next", func(w http.ResponseWriter, r *http.Request) {
        oid, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad id"})
            return
        }
        off, err := db.GetOffer(r.Context(), oid)
        if err != nil {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "offer not found"})
            return
        }
        if off.CommitmentID == "" {
            writeJSON(w, http.StatusOK, map[string]any{
                "offer_id": off.ID,
                "status":   off.Status,
                "ready":    false,
                "reason":   "no commitment linked",
                "submit":   false,
            })
            return
        }
        c, err := db.GetCommitment(r.Context(), off.CommitmentID)
        if err != nil {
            writeJSON(w, http.StatusOK, map[string]any{
                "offer_id":      off.ID,
                "commitment_id": off.CommitmentID,
                "ready":         false,
                "reason":        "commitment not indexed",
                "submit":        false,
            })
            return
        }
        writeJSON(w, http.StatusOK, nextAction(c, off.ID))
    })
}

func nextAction(c store.Commitment, offerID int64) map[string]any {
    out := map[string]any{
        "offer_id":      offerID,
        "commitment_id": c.CommitmentID,
        "status":        c.Status,
        "provider_did":  c.Provider,
        "consumer_did":  c.Consumer,
        "merkle_root":   c.MerkleRoot,
        "submit":        false,
    }
    switch c.Status {
    case "registered":
        out["action"] = "acknowledge_commitment"
        out["caller"] = "consumer"
        out["ready"] = true
        out["reason"] = ""
    case "acknowledged":
        out["action"] = "close_commitment"
        out["caller"] = "consumer"
        out["ready"] = true
        out["reason"] = ""
    case "disputed":
        out["action"] = "counter_dispute"
        out["caller"] = "provider"
        out["ready"] = true
        out["reason"] = ""
    case "settled", "finalized", "expired":
        out["action"] = ""
        out["caller"] = ""
        out["ready"] = false
        out["reason"] = "commitment already finished"
    default:
        out["action"] = ""
        out["caller"] = ""
        out["ready"] = false
        out["reason"] = "unknown status"
    }
    return out
}
