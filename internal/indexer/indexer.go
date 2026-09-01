package indexer

import (
    "context"
    "encoding/hex"
    "fmt"
    "log"

    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"

    "github.com/arthneura/arthneura-market/internal/store"
)

type EventAgentRegistryAgentRegistered struct {
    Phase      types.Phase
    Did        [32]byte
    Controller types.AccountID
    Topics     []types.Hash
}

type EventVectorDbCommitmentRegistered struct {
    Phase        types.Phase
    CommitmentID [32]byte
    Provider     [32]byte
    Consumer     [32]byte
    MerkleRoot   [32]byte
    TotalChunks  types.U64
    ExpiresAt    types.U32
    Topics       []types.Hash
}

type EventRecords struct {
    types.EventRecords
    AgentRegistry_AgentRegistered []EventAgentRegistryAgentRegistered
    VectorDb_CommitmentRegistered []EventVectorDbCommitmentRegistered
}

func Run(ctx context.Context, ws, dsn string) error {
    var db *store.Store
    if dsn != "" {
        s, err := store.Open(ctx, dsn)
        if err != nil {
            return err
        }
        db = s
        defer db.Close()
        log.Printf("postgres connected")
    } else {
        log.Printf("DATABASE_URL empty — log only")
    }

    api, err := gsrpc.NewSubstrateAPI(ws)
    if err != nil {
        return fmt.Errorf("connect %s: %w", ws, err)
    }

    sub, err := api.RPC.Chain.SubscribeNewHeads()
    if err != nil {
        return fmt.Errorf("subscribe heads: %w", err)
    }
    defer sub.Unsubscribe()

    log.Printf("listening on %s", ws)

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case head := <-sub.Chan():
            if err := handleHead(ctx, api, db, head); err != nil {
                log.Printf("block #%d: %v", head.Number, err)
            }
        }
    }
}

func handleHead(ctx context.Context, api *gsrpc.SubstrateAPI, db *store.Store, head types.Header) error {
    hash, err := api.RPC.Chain.GetBlockHash(uint64(head.Number))
    if err != nil {
        return err
    }
    meta, err := api.RPC.State.GetMetadataLatest()
    if err != nil {
        return err
    }
    key, err := types.CreateStorageKey(meta, "System", "Events", nil)
    if err != nil {
        return err
    }
    raw, err := api.RPC.State.GetStorageRaw(key, hash)
    if err != nil {
        return err
    }

    var events EventRecords
    if err := types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events); err != nil {
        log.Printf("block #%d events_bytes=%d decode=%v", head.Number, len(*raw), err)
        return nil
    }

    for _, e := range events.AgentRegistry_AgentRegistered {
        log.Printf("AGENT registered block=#%d did=%s controller=%s",
            head.Number, hex.EncodeToString(e.Did[:]), hex.EncodeToString(e.Controller[:]))
        if db != nil {
            if err := db.UpsertAgent(ctx, e.Did[:], e.Controller[:], uint64(head.Number)); err != nil {
                log.Printf("db write did=%s: %v", hex.EncodeToString(e.Did[:]), err)
            }
        }
    }
    for _, e := range events.VectorDb_CommitmentRegistered {
        log.Printf("COMMITMENT registered block=#%d id=%s",
            head.Number, hex.EncodeToString(e.CommitmentID[:]))
    }
    if len(events.AgentRegistry_AgentRegistered) == 0 &&
        len(events.VectorDb_CommitmentRegistered) == 0 {
        log.Printf("block #%d ok (no market events, %d bytes)", head.Number, len(*raw))
    }
    return nil
}
