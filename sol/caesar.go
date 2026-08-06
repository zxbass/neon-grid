package sol

import "strings"

// Caesar shifts letters A-Z by n (positive = forward). Non-letters pass
// through unchanged. Mission 022.
func Caesar(s string, n int) string {
	var b strings.Builder
	shift := n % 26
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'A' && c <= 'Z':
			v := (int(c-'A') + shift) % 26
			if v < 0 {
				v += 26
			}
			b.WriteByte('A' + byte(v))
		case c >= 'a' && c <= 'z':
			v := (int(c-'a') + shift) % 26
			if v < 0 {
				v += 26
			}
			b.WriteByte('a' + byte(v))
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// Rot13 is Caesar with shift 13.
func Rot13(s string) string { return Caesar(s, 13) }
