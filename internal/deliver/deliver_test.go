package deliver

import (
    "testing"

    "github.com/arthneura/arthneura-market/internal/merkle"
)

func TestBoxRoundTripVerify(t *testing.T) {
    chunks := [][]byte{[]byte("hello"), []byte("world"), []byte("arth")}
    root := merkle.Root(chunks)
    for i, c := range chunks {
        box := EncodeBox(i, c, merkle.Proof(chunks, i))
        if !VerifyBox(box, len(chunks), root) {
            t.Fatalf("good box failed index %d", i)
        }
        box.Chunk = box.Chunk + "xx"
        if VerifyBox(box, len(chunks), root) {
            t.Fatalf("corrupt box passed index %d", i)
        }
    }
}
