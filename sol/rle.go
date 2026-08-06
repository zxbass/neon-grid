package sol

import "encoding/binary"

// RLEUnpack expands (count u16 LE, value byte) pairs (mission 032).
func RLEUnpack(data []byte) ([]byte, error) {
	var out []byte
	for i := 0; i+3 <= len(data); i += 3 {
		count := binary.LittleEndian.Uint16(data[i : i+2])
		v := data[i+2]
		for j := 0; j < int(count); j++ {
			out = append(out, v)
		}
	}
	return out, nil
}

// RLEPack compresses bytes: runs of 3+ identical bytes become (len, value);
// everything else becomes single (1, value) pairs (mission 032).
func RLEPack(data []byte) []byte {
	var out []byte
	i := 0
	for i < len(data) {
		j := i
		for j < len(data) && data[j] == data[i] {
			j++
		}
		run := j - i
		if run >= 3 {
			for run > 0 {
				n := run
				if n > 65535 {
					n = 65535
				}
				var h [3]byte
				binary.LittleEndian.PutUint16(h[:2], uint16(n))
				h[2] = data[i]
				out = append(out, h[:]...)
				run -= n
			}
		} else {
			for k := i; k < j; k++ {
				out = append(out, 1, 0, data[k])
			}
		}
		i = j
	}
	return out
}
