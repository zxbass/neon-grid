package sol

// BitReader reads n-bit fields MSB-first from a byte slice (mission 036).
type BitReader struct {
	data   []byte
	bitpos uint // 0..7 within current byte
	pos    int  // current byte index
}

func NewBitReader(data []byte) *BitReader { return &BitReader{data: data} }

// Read reads n bits (1..32), MSB-first, as an unsigned int.
func (r *BitReader) Read(n uint) (uint64, error) {
	if n > 32 {
		return 0, errInvalidBits
	}
	var acc uint64
	for n > 0 {
		if r.pos >= len(r.data) {
			return 0, errEof
		}
		take := 8 - r.bitpos
		if take > n {
			take = n
		}
		shift := 8 - r.bitpos - take
		mask := byte(0xFF >> (8 - take))
		acc = acc<<take | uint64((r.data[r.pos]>>shift)&mask)
		r.bitpos += take
		n -= take
		if r.bitpos == 8 {
			r.bitpos = 0
			r.pos++
		}
	}
	return acc, nil
}

// BitWriter packs n-bit fields MSB-first into a byte slice (mission 036).
type BitWriter struct {
	data []byte
	bits uint
}

func (w *BitWriter) Write(v uint64, n uint) {
	if n > 32 {
		return
	}
	// fill low n bits of v
	for n > 0 {
		space := 8 - w.bits
		take := space
		if take > n {
			take = n
		}
		shift := n - take
		bits := byte((v >> shift) & (1<<take - 1))
		if w.bits == 0 {
			w.data = append(w.data, 0)
		}
		w.data[len(w.data)-1] |= bits << (space - take)
		w.bits += take
		if w.bits == 8 {
			w.bits = 0
		}
		n -= take
	}
}

func (w *BitWriter) Bytes() []byte { return w.data }

var (
	errInvalidBits = errorString("invalid bit count")
	errEof         = errorString("unexpected end of data")
)
