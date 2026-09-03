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
    msk, err := sr25519.NewMiniSecretKeyFromRaw(seed)
    if err != nil {
        return [64]byte{}, [32]byte{}, err
    }
    kp := msk.ExpandEd25519()
    t := sr25519.NewSigningContext([]byte("substrate"), msg)
    sig, err := kp.Sign(t)
    if err != nil {
        return [64]byte{}, [32]byte{}, err
    }
    var out [64]byte
    copy(out[:], sig.Encode()[:])
    return out, kp.Public().Encode(), nil
}

func Verify(controllerPub [32]byte, sig [64]byte, msg []byte, expiresAt int64) error {
    if expiresAt < time.Now().Unix() {
        return fmt.Errorf("announce expired")
    }
    pub, err := sr25519.NewPublicKey(controllerPub)
    if err != nil {
        return err
    }
    var s sr25519.Signature
    if err := s.Decode(sig); err != nil {
        return fmt.Errorf("bad signature")
    }
    t := sr25519.NewSigningContext([]byte("substrate"), msg)
    if !pub.Verify(&s, t) {
        return fmt.Errorf("bad signature")
    }
    return nil
}
