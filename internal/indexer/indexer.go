package indexer

import (
    "encoding/hex"
    "fmt"
    "log"

    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// Names must match runtime metadata: AgentRegistry / VectorDb.
type EventAgentRegistryAgentRegistered struct {
    Phase      types.Phase
    Did        [32]byte
    Controller types.AccountID
    Topics     []types.Hash
}

type EventVectorDbCommitmentRegistered struct {
    Phase         types.Phase
    CommitmentID  [32]byte
    Provider      [32]byte
    Consumer      [32]byte
    MerkleRoot    [32]byte
    TotalChunks   types.U64
    ExpiresAt     types.U32
    Topics        []types.Hash
}

type EventRecords struct {
    types.EventRecords
    AgentRegistry_AgentRegistered     []EventAgentRegistryAgentRegistered
    VectorDb_CommitmentRegistered     []EventVectorDbCommitmentRegistered
}

func Run(ws string) error {
    api, err := gsrpc.NewSubstrateAPI(ws)
    if err != nil {
        return fmt.Errorf("connect %s: %w", ws, err)
    }

    sub, err := api.RPC.Chain.SubscribeNewHeads()
    if err != nil {
        return fmt.Errorf("subscribe heads: %w", err)
    }
    defer sub.Unsubscribe()

    log.Printf("listening for AgentRegistered + CommitmentRegistered on %s", ws)

    for {
        head := <-sub.Chan()
        hash, err := api.RPC.Chain.GetBlockHash(uint64(head.Number))
        if err != nil {
            log.Printf("block hash #%d: %v", head.Number, err)
            continue
        }

        meta, err := api.RPC.State.GetMetadataLatest()
        if err != nil {
            log.Printf("metadata: %v", err)
            continue
        }

        key, err := types.CreateStorageKey(meta, "System", "Events", nil)
        if err != nil {
            log.Printf("events key: %v", err)
            continue
        }

        raw, err := api.RPC.State.GetStorageRaw(key, hash)
        if err != nil {
            log.Printf("events storage #%d: %v", head.Number, err)
            continue
        }

        var events EventRecords
        err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
        if err != nil {
            log.Printf("block #%d events_bytes=%d decode=%v", head.Number, len(*raw), err)
            continue
        }

        for _, e := range events.AgentRegistry_AgentRegistered {
            log.Printf("AGENT registered block=#%d did=%s controller=%s",
                head.Number,
                hex.EncodeToString(e.Did[:]),
                hex.EncodeToString(e.Controller[:]),
            )
        }
        for _, e := range events.VectorDb_CommitmentRegistered {
            log.Printf("COMMITMENT registered block=#%d id=%s provider=%s consumer=%s chunks=%d",
                head.Number,
                hex.EncodeToString(e.CommitmentID[:]),
                hex.EncodeToString(e.Provider[:]),
                hex.EncodeToString(e.Consumer[:]),
                e.TotalChunks,
            )
        }

        if len(events.AgentRegistry_AgentRegistered) == 0 &&
            len(events.VectorDb_CommitmentRegistered) == 0 {
            log.Printf("block #%d ok (no market events, %d bytes)", head.Number, len(*raw))
        }
    }
}
