package sol

import "encoding/binary"

// BuildDNSQuery builds a DNS query packet for name (mission 044, part 1).
// ID 0x1234, flags 0x0100, QDCOUNT=1, QTYPE=A, QCLASS=IN.
func BuildDNSQuery(id uint16, name string) []byte {
	var pkt []byte
	var h [12]byte
	binary.BigEndian.PutUint16(h[0:2], id)
	binary.BigEndian.PutUint16(h[2:4], 0x0100)
	binary.BigEndian.PutUint16(h[4:6], 1)
	pkt = append(pkt, h[:]...)
	for _, label := range splitLabels(name) {
		pkt = append(pkt, byte(len(label)))
		pkt = append(pkt, label...)
	}
	pkt = append(pkt, 0x00)
	// QTYPE + QCLASS
	var tail [4]byte
	binary.BigEndian.PutUint16(tail[0:2], 1) // A
	binary.BigEndian.PutUint16(tail[2:4], 1) // IN
	pkt = append(pkt, tail[:]...)
	return pkt
}

func splitLabels(name string) []string {
	var out []string
	cur := []byte{}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c == '.' {
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = cur[:0]
			}
			continue
		}
		cur = append(cur, c)
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}
