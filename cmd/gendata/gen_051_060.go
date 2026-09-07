package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	"neon-grid/sol"
)

// ---------------------------------------------------------------- 051 UDP whisper

func (s *set) gen051() {
	id := "051"
	const key = byte(0x4D)
	msgs := []struct {
		seq  byte
		text string
		bad  bool
	}{
		{0, "CROW HERE", false},
		{1, "NEON LISTENS", false},
		{2, "ZEN SLEEPS", true}, // corrupted crc -> BAD
		{4, "WAKE THE GRID", false},
	}
	datagramCRC := func(seq byte, plain []byte) byte {
		crc := seq
		for _, b := range plain {
			crc ^= b
		}
		return crc
	}
	var cap []byte
	for _, m := range msgs {
		plain := []byte(m.text)
		crc := datagramCRC(m.seq, plain)
		if m.bad {
			crc ^= 0xFF
		}
		cap = append(cap, 0x7E, m.seq, byte(len(plain)))
		for _, b := range plain {
			cap = append(cap, b^key)
		}
		cap = append(cap, crc)
	}
	s.file(id, "capture.bin", cap)

	p1 := []string{fmt.Sprintf("DATAGRAMS: %d", len(msgs))}
	for _, m := range msgs {
		p1 = append(p1, fmt.Sprintf("MSG %d: %s", m.seq, m.text))
	}
	var p2 []string
	expected := byte(0)
	for _, m := range msgs {
		switch {
		case m.bad:
			p2 = append(p2, fmt.Sprintf("BAD %d", m.seq))
		case m.seq != expected:
			p2 = append(p2, fmt.Sprintf("REORDER %d", m.seq))
		default:
			p2 = append(p2, fmt.Sprintf("ACK %d", m.seq))
			expected++
		}
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 052 ICMP echo

func (s *set) gen052() {
	id := "052"
	s.text(id, "input.txt", "type=8 code=0 id=0x1234 seq=0x0001 data=hello\n")

	build := func(typ byte, idPkt, seq uint16, data []byte) []byte {
		pkt := []byte{typ, 0, 0, 0, byte(idPkt >> 8), byte(idPkt), byte(seq >> 8), byte(seq)}
		pkt = append(pkt, data...)
		cs := sol.ICMPChecksum(pkt)
		pkt[2], pkt[3] = byte(cs>>8), byte(cs)
		return pkt
	}
	req := build(8, 0x1234, 0x0001, []byte("hello"))
	reply := build(0, 0x1234, 0x0001, []byte("hello"))
	s.file(id, "reply.bin", reply)

	p1 := "PACKET: " + hexStr(req)
	cs := binary.BigEndian.Uint16(reply[2:4])
	p2 := fmt.Sprintf("CHECKSUM OK: 0x%04X\nID=0x1234 SEQ=0x0001\nREPLY OK: %d bytes",
		cs, len(reply)-8)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 053 sniffer

// pcapTCP builds an Ethernet frame (eth+ip+tcp) with correct IP/TCP checksums.
func pcapTCP(srcMAC, dstMAC []byte, srcIP, dstIP []byte,
	sport, dport uint16, seq, ack uint32, flags byte, payload []byte) []byte {

	var eth []byte
	eth = append(eth, dstMAC...)
	eth = append(eth, srcMAC...)
	eth = append(eth, 0x08, 0x00)

	var ip []byte
	ip = append(ip, 0x45, 0x00)
	ipLen := 20 + 20 + len(payload)
	ip = append(ip, byte(ipLen>>8), byte(ipLen))
	ip = append(ip, 0x00, 0x01) // id
	ip = append(ip, 0x40, 0x00) // DF
	ip = append(ip, 64, 6)      // ttl, proto
	ip = append(ip, 0, 0)       // checksum, filled below
	ip = append(ip, srcIP...)
	ip = append(ip, dstIP...)
	cs := sol.ICMPChecksum(ip)
	ip[10], ip[11] = byte(cs>>8), byte(cs)

	var tcp []byte
	tcp = append(tcp, byte(sport>>8), byte(sport))
	tcp = append(tcp, byte(dport>>8), byte(dport))
	tcp = append(tcp, byte(seq>>24), byte(seq>>16), byte(seq>>8), byte(seq))
	tcp = append(tcp, byte(ack>>24), byte(ack>>16), byte(ack>>8), byte(ack))
	tcp = append(tcp, 0x50, flags, 0xFF, 0xFF) // data offset 5, flags, window
	tcp = append(tcp, 0, 0, 0, 0)              // checksum, urg
	tcp = append(tcp, payload...)

	var ph []byte
	ph = append(ph, srcIP...)
	ph = append(ph, dstIP...)
	ph = append(ph, 0, 6)
	ph = append(ph, byte(len(tcp)>>8), byte(len(tcp)))
	ph = append(ph, tcp...)
	if len(ph)%2 == 1 {
		ph = append(ph, 0)
	}
	tcs := sol.ICMPChecksum(ph)
	tcp[16], tcp[17] = byte(tcs>>8), byte(tcs)

	return append(eth, append(ip, tcp...)...)
}

func (s *set) gen053() {
	id := "053"
	macA := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	macB := []byte{0x00, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE}
	ipA := []byte{10, 0, 0, 1}
	ipB := []byte{10, 0, 0, 2}

	pkts := [][]byte{
		// client -> server SYN
		pcapTCP(macA, macB, ipA, ipB, 51234, 80, 0x1000, 0, 0x02, nil),
		// server -> client SYN-ACK
		pcapTCP(macB, macA, ipB, ipA, 80, 51234, 0x2000, 0x1001, 0x12, nil),
		// client -> server PSH+ACK with GET payload
		pcapTCP(macA, macB, ipA, ipB, 51234, 80, 0x1001, 0x2001, 0x18,
			[]byte("GET / HTTP/1.1\r\n\r\n")),
	}
	ts := [][2]uint32{{1234567890, 123456}, {1234567890, 123789}, {1234567890, 124011}}

	var pcap []byte
	pcap = append(pcap, 0xD4, 0xC3, 0xB2, 0xA1) // little-endian magic
	pcap = append(pcap, 2, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	pcap = append(pcap, 0xFF, 0xFF, 0, 0) // snaplen 65535
	pcap = append(pcap, 1, 0, 0, 0)       // linktype Ethernet
	for i, p := range pkts {
		var hdr [16]byte
		u32le(hdr[0:], ts[i][0])
		u32le(hdr[4:], ts[i][1])
		u32le(hdr[8:], uint32(len(p)))
		u32le(hdr[12:], uint32(len(p)))
		pcap = append(pcap, hdr[:]...)
		pcap = append(pcap, p...)
	}
	s.file(id, "capture.pcap", pcap)

	var p1 []string
	for i, p := range pkts {
		p1 = append(p1, fmt.Sprintf("PKT %d: len=%d ts=%d.%06d", i, len(p), ts[i][0], ts[i][1]))
	}
	p2 := []string{
		"10.0.0.1:51234 -> 10.0.0.2:80 (TCP flags=0x02 len=0)",
		"10.0.0.2:80 -> 10.0.0.1:51234 (TCP flags=0x12 len=0)",
		"10.0.0.1:51234 -> 10.0.0.2:80 (TCP flags=0x18 len=18)",
		"SYN: 10.0.0.1:51234 -> 10.0.0.2:80",
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 054 wormhole

func (s *set) gen054() {
	id := "054"
	msg := []byte("NEON GRID IS A GHOST TOWN")
	const fragSize = 8
	total := (len(msg) + fragSize - 1) / fragSize
	var frames []byte
	for i := 0; i < total; i++ {
		end := (i + 1) * fragSize
		if end > len(msg) {
			end = len(msg)
		}
		data := msg[i*fragSize : end]
		var crc byte
		for _, b := range data {
			crc ^= b
		}
		frames = append(frames, 0x77, byte(i), byte(total))
		frames = append(frames, data...)
		frames = append(frames, crc)
	}
	s.file(id, "frames.bin", frames)

	p1 := fmt.Sprintf("FRAMES: %d\nDATA: %s", total, hexStr(msg))
	p2 := []string{}
	for i := 0; i < total; i++ {
		p2 = append(p2, fmt.Sprintf("FRAME %d CRC OK", i))
	}
	p2 = append(p2, "TUNNEL OK")
	s.expected(id, p1, strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 055 relay chain

func (s *set) gen055() {
	id := "055"
	log := "C: 2001 FWD 2002\n" +
		"C: 2002 FWD 2003\n" +
		"C: 2003 FWD 1200\n" +
		"H: 2001 2002 SEND 64\n" +
		"H: 2002 2003 SEND 64\n" +
		"H: 2003 1200 SEND 64\n" +
		"H: 1200 2003 ECHO 64\n" +
		"H: 2003 2002 ECHO 64\n" +
		"H: 2002 2001 ECHO 64\n" +
		"C: 2001 RECV 64\n"
	s.text(id, "hoplog.txt", log)

	lines := strings.Split(strings.TrimSpace(log), "\n")
	var chain []string
	var p1 []string
	for _, l := range lines {
		f := strings.Fields(l)
		if f[0] == "C:" && f[2] == "FWD" {
			p1 = append(p1, fmt.Sprintf("FWD %s -> %s", f[1], f[3]))
			chain = append(chain, f[3])
		}
	}
	hopOf := func(port string) int {
		for i, p := range chain {
			if p == port {
				return i + 1
			}
		}
		return -1
	}
	var p2 []string
	for _, l := range lines {
		f := strings.Fields(l)
		if f[0] != "H:" {
			continue
		}
		a, b, kind := f[1], f[2], f[3]
		if kind == "SEND" {
			p2 = append(p2, fmt.Sprintf("[hop%d] %s -> %s", hopOf(b), a, b))
		} else {
			p2 = append(p2, fmt.Sprintf("[hop%d] %s <- %s", hopOf(a), a, b))
		}
	}
	p2 = append(p2, "ROUNDTRIP OK")
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 056 ARQ

type arqFrame struct {
	seq  byte
	data []byte
	bad  bool
}

func arqFrames(frames []arqFrame) []byte {
	var out []byte
	for _, f := range frames {
		crc := sol.CRC16CCITT(f.data)
		if f.bad {
			crc ^= 0xFFFF
		}
		out = append(out, 0x5A, f.seq)
		out = append(out, f.data...)
		out = append(out, byte(crc>>8), byte(crc))
	}
	return out
}

func (s *set) gen056() {
	id := "056"
	w1 := []arqFrame{
		{0, []byte("NEONGRID"), false},
		{1, []byte("ARQ_IS_A"), true},
		{1, []byte("ARQ_IS_A"), true},
		{1, []byte("ARQ_IS_A"), false},
		{2, []byte("TIREDDOG"), false},
	}
	s.file(id, "window1.bin", arqFrames(w1))
	w4 := []arqFrame{
		{0, []byte("NEONGRID"), false},
		{1, []byte("ARQ_IS_A"), false},
		{2, []byte("TIREDDOG"), true},
		{3, []byte("SHUTDOWN"), false},
		{2, []byte("TIREDDOG"), false},
	}
	s.file(id, "window4.bin", arqFrames(w4))

	p1 := []string{
		"SEQ 0 ACK",
		"SEQ 1 NAK",
		"SEQ 1 ACK (retry 2)",
		"SEQ 2 ACK",
		"DONE: 24 bytes",
	}
	p2 := []string{
		"WINDOW=1 retries=2",
		"WINDOW=4 retries=1",
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 057 gossip

func (s *set) gen057() {
	id := "057"
	text := "5\n0: 1 2\n1: 0 3\n2: 0 4\n3: 1\n4: 2\n"
	s.text(id, "nodes.txt", text)

	adj := map[int][]int{}
	for _, l := range strings.Split(strings.TrimSpace(text), "\n")[1:] {
		nodeStr, rest, _ := strings.Cut(l, ":")
		node := 0
		fmt.Sscanf(nodeStr, "%d", &node)
		for _, f := range strings.Fields(rest) {
			var nb int
			fmt.Sscanf(f, "%d", &nb)
			adj[node] = append(adj[node], nb)
		}
	}
	dist := map[int]int{0: 0}
	queue := []int{0}
	var order []int
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		for _, nb := range adj[n] {
			if _, ok := dist[nb]; !ok {
				dist[nb] = dist[n] + 1
				queue = append(queue, nb)
			}
		}
	}
	var orderStr []string
	for _, n := range order {
		orderStr = append(orderStr, fmt.Sprintf("%d", n))
	}
	p1 := "ORDER: " + strings.Join(orderStr, " ")

	byWave := map[int][]int{}
	maxWave := 0
	total := 0
	for n, d := range dist {
		byWave[d] = append(byWave[d], n)
		if d > maxWave {
			maxWave = d
		}
		total += d
	}
	var p2 []string
	for w := 0; w <= maxWave; w++ {
		wave := append([]int{}, byWave[w]...)
		sort.Ints(wave)
		var ws []string
		for _, n := range wave {
			ws = append(ws, fmt.Sprintf("%d", n))
		}
		p2 = append(p2, fmt.Sprintf("WAVE %d: %s", w, strings.Join(ws, " ")))
	}
	p2 = append(p2, fmt.Sprintf("MAX WAVE: %d", maxWave))
	p2 = append(p2, fmt.Sprintf("AVG WAVE: %.2f", float64(total)/float64(len(dist))))
	s.expected(id, p1, strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 058 routing

func (s *set) gen058() {
	id := "058"
	routes := "10.0.0.0/8 -> 1.1.1.1\n" +
		"10.1.0.0/16 -> 2.2.2.2\n" +
		"10.1.2.0/24 -> 3.3.3.3\n" +
		"0.0.0.0/0 -> 9.9.9.9\n\n" +
		"lookup: 10.1.2.44\n" +
		"lookup: 10.5.1.1\n" +
		"lookup: 8.8.8.8\n"
	s.text(id, "routes.txt", routes)

	parseIP := func(s string) uint32 {
		var a, b, c, d byte
		fmt.Sscanf(s, "%d.%d.%d.%d", &a, &b, &c, &d)
		return uint32(a)<<24 | uint32(b)<<16 | uint32(c)<<8 | uint32(d)
	}
	ipStr := func(v uint32) string {
		return fmt.Sprintf("%d.%d.%d.%d", byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
	}
	type route struct {
		net  uint32
		mask uint32
		len  int
		hop  uint32
	}
	var tbl []route
	var lookups []string
	for _, l := range strings.Split(strings.TrimSpace(routes), "\n") {
		f := strings.Fields(l)
		if len(f) == 0 {
			continue
		}
		if f[0] == "lookup:" {
			lookups = append(lookups, f[1])
			continue
		}
		netStr, plenStr, _ := strings.Cut(f[0], "/")
		plen := 0
		fmt.Sscanf(plenStr, "%d", &plen)
		var mask uint32
		if plen > 0 {
			mask = ^uint32(0) << (32 - plen)
		}
		tbl = append(tbl, route{parseIP(netStr), mask, plen, parseIP(f[2])})
	}
	var p1 []string
	for _, l := range lookups {
		ip := parseIP(l)
		best := -1
		for i, r := range tbl {
			if ip&r.mask == r.net&r.mask && (best == -1 || r.len > tbl[best].len) {
				best = i
			}
		}
		p1 = append(p1, fmt.Sprintf("%s -> %s", l, ipStr(tbl[best].hop)))
	}

	graph := "R1: R2 R3\n" +
		"R2: R1 R4\n" +
		"R3: R1 R5\n" +
		"R4: R2 R6\n" +
		"R5: R3 R6 R7\n" +
		"R6: R4 R5 R7\n" +
		"R7: R5 R6\n" +
		"SRC R1 -> DST R7\n" +
		"SRC R4 -> DST X9\n"
	s.text(id, "graph.txt", graph)

	adj := map[string][]string{}
	var queries [][2]string
	for _, l := range strings.Split(strings.TrimSpace(graph), "\n") {
		f := strings.Fields(l)
		if f[0] == "SRC" {
			queries = append(queries, [2]string{f[1], f[4]})
			continue
		}
		node := strings.TrimSuffix(f[0], ":")
		adj[node] = f[1:]
	}
	var p2 []string
	for _, q := range queries {
		src, dst := q[0], q[1]
		prev := map[string]string{}
		queue := []string{src}
		prev[src] = ""
		found := false
		for len(queue) > 0 && !found {
			n := queue[0]
			queue = queue[1:]
			for _, nb := range adj[n] {
				if _, ok := prev[nb]; ok {
					continue
				}
				prev[nb] = n
				if nb == dst {
					found = true
					break
				}
				queue = append(queue, nb)
			}
		}
		if !found {
			p2 = append(p2, "LOOP DETECTED")
			continue
		}
		var path []string
		for n := dst; n != ""; n = prev[n] {
			path = append([]string{n}, path...)
		}
		p2 = append(p2, fmt.Sprintf("PATH: %s (%d hops)", strings.Join(path, " -> "), len(path)-1))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 059 botnet

func (s *set) gen059() {
	id := "059"
	s.text(id, "job.txt", "KEYS: 0..16777215\nCHUNK: 4096\nHASH: crc32(BE)\nTARGET: 0x42\nWORKERS: 10\n")

	const maxKey = 1 << 24
	const chunk = 4096
	const target = byte(0x42)
	tbl := crc32.MakeTable(crc32.IEEE)
	hits := 0
	for k := 0; k < maxKey; k++ {
		b := [4]byte{byte(k >> 24), byte(k >> 16), byte(k >> 8), byte(k)}
		if byte(crc32.Checksum(b[:], tbl)) == target {
			hits++
		}
	}
	chunks := maxKey / chunk
	lost := 0
	for c := 0; c < chunks; c++ {
		if c%17 == 0 {
			lost++
		}
	}
	p1 := fmt.Sprintf("HITS: %d", hits)
	p2 := fmt.Sprintf("WORKERS: 10  CHUNKS: %d  HITS: %d\nDONE: %d hits (%d chunks retried)",
		chunks, hits, hits, lost)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 060 ASCII-VNC

func (s *set) gen060() {
	id := "060"
	const rows, cols = 3, 30
	fullFrame := func(r0, r1, r2 string) []byte {
		b := []byte{0x01, byte(rows >> 8), byte(rows), byte(cols >> 8), byte(cols)}
		for _, r := range []string{r0, r1, r2} {
			p := []byte(r)
			for i := len(p); i < cols; i++ {
				p = append(p, ' ')
			}
			b = append(b, p...)
		}
		return b
	}
	delta := func(row, col int, data string) []byte {
		b := []byte{0x03, byte(row >> 8), byte(row), byte(col >> 8), byte(col)}
		return append(b, data...)
	}
	var dump []byte
	dump = append(dump, fullFrame("GATEWAY 12 // SYSTEM READY",
		"UNAUTHORIZED ACCESS IS FELONY", "> _")...)
	dump = append(dump, 0x02, 'A', 0x02, 'B', 0x02, 'C')
	dump = append(dump, delta(2, 0, "> ABC_")...)
	dump = append(dump, fullFrame("GATEWAY 12 // SYSTEM READY",
		"UNAUTHORIZED ACCESS IS FELONY", "> ABC_")...)
	dump = append(dump, delta(0, 0, "GATEWAY 42 // SYSTEM READY")...)
	s.file(id, "screen.dump", dump)

	p1 := "GATEWAY 12 // SYSTEM READY\nUNAUTHORIZED ACCESS IS FELONY\n> ABC_"
	p2 := "GATEWAY 42 // SYSTEM READY\nUNAUTHORIZED ACCESS IS FELONY\n> ABC_"
	s.expected(id, p1, p2)
}

func init() {
	register("051", (*set).gen051)
	register("052", (*set).gen052)
	register("053", (*set).gen053)
	register("054", (*set).gen054)
	register("055", (*set).gen055)
	register("056", (*set).gen056)
	register("057", (*set).gen057)
	register("058", (*set).gen058)
	register("059", (*set).gen059)
	register("060", (*set).gen060)
}
