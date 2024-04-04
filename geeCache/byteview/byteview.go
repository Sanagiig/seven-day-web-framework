package byteview

type ByteView struct {
	b []byte
}

func New(bs []byte) ByteView {
	return ByteView{b: bs}
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
	newBytes := make([]byte, len(b))
	copy(newBytes, b)
	return newBytes
}
