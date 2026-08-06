package sol

// Frame parses the "Полтергейст" encoder frame format (mission 038):
// 0x7E | len u16 LE | data | xor-checksum | 0x7E. If the checksum fails but
// the last data byte is 0xFF, that pad byte is dropped and the checksum is
// recomputed (the "fix"). Returns data and whether it was fixed.
func Frame(pkt []byte) (data []byte, fixed, ok bool) {
	if len(pkt) < 6 || pkt[0] != 0x7E || pkt[len(pkt)-1] != 0x7E {
		return nil, false, false
	}
	length := int(pkt[1]) | int(pkt[2])<<8
	body := pkt[3 : len(pkt)-1]
	if len(body) != length+1 {
		return nil, false, false
	}
	payload := body[:length]
	cksum := body[length]
	if xorSum(payload) == cksum {
		return payload, false, true
	}
	if len(payload) > 0 && payload[len(payload)-1] == 0xFF {
		fixedPayload := payload[:len(payload)-1]
		if xorSum(fixedPayload) == cksum {
			return fixedPayload, true, true
		}
	}
	return payload, false, false
}

func xorSum(b []byte) byte {
	var s byte
	for _, v := range b {
		s ^= v
	}
	return s
}
