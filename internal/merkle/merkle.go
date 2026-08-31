package merkle

import "golang.org/x/crypto/blake2b"

func HashBytes(data []byte) [32]byte {
    return blake2b.Sum256(data)
}
