package merkle

import "golang.org/x/crypto/blake2b"

// HashBytes is blake2b-256 — same as sp_io::hashing::blake2_256
// and offchain-vector-db Blake2bHasher::hash.
func HashBytes(data []byte) [32]byte {
    return blake2b.Sum256(data)
}

// HashPair is left || right, then blake2b-256.
// Odd node (no right sibling) stays the left hash — concat_and_hash(None).
func HashPair(left [32]byte, right *[32]byte) [32]byte {
    if right == nil {
        return left
    }
    var buf [64]byte
    copy(buf[0:32], left[:])
    copy(buf[32:64], right[:])
    return blake2b.Sum256(buf[:])
}
