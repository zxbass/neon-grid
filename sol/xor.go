package sol

// XorByte applies a single-byte XOR to data (mission 021, part 1).
func XorByte(data []byte, k byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ k
	}
	return out
}

// XorRepeat applies a repeating multi-byte key (mission 021, part 2).
func XorRepeat(data, key []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out
}

// Hex parses a hex string ("01 ab CD") into bytes.
func Hex(s string) ([]byte, error) {
	clean := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			clean = append(clean, c)
		}
	}
	out := make([]byte, len(clean)/2)
	for i := 0; i < len(out); i++ {
		hi := nibble(clean[2*i])
		lo := nibble(clean[2*i+1])
		out[i] = hi<<4 | lo
	}
	return out, nil
}

func nibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return c - 'A' + 10
	}
}

// HexStr formats bytes as "aa bb cc".
func HexStr(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*3)
	for i, v := range b {
		if i > 0 {
			out = append(out, ' ')
		}
		out = append(out, digits[v>>4], digits[v&0xf])
	}
	return string(out)
}
