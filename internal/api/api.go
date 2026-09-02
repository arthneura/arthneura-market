package api

import (
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
    return mux
}

func writeJSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(body)
}
