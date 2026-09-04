package offersign

import (
    "fmt"

    "github.com/arthneura/arthneura-market/internal/announce"
)

func Message(action string, offerOrListingID int64, did string, price, expiresAt int64) []byte {
    return []byte(fmt.Sprintf("arthneura-offer|%s|%d|%s|%d|%d", action, offerOrListingID, did, price, expiresAt))
}

func Verify(controllerPub [32]byte, sig [64]byte, action string, id int64, did string, price, expiresAt int64) error {
    return announce.Verify(controllerPub, sig, Message(action, id, did, price, expiresAt), expiresAt)
}
