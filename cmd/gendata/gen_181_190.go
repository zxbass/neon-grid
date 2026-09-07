package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------- 181 ethernet

func macStr(b []byte) string {
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02X", x)
	}
	return strings.Join(parts, ":")
}

func (s *set) gen181() {
	id := "181"
	type eth struct {
		dst, src []byte
		etype    uint16
		payload  []byte
	}
	frames := []eth{
		{[]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}, 0x0806,
			[]byte{0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x01, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x0A, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x02}},
		{[]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}, []byte{0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE}, 0x0800,
			[]byte{0x45, 0x00, 0x00, 0x1E, 0x00, 0x01, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x02, 0x08, 0x08, 0x08, 0x08, 0x04, 0xD2, 0x00, 0x35, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05}},
		{[]byte{0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE}, []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}, 0x0800,
			[]byte{0x45, 0x00, 0x00, 0x14, 0x00, 0x02, 0x00, 0x00, 0x40, 0x11, 0x00, 0x00, 0x08, 0x08, 0x08, 0x08, 0x0A, 0x00, 0x00, 0x02, 0x00, 0x35, 0x04, 0xD2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
	}
	var lines []string
	for _, f := range frames {
		var hdr []byte
		hdr = append(hdr, f.dst...)
		hdr = append(hdr, f.src...)
		hdr = append(hdr, byte(f.etype>>8), byte(f.etype))
		lines = append(lines, fmt.Sprintf("%x%s", hdr, hexBytes(f.payload)))
	}
	s.text(id, "frames.txt", strings.Join(lines, "\n"))

	var p1 []string
	typeNames := map[uint16]string{0x0800: "IPv4", 0x0806: "ARP", 0x86DD: "IPv6"}
	for i, f := range frames {
		p1 = append(p1, fmt.Sprintf("FRAME %d: %s -> %s TYPE=%s",
			i, macStr(f.src), macStr(f.dst), typeNames[f.etype]))
	}
	// decode ARP: sender IP = spa (offset 8+14 in arp payload: htype2 ptype2 hlen1 plen1 op2 sha6 spa4)
	arp := frames[0].payload
	spa := fmt.Sprintf("%d.%d.%d.%d", arp[14], arp[15], arp[16], arp[17])
	p2 := fmt.Sprintf("ARP: WHO-HAS? SENDER IP=%s SENDER MAC=%s", spa, macStr(arp[8:14]))
	s.expected(id, strings.Join(p1, "\n"), p2)
}

func hexBytes(b []byte) string {
	var sb strings.Builder
	for _, x := range b {
		fmt.Fprintf(&sb, "%02X", x)
	}
	return sb.String()
}

// ---------------------------------------------------------------- 182 ARP spoof

func (s *set) gen182() {
	id := "182"
	s.text(id, "arp.txt", "WHO-HAS 10.0.0.1 TELL 10.0.0.2\n10.0.0.1 IS-AT 00:11:22:33:44:55\n")
	attacker := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}
	victim := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	target := victim
	// Ethernet: dst = victim, src = attacker, type ARP
	var pkt []byte
	pkt = append(pkt, victim...)
	pkt = append(pkt, attacker...)
	pkt = append(pkt, 0x08, 0x06)
	// ARP reply: htype=1 ptype=0x0800 hlen=6 plen=4 op=2
	pkt = append(pkt, 0x00, 0x01, 0x08, 0x00, 0x06, 0x04, 0x00, 0x02)
	pkt = append(pkt, attacker...)
	pkt = append(pkt, 10, 0, 0, 1) // sender IP = 10.0.0.1 (spoofed)
	pkt = append(pkt, target...)
	pkt = append(pkt, 10, 0, 0, 2) // target IP = victim's IP
	p1 := fmt.Sprintf("SPOOFED_REPLY: %s", hexBytes(pkt))
	p2 := fmt.Sprintf("CACHE POISONED: 10.0.0.1 -> %s", macStr(attacker))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 183 IP fragmentation

