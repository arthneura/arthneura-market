package announce

import (
    "fmt"
    "time"

    sr25519 "github.com/ChainSafe/go-schnorrkel"
)

const Domain = "arthneura-deliver-v1"

func Message(commitmentIDHex, url string, expiresAt int64) []byte {
    return []byte(fmt.Sprintf("%s|%s|%s|%d", Domain, commitmentIDHex, url, expiresAt))
}

func Sign(seed [32]byte, msg []byte) ([64]byte, [32]byte, error) {
    var out [64]byte
    var pubOut [32]byte
    msk, err := sr25519.NewMiniSecretKeyFromRaw(seed)
    if err != nil {
        return out, pubOut, err
    }
    kp := msk.ExpandEd25519()
    t := sr25519.NewSigningContext([]byte("substrate"), msg)
    sig, err := kp.Sign(t)
    if err != nil {
        return out, pubOut, err
    }
    enc := sig.Encode()
    copy(out[:], enc[:])
    pub, err := kp.Public()
    if err != nil {
        return out, pubOut, err
    }
    pubOut = pub.Encode()
    return out, pubOut, nil
}

func Verify(controllerPub [32]byte, sigBytes [64]byte, msg []byte, expiresAt int64) error {
    if expiresAt < time.Now().Unix() {
        return fmt.Errorf("announce expired")
    }
    pub, err := sr25519.NewPublicKey(controllerPub)
    if err != nil {
        return err
    }
    var s sr25519.Signature
    if err := s.Decode(sigBytes); err != nil {
        return fmt.Errorf("bad signature")
    }
    t := sr25519.NewSigningContext([]byte("substrate"), msg)
    ok, err := pub.Verify(&s, t)
    if err != nil {
        return err
    }
    if !ok {
        return fmt.Errorf("bad signature")
    }
    return nil
}
