package announce

import (
    "testing"
    "time"
)

func TestSignVerify(t *testing.T) {
    var seed [32]byte
    copy(seed[:], []byte("arthneura-announce-test-seed-32b!"))
    msg := Message("aa", "http://127.0.0.1:8090", time.Now().Unix()+60)
    sig, pub, err := Sign(seed, msg)
    if err != nil {
        t.Fatal(err)
    }
    if err := Verify(pub, sig, msg, time.Now().Unix()+60); err != nil {
        t.Fatal(err)
    }
    if err := Verify(pub, sig, []byte("tamper"), time.Now().Unix()+60); err == nil {
        t.Fatal("tamper must fail")
    }
}
