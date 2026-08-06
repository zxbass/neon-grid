package sol

const b64alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// Base64Encode implements Base64 by hand (mission 031).
func Base64Encode(data []byte) string {
	var out []byte
	for i := 0; i < len(data); i += 3 {
		var b [3]byte
		n := len(data) - i
		if n > 3 {
			n = 3
		}
		copy(b[:], data[i:i+n])

		out = append(out, b64alphabet[b[0]>>2])
		out = append(out, b64alphabet[(b[0]&0x03)<<4|b[1]>>4])
		if n >= 2 {
			out = append(out, b64alphabet[(b[1]&0x0F)<<2|b[2]>>6])
		} else {
			out = append(out, '=')
		}
		if n >= 3 {
			out = append(out, b64alphabet[b[2]&0x3F])
		} else {
			out = append(out, '=')
		}
	}
	return string(out)
}

// Base64Decode implements Base64 decoding by hand (mission 031).
func Base64Decode(s string) ([]byte, error) {
	idx := func(c byte) (int, bool) {
		switch {
		case c >= 'A' && c <= 'Z':
			return int(c - 'A'), true
		case c >= 'a' && c <= 'z':
			return int(c-'a') + 26, true
		case c >= '0' && c <= '9':
			return int(c-'0') + 52, true
		case c == '+':
			return 62, true
		case c == '/':
			return 63, true
		}
		return 0, false
	}
	var data []byte
	var acc uint32
	nBits := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '=' {
			break
		}
		if c == '\r' || c == '\n' {
			continue
		}
		v, ok := idx(c)
		if !ok {
			return nil, errInvalidBase64
		}
		acc = acc<<6 | uint32(v)
		nBits += 6
		if nBits >= 8 {
			nBits -= 8
			data = append(data, byte(acc>>uint(nBits)))
		}
	}
	return data, nil
}

var errInvalidBase64 = errorString("invalid base64")

type errorString string

func (e errorString) Error() string { return string(e) }