func (s *set) gen183() {
	id := "183"
	msg := []byte("GREETINGS FROM THE NEON GRID: FRAGMENTATION IS THE WAY")
	// 3 fragments
	frag := func(off, mf int, data []byte) []byte {
		h := make([]byte, 20)
		h[0] = 0x45
		total := 20 + len(data)
		h[2], h[3] = byte(total>>8), byte(total) // total length, network order
		h[4], h[5] = 0x13, 0x37                  // ID 0x1337, network order
		flagsOff := (off / 8) & 0x1FFF
		if mf == 1 {
			flagsOff |= 0x2000
		}
		h[6], h[7] = byte(flagsOff>>8), byte(flagsOff) // network byte order
		h[8] = 64
		h[9] = 17 // UDP
		return append(h, data...)
	}
	f1 := frag(0, 1, msg[0:16])
	f2 := frag(16, 1, msg[16:32])
	f3 := frag(32, 0, msg[32:])
	var lines []string
	for _, f := range [][]byte{f1, f2, f3} {
		lines = append(lines, hexBytes(f))
	}
	s.text(id, "fragments.txt", strings.Join(lines, "\n"))
	var p1, p2 []string
	for i, f := range [][]byte{f1, f2, f3} {
		off := (binary.BigEndian.Uint16(f[6:]) & 0x1FFF) * 8
		mf := int(binary.BigEndian.Uint16(f[6:])) >> 13 & 1
		flags := "MF=0"
		if mf == 1 {
			flags = "MF=1"
		}
		p1 = append(p1, fmt.Sprintf("FRAG %d: ID=0x1337 OFF=%d LEN=%d %s", i, off, len(f)-20, flags))
		p2 = append(p2, fmt.Sprintf("OFFSET %d: %s", off, string(f[20:])))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n")+"\nREASSEMBLED: "+string(msg))
}

// ---------------------------------------------------------------- 184 TCP handshake

func tcpSeg(seq, ack uint32, flags byte, data []byte) []byte {
	s := make([]byte, 20)
	u32le(s[4:], seq)
	u32le(s[8:], ack)
	u16le(s[12:], 4096)
	s[13] = flags
	u16le(s[14:], 65535)
	return append(s, data...)
}

func (s *set) gen184() {
	id := "184"
	// handshake + data + close
	segments := [][]byte{
		tcpSeg(1000, 0, 0x02, nil),
		tcpSeg(5000, 1001, 0x12, nil),
		tcpSeg(1001, 5001, 0x10, nil),
		tcpSeg(1001, 5001, 0x18, []byte("GET / HTTP/1.0")),
		tcpSeg(1001, 5001, 0x11, nil),
		tcpSeg(5001, 1002, 0x11, nil),
		tcpSeg(1002, 5002, 0x10, nil),
	}
	var lines []string
	for _, s := range segments {
		lines = append(lines, hexBytes(s))
	}
	s.text(id, "tcp.txt", strings.Join(lines, "\n"))
	p1 := "SYN SEQ=1000\nSYN-ACK SEQ=5000 ACK=1001\nACK SEQ=1001 ACK=5001\nESTABLISHED"
	p2 := fmt.Sprintf("DATA: %d bytes\nPAYLOAD: GET / HTTP/1.0\nCLOSE: FIN-FIN-ACK-ACK", len("GET / HTTP/1.0"))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 185 UDP checksum

func udpChecksum(srcIP, dstIP [4]byte, udp []byte) uint16 {
	sum := func(b []byte) uint32 {
		var acc uint32
		for i := 0; i+1 < len(b); i += 2 {
			acc += uint32(b[i])<<8 | uint32(b[i+1])
		}
		if len(b)%2 == 1 {
			acc += uint32(b[len(b)-1]) << 8
		}
		return acc
	}
	var ph []byte
	ph = append(ph, srcIP[:]...)
	ph = append(ph, dstIP[:]...)
	ph = append(ph, 0, 17)
	ph = append(ph, byte(len(udp)>>8), byte(len(udp)))
	acc := sum(ph) + sum(udp)
	for acc>>16 != 0 {
		acc = (acc & 0xFFFF) + (acc >> 16)
	}
	return ^uint16(acc)
}

func (s *set) gen185() {
	id := "185"
	src := [4]byte{10, 0, 0, 2}
	dst := [4]byte{8, 8, 8, 8}
	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	udp := make([]byte, 8+len(payload))
	u16le(udp[0:], 1234) // sport
	u16le(udp[2:], 53)   // dport
	u16le(udp[4:], uint16(len(udp)))
	u16le(udp[6:], 0) // checksum placeholder
	copy(udp[8:], payload)
	cs := udpChecksum(src, dst, udp)
	u16le(udp[6:], cs)

	var ip []byte
	ip = append(ip, 0x45, 0x00, 0x00, byte(20+len(udp)), 0x00, 0x01, 0x00, 0x00, 40, 17, 0x00, 0x00)
	ip = append(ip, src[:]...)
	ip = append(ip, dst[:]...)
	ip = append(ip, udp...)
	s.text(id, "packet.txt", hexBytes(ip))

	// corrupted packet: flip a payload byte
	bad := append([]byte{}, udp...)
	bad[8] ^= 0xFF
	badCS := udpChecksum(src, dst, bad)
	var ip2 []byte
	ip2 = append(ip2, 0x45, 0x00, 0x00, byte(20+len(bad)), 0x00, 0x02, 0x00, 0x00, 40, 17, 0x00, 0x00)
	ip2 = append(ip2, src[:]...)
	ip2 = append(ip2, dst[:]...)
	ip2 = append(ip2, bad...)
	s.text(id, "corrupt.txt", hexBytes(ip2))
	_ = badCS

	p1 := fmt.Sprintf("CHECKSUM: 0x%04X", cs)
	p2 := "VALID: YES\nCORRUPT: BAD (checksum mismatch)"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 186 PCAP

func (s *set) gen186() {
	id := "186"
	// build pcap: global header + 3 records
	var pcap []byte
	pcap = append(pcap, 0xD4, 0xC3, 0xB2, 0xA1) // magic little-endian
	pcap = append(pcap, 2, 0, 4, 0)              // version 2.4
	pcap = append(pcap, 0, 0, 0, 0)              // thiszone
	pcap = append(pcap, 0, 0, 0, 0)              // sigfigs
	var sn [4]byte
	u32le(sn[:], 65535)
	pcap = append(pcap, sn[:]...) // snaplen
	var lt [4]byte
	u32le(lt[:], 1)
	pcap = append(pcap, lt[:]...) // LINKTYPE_ETHERNET
	pkts := [][]byte{
		tcpSeg(1000, 0, 0x02, nil),
		tcpSeg(1000, 0, 0x18, []byte("GET /index.html HTTP/1.1\r\nHost: neon.grid\r\n\r\n")),
		tcpSeg(5000, 1001, 0x18, []byte("HTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nHELLO GRID")),
	}
	ts := uint32(1000000)
	for i, p := range pkts {
		rec := make([]byte, 16)
		u32le(rec[0:], ts+uint32(i*100))
		u32le(rec[4:], 0)
		u32le(rec[8:], uint32(len(p)))
		u32le(rec[12:], uint32(len(p)))
		pcap = append(pcap, rec...)
		pcap = append(pcap, p...)
	}
	s.file(id, "capture.pcap", pcap)

	var p1 []string
	p1 = append(p1, fmt.Sprintf("PACKETS: %d", len(pkts)))
	for i, p := range pkts {
		p1 = append(p1, fmt.Sprintf("PKT %d: T=%.2f LEN=%d", i, float64(ts+uint32(i*100))/1e6, len(p)))
	}
	p2 := "STREAM: GET /index.html HTTP/1.1\r\nHost: neon.grid\r\n\r\nHTTP/1.1 200 OK\r\nContent-Length: 10\r\n\r\nHELLO GRID"
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 187 DNS wire

func (s *set) gen187() {
	id := "187"
	name := []byte{4, 'n', 'e', 'o', 'n', 4, 'g', 'r', 'i', 'd', 0}
	query := []byte{0x12, 0x34, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	query = append(query, name...)
	query = append(query, 0x00, 0x01, 0x00, 0x01)
	response := []byte{0x12, 0x34, 0x81, 0x80, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	response = append(response, name...)
	response = append(response, 0x00, 0x01, 0x00, 0x01)
	response = append(response, 0xC0, 0x0C)
	response = append(response, 0x00, 0x01, 0x00, 0x01) // A, IN
	u32be(response[len(response)-4:], 300)              // TTL
	response = append(response, 0x00, 0x04, 93, 184, 216, 34)
	s.text(id, "query.txt", hexBytes(query))
	s.text(id, "response.txt", hexBytes(response))
	p1 := "QUERY: neon.grid TYPE=A ID=0x1234"
	p2 := "ANSWER: neon.grid -> 93.184.216.34 TTL=300"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 188 HTTP sniff

func (s *set) gen188() {
	id := "188"
	req := "GET /index.html HTTP/1.1\r\nHost: neon.grid\r\nUser-Agent: zen-1.0\r\n\r\n"
	resp := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nNEON ONLINE!"
	s.text(id, "request.txt", req)
	s.text(id, "response.txt", resp)
	p1 := "METHOD: GET PATH: /index.html HOST: neon.grid"
	p2 := "CODE: 200 LENGTH: 12\nBODY: NEON ONLINE!"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 189 TLS ClientHello

var tlsCiphers = map[uint16]string{
	0x1301: "TLS_AES_128_GCM_SHA256",
	0x1302: "TLS_AES_256_GCM_SHA384",
	0x1303: "TLS_CHACHA20_POLY1305_SHA256",
	0xC02B: "ECDHE_RSA_AES128_GCM_SHA256",
	0xC02F: "ECDHE_RSA_AES256_GCM_SHA384",
}

func (s *set) gen189() {
	id := "189"
	sni := "neon.grid"
	var hello []byte
	hello = append(hello, 0x03, 0x03) // client version
	hello = append(hello, make([]byte, 32)...)
	hello = append(hello, 0) // session id len
	suites := []uint16{0x1301, 0x1302, 0x1303, 0xC02B, 0xC02F}
	hello = append(hello, byte(len(suites)*2))
	for _, s := range suites {
		hello = append(hello, byte(s>>8), byte(s))
	}
	hello = append(hello, 1, 0) // compression
	// extensions: SNI
	var sniExt []byte
	sniExt = append(sniExt, 0x00, 0x00) // type server_name
	nameBytes := []byte(sni)
	var lst []byte
	lst = append(lst, 0x00) // name type
	lst = append(lst, byte(len(nameBytes)>>8), byte(len(nameBytes)))
	lst = append(lst, nameBytes...)
	sniExt = append(sniExt, byte(len(lst)>>8), byte(len(lst)))
	sniExt = append(sniExt, lst...)
	hello = append(hello, byte(len(sniExt)>>8), byte(len(sniExt)))
	hello = append(hello, sniExt...)
	// handshake: type 1, len
	var hs []byte
	hs = append(hs, 0x01)
	hs = append(hs, byte(len(hello)>>16), byte(len(hello)>>8), byte(len(hello)))
	hs = append(hs, hello...)
	// record
	var rec []byte
	rec = append(rec, 0x16, 0x03, 0x01)
	rec = append(rec, byte(len(hs)>>8), byte(len(hs)))
	rec = append(rec, hs...)
	s.file(id, "hello.bin", rec)
	var ciphers []string
	for _, c := range suites {
		ciphers = append(ciphers, tlsCiphers[c])
	}
	p1 := fmt.Sprintf("TLS: 0x0303 SNI: %s", sni)
	p2 := "CIPHERS: " + strings.Join(ciphers, " ")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 190 bittorrent

func bencodeString(s string) []byte {
	return []byte(fmt.Sprintf("%d:%s", len(s), s))
}

func bencodeInt(n int) []byte {
	return []byte(fmt.Sprintf("i%de", n))
}

func (s *set) gen190() {
	id := "190"
	info := []byte{}
	info = append(info, 'd')
	info = append(info, bencodeString("name")...)
	info = append(info, bencodeString("grid.iso")...)
	info = append(info, bencodeString("length")...)
	info = append(info, bencodeInt(1337)...)
	info = append(info, bencodeString("pieces")...)
	info = append(info, bencodeString("ABCDEFGH")...)
	info = append(info, 'e')

	metainfo := []byte{}
	metainfo = append(metainfo, 'd')
	metainfo = append(metainfo, bencodeString("announce")...)
	metainfo = append(metainfo, bencodeString("udp://tracker.neon:6969")...)
	metainfo = append(metainfo, bencodeString("info")...)
	metainfo = append(metainfo, info...)
	metainfo = append(metainfo, 'e')
	s.text(id, "metainfo.txt", string(metainfo))

	// infohash = SHA1 of info dict bytes (exactly)
	hash := sha1.Sum(info)
	// handshake: pstrlen 19 + "BitTorrent protocol" + reserved + infohash + peerid
	handshake := []byte{19}
	handshake = append(handshake, []byte("BitTorrent protocol")...)
	handshake = append(handshake, make([]byte, 8)...)
	handshake = append(handshake, hash[:]...)
	handshake = append(handshake, []byte("ZEN-2049")...)
	s.file(id, "handshake.bin", handshake)
	p1 := fmt.Sprintf("INFO_HASH: %x", hash)
	p2 := fmt.Sprintf("ANNOUNCE: udp://tracker.neon:6969\nNAME: grid.iso LENGTH: 1337\nHANDSHAKE: %s", hexBytes(handshake))
	s.expected(id, p1, p2)
}

func init() {
	register("181", (*set).gen181)
	register("182", (*set).gen182)
	register("183", (*set).gen183)
	register("184", (*set).gen184)
	register("185", (*set).gen185)
	register("186", (*set).gen186)
	register("187", (*set).gen187)
	register("188", (*set).gen188)
	register("189", (*set).gen189)
	register("190", (*set).gen190)
}
