package api

import (
    "encoding/hex"
    "encoding/json"
    "errors"
    "net/http"
    "strings"

    "github.com/jackc/pgx/v5"

    "github.com/arthneura/arthneura-market/internal/store"
)

func NewMux(db *store.Store) *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
        writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
    })
    mux.HandleFunc("GET /v1/agents", func(w http.ResponseWriter, r *http.Request) {
        agents, err := db.ListAgents(r.Context())
        if err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        if agents == nil {
            agents = []store.Agent{}
        }
        writeJSON(w, http.StatusOK, map[string]any{"agents": agents})
    })
    mux.HandleFunc("GET /v1/agents/{did}", func(w http.ResponseWriter, r *http.Request) {
        did := strings.TrimSpace(r.PathValue("did"))
        agent, err := db.GetAgent(r.Context(), did)
        if errors.Is(err, pgx.ErrNoRows) {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "agent not found"})
            return
        }
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, agent)
    })
    mux.HandleFunc("GET /v1/commitments", func(w http.ResponseWriter, r *http.Request) {
        items, err := db.ListCommitments(r.Context())
        if err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        if items == nil {
            items = []store.Commitment{}
        }
        writeJSON(w, http.StatusOK, map[string]any{"commitments": items})
    })
    mux.HandleFunc("GET /v1/commitments/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := strings.TrimSpace(r.PathValue("id"))
        item, err := db.GetCommitment(r.Context(), id)
        if errors.Is(err, pgx.ErrNoRows) {
            writeJSON(w, http.StatusNotFound, map[string]string{"error": "commitment not found"})
            return
        }
        if err != nil {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })
    mux.HandleFunc("POST /v1/commitments/{id}/deliver", func(w http.ResponseWriter, r *http.Request) {
        idHex := strings.TrimSpace(r.PathValue("id"))
        id, err := hex.DecodeString(idHex)
        if err != nil || len(id) != 32 {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad commitment id"})
            return
        }
        var body struct {
            URL string `json:"url"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.URL) == "" {
            writeJSON(w, http.StatusBadRequest, map[string]string{"error": "url required"})
            return
        }
        if err := db.SetDeliverURL(r.Context(), id, strings.TrimSpace(body.URL)); err != nil {
            writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
            return
        }
        item, err := db.GetCommitment(r.Context(), idHex)
        if err != nil {
            writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
            return
        }
        writeJSON(w, http.StatusOK, item)
    })
    return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(body)
}
