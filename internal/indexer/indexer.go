package indexer

import (
    "fmt"
    "log"

    gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
    "github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

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

    log.Printf("listening for blocks on %s", ws)

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

        log.Printf("block #%d hash=%s events_bytes=%d", head.Number, hash.Hex(), len(*raw))

        var events types.EventRecords
        err = types.EventRecordsRaw(*raw).DecodeEventRecords(meta, &events)
        if err != nil {
            // custom pallet events often fail until we register them — still useful
            log.Printf("decode note #%d: %v", head.Number, err)
            continue
        }

        for _, e := range events.System_ExtrinsicSuccess {
            _ = e
        }
        log.Printf("block #%d system-decode ok", head.Number)
    }
}
