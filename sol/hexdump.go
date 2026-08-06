package sol

import "fmt"

const hexdigits = "0123456789abcdef"

// HexDump formats data as 16-byte lines grouped by 4 with an ASCII column
// (mission 010). Non-printable bytes render as '.'.
func HexDump(data []byte) []string {
	var lines []string
	for off := 0; off < len(data); off += 16 {
		end := off + 16
		if end > len(data) {
			end = len(data)
		}
		chunk := data[off:end]
		hexPart := make([]byte, 0, 35)
		ascii := make([]byte, 16)
		for i := 0; i < 16; i++ {
			if i > 0 && i%4 == 0 {
				hexPart = append(hexPart, ' ')
			}
			if i < len(chunk) {
				b := chunk[i]
				hexPart = append(hexPart, hexdigits[b>>4], hexdigits[b&0xf])
				if b >= 0x20 && b < 0x7f {
					ascii[i] = b
				} else {
					ascii[i] = '.'
				}
			} else {
				hexPart = append(hexPart, ' ', ' ')
				ascii[i] = ' '
			}
		}
		lines = append(lines, fmt.Sprintf("%08x  %s  |%s|", off, hexPart, ascii))
	}
	return lines
}
