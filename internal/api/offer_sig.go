package api

import (
    "context"
    "encoding/hex"
    "fmt"
    "strings"

    "github.com/arthneura/arthneura-market/internal/offersign"
    "github.com/arthneura/arthneura-market/internal/store"
)

func verifyOfferSig(ctx context.Context, db *store.Store, action string, id int64, did string, price, expiresAt int64, sigHex string) error {
    did = strings.TrimSpace(did)
    if did == "" || strings.TrimSpace(sigHex) == "" {
        return fmt.Errorf("did and signature required")
    }
    agent, err := db.GetAgent(ctx, did)
    if err != nil {
        return fmt.Errorf("signer agent not indexed")
    }
    ctrl, err := hex.DecodeString(agent.Controller)
    if err != nil || len(ctrl) != 32 {
        return fmt.Errorf("bad controller")
    }
    sb, err := hex.DecodeString(strings.TrimSpace(sigHex))
    if err != nil || len(sb) != 64 {
        return fmt.Errorf("signature must be 64-byte hex")
    }
    var pub [32]byte
    var sig [64]byte
    copy(pub[:], ctrl)
    copy(sig[:], sb)
    return offersign.Verify(pub, sig, action, id, did, price, expiresAt)
}
