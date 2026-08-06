package sol

import "strings"

// Vigenere encrypts/decrypts A-Z text. When skip is true, the key only
// advances on letters and non-letters pass through untouched (mission 023,
// part 2); otherwise every byte is processed positionally.
func Vigenere(s, key string, decrypt, skip bool) string {
	var b strings.Builder
	k := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			if skip {
				b.WriteByte(c)
				continue
			}
			b.WriteByte(c)
			k++
			continue
		}
		kb := key[k%len(key)] - 'A'
		var v byte
		if decrypt {
			v = (c - 'A' - kb + 26) % 26
		} else {
			v = (c - 'A' + kb) % 26
		}
		b.WriteByte('A' + v)
		k++
	}
	return b.String()
}
