package sol

import "hash/crc32"

// CRC16CCITT is CRC-CCITT "normal": poly 0x1021, init 0xFFFF, no reflection,
// no final xor. Used in missions 016 (part 2), 056, 086, 113, 119.
func CRC16CCITT(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

// CRC32IEEE is the standard CRC-32 IEEE 802.3 used by zlib/PNG (mission 037).
func CRC32IEEE(data []byte) uint32 {
	t := crc32.MakeTable(crc32.IEEE)
	return crc32.Checksum(data, t)
}
