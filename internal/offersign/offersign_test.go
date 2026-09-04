package offersign

import (
    "testing"
    "time"

    "github.com/arthneura/arthneura-market/internal/announce"
)

func TestSignVerifyOffer(t *testing.T) {
    var seed [32]byte
    copy(seed[:], []byte("offer-sign-seed-arthneura-00001"))
    exp := time.Now().Unix() + 3600
    msg := Message("create", 1, "aa", 75, exp)
    sig, pub, err := announce.Sign(seed, msg)
    if err != nil {
        t.Fatal(err)
    }
    if err := Verify(pub, sig, "create", 1, "aa", 75, exp); err != nil {
        t.Fatal(err)
    }
    if err := Verify(pub, sig, "accept", 1, "aa", 75, exp); err == nil {
        t.Fatal("wrong action must fail")
    }
}
