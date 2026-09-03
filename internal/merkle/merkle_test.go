package merkle

import (
    "encoding/hex"
    "testing"
)

func must32(t *testing.T, s string) [32]byte {
    t.Helper()
    b, err := hex.DecodeString(s)
    if err != nil || len(b) != 32 {
        t.Fatalf("bad hex %s", s)
    }
    var out [32]byte
    copy(out[:], b)
    return out
}

func TestHashBytesFixtures(t *testing.T) {
    cases := []struct {
        in   []byte
        want string
    }{
        {[]byte(""), "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8"},
        {[]byte("arthneura"), "be7d22bb8379d528f766e51527e20820171c508c672dbf575f74b5b766c3ebfb"},
        {[]byte("hello"), "324dcf027dd4a30a932c441f365a25e86b173defa4b8e58948253471b81b72cf"},
        {[]byte("world"), "9a3440c9d1529b122faceef33739b6e814616658d53faaf6e4f129fb20edfb13"},
    }
    for _, c := range cases {
        got := HashBytes(c.in)
        if hex.EncodeToString(got[:]) != c.want {
            t.Fatalf("hash(%q)=%x want %s", c.in, got, c.want)
        }
    }
}

func TestHashPairLeftRight(t *testing.T) {
    left := HashBytes([]byte("hello"))
    right := HashBytes([]byte("world"))
    got := HashPair(left, &right)
    want := "37d9b9204f631c0e327932eaecbfcc1587f13a40866059169e1a1a050593181a"
    if hex.EncodeToString(got[:]) != want {
        t.Fatalf("pair=%x want %s", got, want)
    }
}

func TestHashPairOddIsLeft(t *testing.T) {
    left := HashBytes([]byte("hello"))
    got := HashPair(left, nil)
    if got != left {
        t.Fatal("odd pair must copy left")
    }
}

func TestMust32Helper(t *testing.T) {
    _ = must32(t, "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8")
}

func TestRootTwoLeaves(t *testing.T) {
    root := Root([][]byte{[]byte("hello"), []byte("world")})
    want := "37d9b9204f631c0e327932eaecbfcc1587f13a40866059169e1a1a050593181a"
    if hex.EncodeToString(root[:]) != want {
        t.Fatalf("root=%x want %s", root, want)
    }
}

func TestRootThreeLeaves(t *testing.T) {
    root := Root([][]byte{[]byte("a"), []byte("b"), []byte("c")})
    want := "350bf288b7179755b0d6f6e91f8e6fa5b2ac2d8bcf6d0b3882b81a2d3171bc8c"
    if hex.EncodeToString(root[:]) != want {
        t.Fatalf("root3=%x want %s", root, want)
    }
}

func TestProofVerify(t *testing.T) {
    chunks := [][]byte{[]byte("hello"), []byte("world"), []byte("arth")}
    root := Root(chunks)
    for i, c := range chunks {
        p := Proof(chunks, i)
        if !Verify(c, i, p, root) {
            t.Fatalf("verify failed index %d", i)
        }
        if Verify([]byte("nope"), i, p, root) {
            t.Fatalf("bad chunk verified index %d", i)
        }
    }
}
