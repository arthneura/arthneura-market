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


func ListingMessage(sellerDid, title string, price, expiresAt int64) []byte {
    return []byte(fmt.Sprintf("arthneura-listing|%s|%s|%d|%d", sellerDid, title, price, expiresAt))
}

func VerifyListing(controllerPub [32]byte, sig [64]byte, sellerDid, title string, price, expiresAt int64) error {
    return announce.Verify(controllerPub, sig, ListingMessage(sellerDid, title, price, expiresAt), expiresAt)
}
