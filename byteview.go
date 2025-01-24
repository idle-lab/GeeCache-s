package geecaches

// A ByteView holds an immutable view of bytes.
type ByteView struct {
	btyes []byte
}

// Size returns the view's length
func (bv ByteView) Size() int64 {
	return int64(len(bv.btyes))
}

// String returns the data as a string, making a copy if necessary.
func (bv ByteView) String() string {
	return string(bv.btyes)
}

// ByteSlice returns a copy of the data as a byte slice.
func (bv ByteView) ByteSlice() []byte {
	return cloneBytes(bv.btyes)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
