package deliver

import (
    "encoding/base64"
    "encoding/hex"

    "github.com/arthneura/arthneura-market/internal/merkle"
)

type ChunkBox struct {
    Index int      `json:"index"`
    Chunk string   `json:"chunk"` // base64
    Proof []string `json:"proof"` // hex of 32-byte hashes
}

func EncodeBox(index int, raw []byte, proof [][32]byte) ChunkBox {
    ps := make([]string, len(proof))
    for i, p := range proof {
        ps[i] = hex.EncodeToString(p[:])
    }
    return ChunkBox{
        Index: index,
        Chunk: base64.StdEncoding.EncodeToString(raw),
        Proof: ps,
    }
}

func DecodeBox(box ChunkBox) (raw []byte, proof [][32]byte, err error) {
    raw, err = base64.StdEncoding.DecodeString(box.Chunk)
    if err != nil {
        return nil, nil, err
    }
    proof = make([][32]byte, len(box.Proof))
    for i, h := range box.Proof {
        b, e := hex.DecodeString(h)
        if e != nil || len(b) != 32 {
            return nil, nil, errOr(e, "bad proof hash")
        }
        copy(proof[i][:], b)
    }
    return raw, proof, nil
}

func errOr(e error, msg string) error {
    if e != nil {
        return e
    }
    return &boxError{msg}
}

type boxError struct{ s string }

func (e *boxError) Error() string { return e.s }

func VerifyBox(box ChunkBox, total int, root [32]byte) bool {
    raw, proof, err := DecodeBox(box)
    if err != nil {
        return false
    }
    return merkle.Verify(raw, box.Index, total, proof, root)
}
