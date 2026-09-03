package merkle

func leafHashes(chunks [][]byte) [][32]byte {
    out := make([][32]byte, len(chunks))
    for i, c := range chunks {
        out[i] = HashBytes(c)
    }
    return out
}

func nextLayer(layer [][32]byte) [][32]byte {
    var next [][32]byte
    for i := 0; i < len(layer); i += 2 {
        if i+1 >= len(layer) {
            next = append(next, HashPair(layer[i], nil))
            continue
        }
        r := layer[i+1]
        next = append(next, HashPair(layer[i], &r))
    }
    return next
}

func Root(chunks [][]byte) [32]byte {
    if len(chunks) == 0 {
        return HashBytes(nil)
    }
    layer := leafHashes(chunks)
    for len(layer) > 1 {
        layer = nextLayer(layer)
    }
    return layer[0]
}

func Proof(chunks [][]byte, index int) [][32]byte {
    layer := leafHashes(chunks)
    var proof [][32]byte
    idx := index
    for len(layer) > 1 {
        if idx%2 == 0 {
            if idx+1 < len(layer) {
                proof = append(proof, layer[idx+1])
            }
        } else {
            proof = append(proof, layer[idx-1])
        }
        layer = nextLayer(layer)
        idx /= 2
    }
    return proof
}

func Verify(chunk []byte, index int, proof [][32]byte, root [32]byte) bool {
    h := HashBytes(chunk)
    idx := index
    for _, sib := range proof {
        s := sib
        if idx%2 == 0 {
            h = HashPair(h, &s)
        } else {
            h = HashPair(s, &h)
        }
        idx /= 2
    }
    return h == root
}
