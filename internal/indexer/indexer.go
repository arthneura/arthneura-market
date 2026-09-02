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

type EventVectorDbCommitmentAcknowledged struct {
    Phase           types.Phase
    CommitmentID    [32]byte
    AcknowledgedAt  types.U32
    Topics          []types.Hash
}

type EventVectorDbCommitmentSettled struct {
    Phase           types.Phase
    CommitmentID    [32]byte
    FinalStreamHash [32]byte
    ChunkCount      types.U64
    Topics          []types.Hash
}

type EventVectorDbDisputeRaised struct {
    Phase               types.Phase
    CommitmentID        [32]byte
    MerkleRoot          [32]byte
    DisputedChunkIndex  types.U64
    ReceivedChunkHash   [32]byte
    CounterDeadline     types.U32
    Topics              []types.Hash
}

type EventVectorDbDisputeCountered struct {
    Phase        types.Phase
    CommitmentID [32]byte
    Verdict      types.U8
    Topics       []types.Hash
}

type EventVectorDbDisputeFinalized struct {
    Phase        types.Phase
    CommitmentID [32]byte
    Verdict      types.U8
    Provider     [32]byte
    Consumer     [32]byte
    Topics       []types.Hash
}

type EventVectorDbCommitmentExpired struct {
    Phase        types.Phase
    CommitmentID [32]byte
    Topics       []types.Hash
}

type EventEscrowFundsLocked struct {
    Phase     types.Phase
    EscrowID  [32]byte
    Payer     types.AccountID
    Payee     types.AccountID
    Amount    types.U128
    Topics    []types.Hash
}

type EventEscrowFundsReleased struct {
    Phase     types.Phase
    EscrowID  [32]byte
    Payer     types.AccountID
    Payee     types.AccountID
    Amount    types.U128
    Topics    []types.Hash
}

type EventEscrowFundsRefunded struct {
    Phase     types.Phase
    EscrowID  [32]byte
    Payer     types.AccountID
    Amount    types.U128
    Topics    []types.Hash
}

type EventAgentRegistryReputationSlashed struct {
    Phase     types.Phase
    Did       [32]byte
    Amount    types.U32
    NewScore  types.U32
    Topics    []types.Hash
}

type EventRecords struct {
    types.EventRecords
    AgentRegistry_AgentRegistered     []EventAgentRegistryAgentRegistered
    AgentRegistry_ReputationSlashed   []EventAgentRegistryReputationSlashed
    VectorDb_CommitmentRegistered     []EventVectorDbCommitmentRegistered
    VectorDb_CommitmentAcknowledged   []EventVectorDbCommitmentAcknowledged
    VectorDb_CommitmentSettled        []EventVectorDbCommitmentSettled
    VectorDb_DisputeRaised            []EventVectorDbDisputeRaised
    VectorDb_DisputeCountered         []EventVectorDbDisputeCountered
    VectorDb_DisputeFinalized         []EventVectorDbDisputeFinalized
    VectorDb_CommitmentExpired        []EventVectorDbCommitmentExpired
    Escrow_FundsLocked                []EventEscrowFundsLocked
    Escrow_FundsReleased              []EventEscrowFundsReleased
    Escrow_FundsRefunded              []EventEscrowFundsRefunded
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

func setStatus(ctx context.Context, db *store.Store, id [32]byte, status string) {
    if db == nil {
        return
    }
    if err := db.SetCommitmentStatus(ctx, id[:], status); err != nil {
        log.Printf("status %s id=%s: %v", status, hex.EncodeToString(id[:]), err)
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
        log.Printf("AGENT registered block=#%d did=%s", head.Number, hex.EncodeToString(e.Did[:]))
        if db != nil {
            _ = db.UpsertAgent(ctx, e.Did[:], e.Controller[:], uint64(head.Number))
        }
    }
    for _, e := range events.VectorDb_CommitmentRegistered {
        log.Printf("COMMITMENT registered block=#%d id=%s", head.Number, hex.EncodeToString(e.CommitmentID[:]))
        if db != nil {
            _ = db.UpsertCommitment(ctx, e.CommitmentID[:], e.Provider[:], e.Consumer[:], e.MerkleRoot[:],
                uint64(e.TotalChunks), uint64(e.ExpiresAt), uint64(head.Number))
        }
    }
    for _, e := range events.VectorDb_CommitmentAcknowledged {
        log.Printf("COMMITMENT acknowledged id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "acknowledged")
    }
    for _, e := range events.VectorDb_CommitmentSettled {
        log.Printf("COMMITMENT settled id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "settled")
    }
    for _, e := range events.VectorDb_DisputeRaised {
        log.Printf("DISPUTE raised id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "disputed")
    }
    for _, e := range events.VectorDb_DisputeCountered {
        log.Printf("DISPUTE countered id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "disputed")
    }
    for _, e := range events.VectorDb_DisputeFinalized {
        log.Printf("DISPUTE finalized id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "finalized")
    }
    for _, e := range events.VectorDb_CommitmentExpired {
        log.Printf("COMMITMENT expired id=%s", hex.EncodeToString(e.CommitmentID[:]))
        setStatus(ctx, db, e.CommitmentID, "expired")
    }
    for _, e := range events.Escrow_FundsLocked {
        log.Printf("ESCROW locked id=%s", hex.EncodeToString(e.EscrowID[:]))
    }
    for _, e := range events.Escrow_FundsReleased {
        log.Printf("ESCROW released id=%s", hex.EncodeToString(e.EscrowID[:]))
    }
    for _, e := range events.Escrow_FundsRefunded {
        log.Printf("ESCROW refunded id=%s", hex.EncodeToString(e.EscrowID[:]))
    }
    return nil
}
