package geeCache2

type ByteView struct {
	b []byte
}

func (bv ByteView) Len() int {
	return len(bv.b)
}

func (bv ByteView) ByteSlice() []byte {
	return CloneBytes(bv.b)
}

func (bv ByteView) String() string {
	return string(bv.b)
}

func CloneBytes(b []byte) []byte {
	newBytes := make([]byte, 0, len(b))
	copy(newBytes, b)
	return newBytes
}
