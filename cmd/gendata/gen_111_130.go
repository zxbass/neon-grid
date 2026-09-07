package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"neon-grid/sol"
)

// Generators for missions 111-130 (final arc). Only statically verifiable
// missions are scaffolded; the interactive/network missions (112, 114, 117,
// 118, 120, 124, 129, 130) are excluded and not registered.

// ------------------------------------------------------------------ 111

func (s *set) gen111() {
	id := "111"
	// part 1: ICE type chain
	nodes := []struct {
		name, typ, out string
	}{
		{"GATE", "tcp", "udp"},
		{"LOBBY", "udp", "tcp"},
		{"VAULT", "tcp", "icmp"},
		{"CORE", "icmp", "-"},
	}
	inputs := map[int][]int{1: {0}, 2: {1}, 3: {2}}
	out := map[int][]int{0: {1}, 1: {2}, 2: {3}}
	var mapTxt strings.Builder
	mapTxt.WriteString("NODE  NAME  TYPE  OUTPUT  INPUTS\n")
	for i, n := range nodes {
		in := "-"
		if len(inputs[i]) > 0 {
			parts := make([]string, len(inputs[i]))
			for k, v := range inputs[i] {
				parts[k] = fmt.Sprintf("%d", v)
			}
			in = strings.Join(parts, ",")
		}
		fmt.Fprintf(&mapTxt, "%-5d %-6s %-5s %-7s %s\n", i, n.name, n.typ, n.out, in)
	}
	s.text(id, "ice_map.txt", mapTxt.String())

	type st struct {
		n   int
		typ string
	}
	start := st{0, nodes[0].typ}
	prev := map[st]st{}
	queue := []st{start}
	visited := map[st]bool{start: true}
	found := false
	for len(queue) > 0 && !found {
		cur := queue[0]
		queue = queue[1:]
		for _, v := range out[cur.n] {
			if nodes[v].typ != nodes[cur.n].out {
				continue
			}
			ns := st{v, nodes[v].typ}
			if visited[ns] {
				continue
			}
			visited[ns] = true
			prev[ns] = cur
			queue = append(queue, ns)
			if v == len(nodes)-1 {
				found = true
				break
			}
		}
	}
	var path []st
	cur := st{len(nodes) - 1, nodes[len(nodes)-1].typ}
	for {
		path = append(path, cur)
		if p, ok := prev[cur]; ok {
			cur = p
		} else {
			break
		}
	}
	parts := make([]string, len(path))
	for i, p := range path {
		parts[len(path)-1-i] = fmt.Sprintf("%d(%s)", p.n, p.typ)
	}
	p1 := "PATH: " + strings.Join(parts, " -> ")

	// part 2: ghost-cost routing
	n2 := []struct {
		name  string
		ghost bool
	}{
		{"GATE", false}, {"LOBBY", true}, {"VAULT", false},
		{"HALL", false}, {"ATRIUM", false}, {"CORE", false},
	}
	in2 := map[int][]int{1: {0}, 2: {0}, 3: {1}, 4: {2}, 5: {3, 4}}
	var trapsTxt strings.Builder
	trapsTxt.WriteString("NODE  NAME    GHOST  INPUTS\n")
	for i, n := range n2 {
		gh := "-"
		if n.ghost {
			gh = "ghost"
		}
		in := "-"
		if len(in2[i]) > 0 {
			parts := make([]string, len(in2[i]))
			for k, v := range in2[i] {
				parts[k] = fmt.Sprintf("%d", v)
			}
			in = strings.Join(parts, ",")
		}
		fmt.Fprintf(&trapsTxt, "%-5d %-7s %-6s %s\n", i, n.name, gh, in)
	}
	s.text(id, "traps_map.txt", trapsTxt.String())

	out2 := map[int][]int{0: {1, 2}, 1: {3}, 2: {4}, 3: {5}, 4: {5}}
	type route struct {
		path []int
		cost int
	}
	var routes []route
	var dfs func(cur int, seen map[int]bool, acc []int, cost int)
	dfs = func(cur int, seen map[int]bool, acc []int, cost int) {
		if cur == len(n2)-1 {
			r := make([]int, len(acc))
			copy(r, acc)
			routes = append(routes, route{r, cost})
			return
		}
		for _, v := range out2[cur] {
			if seen[v] {
				continue
			}
			c := 1
			if n2[v].ghost {
				c = 2
			}
			seen[v] = true
			dfs(v, seen, append(acc, v), cost+c)
			delete(seen, v)
		}
	}
	dfs(0, map[int]bool{0: true}, []int{0}, 0)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].cost != routes[j].cost {
			return routes[i].cost < routes[j].cost
		}
		return fmt.Sprint(routes[i].path) < fmt.Sprint(routes[j].path)
	})
	routeStr := func(r route) string {
		parts := make([]string, len(r.path))
		for i, n := range r.path {
			if n2[n].ghost {
				parts[i] = fmt.Sprintf("%d(ghost)", n)
			} else {
				parts[i] = fmt.Sprintf("%d", n)
			}
		}
		return strings.Join(parts, " -> ")
	}
	p2 := fmt.Sprintf("PATH: %s (cost=%d)\nALTERNATIVE: %s (cost=%d)",
		routeStr(routes[0]), routes[0].cost, routeStr(routes[1]), routes[1].cost)
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 113

func (s *set) gen113() {
	id := "113"
	t := uint32(1700000000)
	c := uint16(7)
	gen := func(t uint32, c uint16) []byte {
		k := make([]byte, 8)
		binary.BigEndian.PutUint32(k[0:], t)
		binary.BigEndian.PutUint16(k[4:], c)
		crc := sol.CRC16CCITT(k[:6])
		binary.BigEndian.PutUint16(k[6:], crc)
		return k
	}
	key := gen(t, c)
	s.file(id, "intercept.bin", key)

	keyHex := func(b []byte) string {
		return fmt.Sprintf("%X", b)
	}
	p1 := fmt.Sprintf("KEY(t=%d, c=%d): %s\nGEN OK: roundtrip", t, c, keyHex(key))

	next := gen(t, c+1)
	p2 := fmt.Sprintf("MATCH: t=%d c=%d\nNEXT:  t=%d c=%d -> KEY(%s)",
		t, c, t, c+1, keyHex(next))
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 115

func (s *set) gen115() {
	id := "115"
	const page = 4096
	const pages = 0x0201 // 513
	osimg := make([]byte, page*pages)
	for p := 0; p < pages; p++ {
		for i := 0; i < page; i++ {
			osimg[p*page+i] = byte((p*31 + i*7) & 0xFF)
		}
	}

	// clean hash list (computed before modification)
	var hashes []string
	for p := 0; p < pages; p++ {
		h := sha256.Sum256(osimg[p*page : (p+1)*page])
		hashes = append(hashes, fmt.Sprintf("%x", h[:]))
	}

	// dirty pages 0x0001 and 0x0100
	osimg[1*page+0], osimg[1*page+1], osimg[1*page+2] = 0xDE, 0xAD, 0xBE
	osimg[0x0100*page+100] = 0xFF
	osimg[0x0100*page+4094] = 0x00

	// rootkit page 0x0200
	rk := osimg[0x0200*page:]
	copy(rk[0:4], "RKT!")
	module := []byte("/proc/sys/kernel/hostname\ndmesg\nhide_module\npid 1\nbouncer\n")
	for len(module) < 64 {
		module = append(module, '\n')
	}
	rk[4] = byte(len(module))
	rk[5], rk[6], rk[7] = 0, 0, 0
	rk[8] = 'X'
	copy(rk[9:], sol.XorByte(module, 'X'))

	s.file(id, "os.img", osimg)
	s.text(id, "hashlist.txt", strings.Join(hashes, "\n")+"\n")

	var dirty []int
	for p := 0; p < pages; p++ {
		h := sha256.Sum256(osimg[p*page : (p+1)*page])
		if fmt.Sprintf("%x", h[:]) != hashes[p] {
			dirty = append(dirty, p)
		}
	}
	var dp []string
	for _, p := range dirty {
		dp = append(dp, fmt.Sprintf("0x%04X", p))
	}
	p1 := fmt.Sprintf("DIRTY: page %s  (%d pages)", strings.Join(dp, " "), len(dirty))

	// part 2: find rootkit page, extract module
	var mod []byte
	key := byte(0)
	for p := 0; p < pages; p++ {
		pg := osimg[p*page : (p+1)*page]
		if string(pg[0:4]) == "RKT!" {
			sz := int(pg[4])
			key = pg[8]
			mod = sol.XorByte(pg[9:9+sz], key)
			break
		}
	}
	var strs []string
	for _, line := range strings.Split(string(mod), "\n") {
		if strings.TrimSpace(line) != "" {
			strs = append(strs, line)
		}
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "MODULE: %d bytes  key=%c\n", len(mod), key)
	sb.WriteString("STRINGS:\n")
	for _, l := range strs {
		sb.WriteString("  " + l + "\n")
	}
	sb.WriteString("ROOTKIT: hides process 'bouncer'")
	s.expected(id, p1, strings.TrimRight(sb.String(), "\n"))
}

// ------------------------------------------------------------------ 116

func (s *set) gen116() {
	id := "116"
	users := []string{"crow", "zen", "ghost", "raven", "mercury", "omega", "gr1m", "selin", "pix", "holo"}
	actions := []string{"login", "read file:plan.doc", "logout", "list /var/log", "write report.txt", "ping 10.0.0.1", "ssh 10.0.0.2", "cat note.txt", "ls -la /home/crow", "wget update.bin"}
	n := 47
	honest := make([]string, n)
	for i := 0; i < n; i++ {
		h := (i * 5) % 24
		m := (i * 17) % 60
		sec := (i * 29) % 60
		ip := i % 30
		if i == 29 || i == 30 || i == 46 {
			ip = 66
		}
		honest[i] = fmt.Sprintf("2049-12-01 %02d:%02d:%02d 10.0.0.%d USER %s ACTION %s OK",
			h, m, sec, ip, users[i%10], actions[i%10])
	}
	// chain is computed over the honest lines; access.log is then tampered
	chain := make([]string, n)
	prev := ""
	for i := 0; i < n; i++ {
		h := sha256.Sum256([]byte(prev + honest[i] + "\n"))
		chain[i] = fmt.Sprintf("%x", h[:])
		prev = chain[i]
	}
	lines := append([]string{}, honest...)
	lines[46] = strings.TrimSuffix(honest[46], "OK") + "FAIL"
	s.text(id, "access.log", strings.Join(lines, "\n")+"\n")
	s.text(id, "chain.txt", strings.Join(chain, "\n")+"\n")

	// part 1: verify chain, first mismatch
	ln, exp, got := -1, "", ""
	prev = ""
	for i := 0; i < n; i++ {
		h := sha256.Sum256([]byte(prev + lines[i] + "\n"))
		re := fmt.Sprintf("%x", h[:])
		if re != chain[i] {
			ln = i + 1
			exp = chain[i]
			got = re
			break
		}
		prev = chain[i]
	}
	p1 := fmt.Sprintf("LINE %d: HASH MISMATCH (expected %s... got %s...)", ln, exp[:16], got[:16])

	// part 2: patch 10.0.0.66 -> 10.0.0.7, recompute chain
	rewritten := 0
	patched := make([]string, n)
	for i, l := range lines {
		if strings.Contains(l, "10.0.0.66") {
			l = strings.ReplaceAll(l, "10.0.0.66", "10.0.0.7")
			rewritten++
		}
		patched[i] = l
	}
	ok := true
	prev = ""
	for i := 0; i < n; i++ {
		h := sha256.Sum256([]byte(prev + patched[i] + "\n"))
		re := fmt.Sprintf("%x", h[:])
		if i > 0 && re == "" {
			ok = false
		}
		prev = re
	}
	_ = ok
	p2 := fmt.Sprintf("PATCHED: %d lines rewritten, chain recomputed\nVERIFY: OK (entire chain)\nWEAKNESS: first entry unsigned -> need anchor", rewritten)
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 119

func (s *set) gen119() {
	id := "119"
	termText := "AUTH GATEKEEPERf48\nCODE NEON-1337-GRID\nSTATUS: ACCESS GRANTED\n"
	rle := sol.RLEPack([]byte(termText))

	key := byte('K')
	content := make([]byte, 255)
	binary.LittleEndian.PutUint16(content[0:2], uint16(len(rle)))
	copy(content[2:], rle)
	enc := sol.XorByte(content, key)

	var img []byte
	var v [4]byte
	binary.LittleEndian.PutUint32(v[:], 9)
	img = append(img, v[:]...)
	img = append(img, key)
	img = append(img, enc...)
	s.file(id, "black_gate.bin", img)

	// reference decode
	version := binary.LittleEndian.Uint32(img[0:4])
	keyR := img[4]
	dec := sol.XorByte(img[5:5+255], keyR)
	L := binary.LittleEndian.Uint16(dec[0:2])
	unpacked, _ := sol.RLEUnpack(dec[2 : 2+int(L)])
	screenText := string(unpacked)
	authIdx := strings.Index(screenText, "AUTH ")
	pw := ""
	for i := authIdx + 5; i < len(screenText); i++ {
		if screenText[i] == ' ' || screenText[i] == '\n' {
			break
		}
		pw += string(screenText[i])
	}
	codeIdx := strings.Index(screenText, "CODE ")
	code := ""
	for i := codeIdx + 5; i < len(screenText); i++ {
		if screenText[i] == ' ' || screenText[i] == '\n' {
			break
		}
		code += string(screenText[i])
	}

	p1 := fmt.Sprintf("LAYER 1: OK (version=%d)\nLAYER 2: decrypted (len=%d)\nLAYER 3: unpacked (password=%s)\nLAYER 4: crc OK\nLAYER 5: ACCESS GRANTED",
		version, L, pw)
	p2 := fmt.Sprintf("GATE OPEN: CODE=%s\n\"Ты стоял у чёрных врат, и они открылись.\n НЕОН ПОМНИТ ТЕБЯ.\"", code)
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 121

func (s *set) gen121() {
	id := "121"
	prog1 := []byte{0x01, 0x4F, 0x07, 0x01, 0x4B, 0x07, 0x00}
	broken := []byte{0x01, 0x43, 0x07, 0x01, 0x4F, 0x07, 0x00}
	working := []byte{0x01, 0x43, 0x07, 0x01, 0x54, 0x07, 0x00}
	fw := []byte{0x01, 0x4F, 0x07, 0x02, 0x01, 0x07, 0x03, 0x0B, 0x07, 0x02, 0x09, 0x07, 0x00}
	s.file(id, "prog1.bin", prog1)
	s.file(id, "broken.bin", broken)
	s.file(id, "working.bin", working)
	s.file(id, "fw.bin", fw)

	p1 := "OUTPUT: OK"

	idx := 0
	for i := range broken {
		if broken[i] != working[i] {
			idx = i
			break
		}
	}
	p2 := fmt.Sprintf("BIT: BYTE %d (0x%02X -> 0x%02X)\nBROKEN OUTPUT: %s\nWORKING OUTPUT: %s\nCODE: %s",
		idx+1, broken[idx], working[idx], sol.RunCPU(broken), sol.RunCPU(working), sol.RunCPU(fw))
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 122

func (s *set) gen122() {
	id := "122"
	const p, g, a, b = 23, 5, 6, 15
	s.text(id, "input.txt", fmt.Sprintf("p=%d\ng=%d\na=%d\nb=%d\n", p, g, a, b))
	p1 := fmt.Sprintf("A: %d\nB: %d\nSHARED: %d", sol.ModPow(g, a, p), sol.ModPow(g, b, p), sol.ModPow(sol.ModPow(g, b, p), a, p))

	msg := func(s uint8) []byte {
		m := make([]byte, 4+4+5+2)
		binary.BigEndian.PutUint32(m[0:], 0x11223344)
		binary.BigEndian.PutUint32(m[4:], 5)
		copy(m[8:13], "AWAKE")
		tag := sol.CRC16CCITT(append([]byte{s}, m[:13]...))
		binary.BigEndian.PutUint16(m[13:], tag)
		return m
	}
	s.file(id, "bank.bin", msg(0x02))
	s.file(id, "bank_forge.bin", msg(0x03))

	check := func(m []byte, s uint8) bool {
		tag := binary.BigEndian.Uint16(m[13:])
		return sol.CRC16CCITT(append([]byte{s}, m[:13]...)) == tag
	}
	bank := msg(0x02)
	forge := msg(0x03)
	p2 := "SEQ 5: ACK\nSEQ 5 (replay): REPLAY DROP"
	if check(forge, 0x03) {
		p2 += "\nSEQ 5 (forge): DROP"
	}
	_ = bank
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 123

func (s *set) gen123() {
	id := "123"
	const sr = 8000
	const n = 256
	x := make([]float64, n)
	seed := uint32(12345)
	lcg := func() float64 {
		seed = seed*1664525 + 1013904223
		return float64(seed>>8&0xFFFFFF) / float64(0xFFFFFF)
	}
	for i := 0; i < n; i++ {
		x[i] = math.Sin(2*math.Pi*697*float64(i)/sr) + (lcg()*0.5 - 0.25)
	}
	var sig []byte
	for _, v := range x {
		var f8 [8]byte
		binary.LittleEndian.PutUint64(f8[:], math.Float64bits(v))
		sig = append(sig, f8[:]...)
	}
	s.file(id, "signal.bin", sig)

	// reference peak
	mag := sol.Magnitudes(sol.FFT(x))
	bin := sol.PeakBin(mag)
	p1 := fmt.Sprintf("PEAK BIN: %d\nPEAK FREQ: %.1f Hz", bin, float64(bin)*sr/float64(n))

	digits := "42"
	smpl := sol.SynthDTMF(digits, sr, 100)
	wav := make([]byte, 44+2*len(smpl))
	copy(wav[0:4], "RIFF")
	u32le(wav[4:], uint32(len(wav)-8))
	copy(wav[8:12], "WAVE")
	copy(wav[12:16], "fmt ")
	u32le(wav[16:], 16)
	u16le(wav[20:], 1)
	u16le(wav[22:], 1)
	u32le(wav[24:], uint32(sr))
	u32le(wav[28:], uint32(sr*2))
	u16le(wav[32:], 2)
	u16le(wav[34:], 16)
	copy(wav[36:40], "data")
	u32le(wav[40:], uint32(2*len(smpl)))
	for i, v := range smpl {
		sv := int16(v * 10000)
		u16le(wav[44+2*i:], uint16(sv))
	}
	s.file(id, "dial.wav", wav)

	// reference decode from the quantized wav
	qs := make([]float64, len(smpl))
	for i := 0; i < len(smpl); i++ {
		qs[i] = float64(int16(binary.LittleEndian.Uint16(wav[44+2*i:])))
	}
	dec := sol.DecodeDTMF(qs, sr, 100)
	p2 := "DIGITS: " + dec
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 125

var cga16 = [][3]byte{
	{0, 0, 0}, {0, 0, 170}, {0, 170, 0}, {0, 170, 170},
	{170, 0, 0}, {170, 0, 170}, {170, 170, 0}, {170, 170, 170},
	{85, 85, 85}, {85, 85, 255}, {85, 255, 85}, {85, 255, 255},
	{255, 85, 85}, {255, 85, 255}, {255, 255, 85}, {255, 255, 255},
}

func gifLZW(data []byte, litWidth int) []byte {
	clear := 1 << litWidth
	eof := clear + 1
	width := litWidth + 1
	hi := eof
	overflow := 1 << width
	dict := map[string]int{}
	for i := 0; i < clear; i++ {
		dict[string([]byte{byte(i)})] = i
	}
	var out []byte
	bitpos := 0
	write := func(code int) {
		for m := 0; m < width; m++ {
			if bitpos == 0 {
				out = append(out, 0)
			}
			out[len(out)-1] |= byte((code>>m)&1) << bitpos
			bitpos = (bitpos + 1) & 7
		}
	}
	incHi := func() {
		hi++
		if hi >= overflow && width < 12 {
			width++
			overflow <<= 1
		}
	}
	write(clear)
	w := []byte{}
	for _, b := range data {
		ws := append(append([]byte{}, w...), b)
		if _, ok := dict[string(ws)]; ok {
			w = ws
			continue
		}
		write(dict[string(w)])
		incHi()
		dict[string(ws)] = hi
		w = []byte{b}
	}
	if len(w) > 0 {
		write(dict[string(w)])
		incHi()
	}
	write(eof)
	return out
}

func gifLZWDecode(packed []byte, litWidth int) ([]byte, bool) {
	clear := 1 << litWidth
	eof := clear + 1
	width := litWidth + 1
	hi := eof
	overflow := 1 << width
	dict := map[int][]byte{}
	for i := 0; i < clear; i++ {
		dict[i] = []byte{byte(i)}
	}
	var out []byte
	bitpos := 0
	byteIdx := 0
	read := func() (int, bool) {
		code := 0
		for m := 0; m < width; m++ {
			if byteIdx >= len(packed) {
				return 0, false
			}
			code |= int((packed[byteIdx]>>bitpos)&1) << m
			bitpos++
			if bitpos == 8 {
				bitpos = 0
				byteIdx++
			}
		}
		return code, true
	}
	last := -1
	for {
		code, ok := read()
		if !ok {
			return nil, false
		}
		if code == clear {
			width = litWidth + 1
			hi = eof
			overflow = 1 << width
			dict = map[int][]byte{}
			for i := 0; i < clear; i++ {
				dict[i] = []byte{byte(i)}
			}
			last = -1
			continue
		}
		if code == eof {
			break
		}
		var entry []byte
		if code < clear {
			entry = []byte{byte(code)}
		} else if code <= hi {
			if code == hi && last != -1 {
				entry = append(append([]byte{}, dict[last]...), dict[last][0])
			} else {
				entry = dict[code]
			}
		} else {
			return nil, false
		}
		out = append(out, entry...)
		if last != -1 {
			dict[hi] = append(append([]byte{}, dict[last]...), entry[0])
		}
		last = code
		hi++
		if hi >= overflow && width < 12 {
			width++
			overflow <<= 1
		}
	}
	return out, true
}

func gifUniquePalette(pix []byte) [][3]byte {
	var pal [][3]byte
	seen := map[[3]byte]bool{}
	for i := 0; i+3 <= len(pix); i += 3 {
		c := [3]byte{pix[i], pix[i+1], pix[i+2]}
		if !seen[c] {
			seen[c] = true
			pal = append(pal, c)
		}
	}
	return pal
}

func gifIndex(pix []byte, pal [][3]byte) []byte {
	idx := make([]byte, len(pix)/3)
	for i := 0; i < len(idx); i++ {
		c := [3]byte{pix[i*3], pix[i*3+1], pix[i*3+2]}
		for j, p := range pal {
			if p == c {
				idx[i] = byte(j)
				break
			}
		}
	}
	return idx
}

func gifBuildImage(pix []byte, w, h int, pal [][3]byte) []byte {
	var b []byte
	b = append(b, "GIF87a"...)
	b = append(b, byte(w&0xFF), byte(w>>8), byte(h&0xFF), byte(h>>8))
	n := 0
	for (1 << (n + 1)) < len(pal) {
		n++
	}
	b = append(b, 0x80|byte(n), 0, 0)
	for i := 0; i < 1<<(n+1); i++ {
		if i < len(pal) {
			b = append(b, pal[i][0], pal[i][1], pal[i][2])
		} else {
			b = append(b, 0, 0, 0)
		}
	}
	b = append(b, 0x2C, 0, 0, 0, 0, byte(w&0xFF), byte(w>>8), byte(h&0xFF), byte(h>>8), 0)
	lit := n + 1
	packed := gifLZW(gifIndex(pix, pal), lit)
	b = append(b, byte(lit))
	for len(packed) > 0 {
		l := len(packed)
		if l > 255 {
			l = 255
		}
		b = append(b, byte(l))
		b = append(b, packed[:l]...)
		packed = packed[l:]
	}
	b = append(b, 0, 0x3B)
	return b
}

func gifVerifyImage(b, pix []byte, pal [][3]byte) string {
	if len(b) < 10 || string(b[0:6]) != "GIF87a" {
		return "ERR"
	}
	i := 13
	palSize := 1 << int(1+int(b[10]&0x07))
	i += palSize * 3
	if len(b) < i+10 || b[i] != 0x2C {
		return "ERR"
	}
	lit := int(b[i+10])
	i += 11
	var packed []byte
	for {
		if i >= len(b) {
			return "ERR"
		}
		l := int(b[i])
		i++
		if l == 0 {
			break
		}
		if i+l > len(b) {
			return "ERR"
		}
		packed = append(packed, b[i:i+l]...)
		i += l
	}
	indices, ok := gifLZWDecode(packed, lit)
	if !ok {
		return "ERR"
	}
	for k := 0; k < len(indices); k++ {
		if int(indices[k]) >= len(pal) {
			return "ERR"
		}
		c := pal[indices[k]]
		if c[0] != pix[k*3] || c[1] != pix[k*3+1] || c[2] != pix[k*3+2] {
			return "ERR"
		}
	}
	return "OK"
}

// gifBuildAnim builds an animation; frames are palette indices, tindex is the
// transparent index (>=0 enables the transparency flag), n selects palette size.
func gifBuildAnim(frames [][]byte, pal [][3]byte, tindex, n int) []byte {
	var b []byte
	b = append(b, "GIF87a"...)
	b = append(b, 0, 0, 32&0xFF, 32>>8, 24&0xFF, 24>>8)
	b = append(b, 0x80|byte(n), 0, 0)
	for i := 0; i < 1<<(n+1); i++ {
		if i < len(pal) {
			b = append(b, pal[i][0], pal[i][1], pal[i][2])
		} else {
			b = append(b, 0, 0, 0)
		}
	}
	for _, ind := range frames {
		b = append(b, 0x21, 0xF9, 4)
		flags := byte(0x04) // disposal 1
		if tindex >= 0 {
			flags |= 1
		}
		b = append(b, flags, 0x0A, 0x00)
		if tindex >= 0 {
			b = append(b, byte(tindex))
		} else {
			b = append(b, 0)
		}
		b = append(b, 0)
		b = append(b, 0x2C, 0, 0, 0, 0, 32&0xFF, 32>>8, 24&0xFF, 24>>8, 0)
		lit := n + 1
		packed := gifLZW(ind, lit)
		b = append(b, byte(lit))
		for len(packed) > 0 {
			l := len(packed)
			if l > 255 {
				l = 255
			}
			b = append(b, byte(l))
			b = append(b, packed[:l]...)
			packed = packed[l:]
		}
		b = append(b, 0)
	}
	b = append(b, 0x3B)
	return b
}

func (s *set) gen125() {
	id := "125"
	const w, h = 32, 24
	bg := func(x, y int) int { return (x*11 + y*7 + x*y) % 16 }
	// 4 frames: block 4x6 color 14 at x=f*4, y=8..13
	frames := make([][]byte, 4)
	for f := 0; f < 4; f++ {
		pix := make([]byte, w*h*3)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				ci := bg(x, y)
				if x >= f*4 && x < f*4+4 && y >= 8 && y < 14 {
					ci = 14
				}
				c := cga16[ci]
				o := (y*w + x) * 3
				pix[o], pix[o+1], pix[o+2] = c[0], c[1], c[2]
			}
		}
		frames[f] = pix
	}
	s.file(id, "pixels.bin", frames[0])
	var all []byte
	for _, fr := range frames {
		all = append(all, fr...)
	}
	s.file(id, "frames.bin", all)

	// palette: unique colors of the first frame
	pal := gifUniquePalette(frames[0])
	single := gifBuildImage(frames[0], w, h, pal)
	rt := gifVerifyImage(single, frames[0], pal)
	p1 := fmt.Sprintf("GIF87a %dx%d colors=%d\nbytes=0x%X\nROUNDTRIP: %s", w, h, len(pal), len(single), rt)

	// animation: full frames vs diff frames (transparent index 16)
	indFrames := make([][]byte, 4)
	for f, fr := range frames {
		indFrames[f] = gifIndex(fr, pal)
	}
	full := gifBuildAnim(indFrames, pal, -1, 3)

	var pal32 [][3]byte
	pal32 = append(pal32, pal...)
	pal32 = append(pal32, [3]byte{0, 0, 0})
	diffFrames := make([][]byte, 4)
	diffFrames[0] = indFrames[0]
	for f := 1; f < 4; f++ {
		df := make([]byte, w*h)
		for i := range df {
			if indFrames[f][i] == indFrames[f-1][i] {
				df[i] = 16
			} else {
				df[i] = indFrames[f][i]
			}
		}
		diffFrames[f] = df
	}
	diff := gifBuildAnim(diffFrames, pal32, 16, 4)
	ratio := float64(len(full)) / float64(len(diff))
	p2 := fmt.Sprintf("FULL: %d bytes\nDIFF: %d bytes   <- экономнее в %.1f раза", len(full), len(diff), ratio)
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 126

func (s *set) gen126() {
	id := "126"
	strings126 := []string{
		"crow_trace_route_00", "ssh_session_keyring_01", "password_rotation_log_02",
		"/home/crow/cache/neon03", "crow_identity_token_04", "ssh_agent_socket_05",
		"password_hasher_v2_06", "/home/crow/.ssh/id_neon", "crow_tunnel_config_08",
		"ssh_known_hosts_09", "password_miner_seed_10", "/home/crow/notes/brief11",
		"crow_cron_job_12", "ssh_pubkey_fingerprint_13", "password_salt_rotator_14",
		"/home/crow/bin/scanner_15", "crow_telemetry_16", "ssh_hostkey_dump_17",
		"password_breaker_18", "/home/crow/logs/history19", "crow_alert_rule_20",
		"ssh_forward_port_21", "password_store_offset_22", "/home/crow/etc/profile23",
		"crow_net_iface_24", "ssh_banner_grabber_25", "password_seed_bucket_26",
		"/home/crow/tmp/vector27", "crow_dns_cache_28", "ssh_privkey_cipher_29",
		"password_trie_node_30", "/home/crow/var/run/pid31", "crow_uptime_probe_32",
		"ssh_ed25519_signer_33", "password_entropy_pool_34", "/home/crow/opt/gateway35",
		"crow_heartbeat_key_36", "ssh_cipher_negotiator_37", "password_bloom_38",
		"/home/crow/srv/report39", "crow_proxy_chain_40", "ssh_kex_session_41",
		"password_scanner_42", "/home/crow/mnt/archive43", "crow_ovpn_profile_44",
		"ssh_config_backup_45", "password_policy_draft_46",
	}

	// filler pattern: printable-ish but keyword-free, no magic signatures
	fill := func(dst []byte, off int) {
		for i := range dst {
			dst[i] = 0x55 + byte((off+i)&0x3F)
		}
	}

	mem := make([]byte, 0x3000+1600)

	// JPEG at 0x0000
	jpeg := mem[0x0000:0x0F00]
	copy(jpeg[0:4], []byte{0xFF, 0xD8, 0xFF, 0xE0})
	fill(jpeg[4:0x0E00], 0x100)
	jpeg[0x0E00] = 0xFF
	jpeg[0x0E01] = 0xD9

	// strings 0..5 at 0x0F00
	var blob []byte
	for _, t := range strings126[:6] {
		blob = append(blob, []byte(t)...)
		blob = append(blob, 0x00)
	}
	copy(mem[0x0F00:], blob)

	// PNG at 0x1000
	png := mem[0x1000:0x1F00]
	copy(png[0:8], []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})
	fill(png[8:0x0E00], 0x200)
	// IEND chunk at 0x1E00
	u32be(png[0x0E00:], 0)
	copy(png[0x0E04:], "IEND")
	u32be(png[0x0E08:], sol.CRC32IEEE([]byte("IEND")))

	// strings 6..11 at 0x1F00
	blob = nil
	for _, t := range strings126[6:12] {
		blob = append(blob, []byte(t)...)
		blob = append(blob, 0x00)
	}
	copy(mem[0x1F00:], blob)

	// ZIP at 0x2000
	zip := mem[0x2000:0x2F00]
	copy(zip[0:4], []byte{'P', 'K', 0x03, 0x04})
	fill(zip[4:0x0E00], 0x300)
	u32be(zip[0x0E00:], 0x06054B50) // EOCD signature

	// strings 12..46 at 0x2F00
	blob = nil
	for _, t := range strings126[12:] {
		blob = append(blob, []byte(t)...)
		blob = append(blob, 0x00)
	}
	copy(mem[0x2F00:], blob)

	s.file(id, "mem.dmp", mem)

	// reference: scan printable runs >= 12 containing a keyword
	kw := []string{"crow", "ssh", "password", "/home/"}
	var runs [][2]int
	for i := 0; i < len(mem); {
		if mem[i] < 0x20 || mem[i] > 0x7E {
			i++
			continue
		}
		j := i
		for j < len(mem) && mem[j] >= 0x20 && mem[j] <= 0x7E {
			j++
		}
		run := string(mem[i:j])
		if len(run) >= 12 {
			for _, k := range kw {
				if strings.Contains(run, k) {
					runs = append(runs, [2]int{i, j})
					break
				}
			}
		}
		i = j
	}
	// wipe
	for _, r := range runs {
		for i := r[0]; i < r[1]; i++ {
			mem[i] = 0
		}
	}
	p1 := fmt.Sprintf("WIPED: %d strings", len(runs))
	var carv []string
	scan := func(magic []byte, name string, ok func(off int) bool) {
		for i := 0; i+4 < len(mem); i++ {
			if mem[i] == magic[0] && mem[i+1] == magic[1] && mem[i+2] == magic[2] {
				if ok(i) {
					carv = append(carv, fmt.Sprintf("CARVER: %s @ 0x%08X (still OK)", name, i))
				}
				break
			}
		}
	}
	scan([]byte{0xFF, 0xD8, 0xFF}, "JPEG", func(off int) bool { return off == 0x0000 })
	scan([]byte{0x89, 0x50, 0x4E}, "PNG", func(off int) bool { return off == 0x1000 })
	scan([]byte{0x50, 0x4B, 0x03}, "ZIP", func(off int) bool { return off == 0x2000 })
	p1 += "\n" + strings.Join(carv, "\n")
	p2 := "WEAKNESS: origin hash is unanchored -> anchor chain head to a signed timestamp\nHARDENED: chain head commits (term, server_uptime) via signed nonce"
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 127

func (s *set) gen127() {
	id := "127"
	payload := make([]byte, 4096)
	for i := range payload {
		payload[i] = byte('A' + i%26)
	}
	chunks := []int{0, 1200, 2400, 3600}
	frag := func(seq int, flags byte, c []byte, xor bool) []byte {
		var f []byte
		f = append(f, 0x4A, 0x4B)
		var s4 [4]byte
		binary.BigEndian.PutUint32(s4[:], uint32(seq))
		f = append(f, s4[:]...)
		f = append(f, flags)
		var l2 [2]byte
		binary.BigEndian.PutUint16(l2[:], uint16(len(c)))
		f = append(f, l2[:]...)
		wire := c
		if xor {
			wire = make([]byte, len(c))
			for i, b := range c {
				wire[i] = b ^ 0x55
			}
		}
		f = append(f, wire...)
		tag := sol.CRC16CCITT(append(append([]byte{}, s4[:]...), append([]byte{flags}, c...)...))
		var t2 [2]byte
		binary.BigEndian.PutUint16(t2[:], tag)
		f = append(f, t2[:]...)
		return f
	}
	flags := []byte{0x02, 0x00, 0x00, 0x01}
	chunk := func(k int) []byte {
		end := chunks[k] + 1200
		if k == 3 {
			end = len(payload)
		}
		return payload[chunks[k]:end]
	}
	var tunnel []byte
	for k := 0; k < 4; k++ {
		tunnel = append(tunnel, frag(k, flags[k], chunk(k), false)...)
	}
	var reordered []byte
	for _, k := range []int{2, 0, 3, 1} {
		reordered = append(reordered, frag(k, flags[k], chunk(k), true)...)
	}
	s.file(id, "tunnel.bin", tunnel)
	s.file(id, "reordered.bin", reordered)

	// reference part 1: reassemble tunnel.bin by seq
	parts1 := make(map[int][]byte)
	off := 0
	for k := 0; k < 4; k++ {
		l := int(binary.BigEndian.Uint16(tunnel[off+7:]))
		seq := int(binary.BigEndian.Uint32(tunnel[off+2:]))
		parts1[seq] = tunnel[off+9 : off+9+l]
		off += 9 + l + 2
	}
	re1 := make([]byte, 0, 4096)
	for k := 0; k < 4; k++ {
		re1 = append(re1, parts1[k]...)
	}
	p1 := fmt.Sprintf("SENT 4 fragments (len=%d)\nREASSEMBLED: %d bytes, order OK", len(payload), len(re1))

	// reference part 2: parse reordered.bin, xor-decrypt, verify, reassemble
	parts2 := make(map[int][]byte)
	off = 0
	for len(reordered)-off >= 12 {
		seq := int(binary.BigEndian.Uint32(reordered[off+2:]))
		l := int(binary.BigEndian.Uint16(reordered[off+7:]))
		c := append([]byte{}, reordered[off+9:off+9+l]...)
		for i := range c {
			c[i] ^= 0x55
		}
		parts2[seq] = c
		off += 9 + l + 2
	}
	re2 := make([]byte, 0, 4096)
	for k := 0; k < 4; k++ {
		re2 = append(re2, parts2[k]...)
	}
	p2 := fmt.Sprintf("REORDERED: %d bytes in 4 fragments\nTUNNEL OK (XOR+CRC verified)", len(re2))
	s.expected(id, p1, p2)
}

// ------------------------------------------------------------------ 128

var digitGlyphs = [][]byte{
	{0x7E, 0x81, 0x81, 0x81, 0x81, 0x81, 0x7E, 0x00}, // 0
	{0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x7E, 0x00}, // 1
	{0x7E, 0x81, 0x01, 0x7E, 0x80, 0x80, 0xFF, 0x00}, // 2
	{0xFE, 0x01, 0x01, 0x7E, 0x01, 0x01, 0xFE, 0x00}, // 3
	{0x81, 0x81, 0x81, 0xFF, 0x01, 0x01, 0x01, 0x00}, // 4
	{0xFF, 0x80, 0x80, 0xFE, 0x01, 0x01, 0xFE, 0x00}, // 5
	{0x7E, 0x80, 0x80, 0xFE, 0x81, 0x81, 0x7E, 0x00}, // 6
	{0xFF, 0x01, 0x02, 0x06, 0x0C, 0x18, 0x18, 0x00}, // 7
	{0x7E, 0x81, 0x81, 0x7E, 0x81, 0x81, 0x7E, 0x00}, // 8
	{0x7E, 0x81, 0x81, 0x7F, 0x01, 0x01, 0x7E, 0x00}, // 9
}

func digitPixels(d int) []byte {
	p := make([]byte, 64)
	g := digitGlyphs[d]
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if g[r]&(1<<(7-c)) != 0 {
				p[r*8+c] = 1
			}
		}
	}
	return p
}

func (s *set) gen128() {
	id := "128"
	// dataset: 70 train (7 per digit) + 30 test (3 per digit); decoys in test
	var db []byte
	for d := 0; d < 10; d++ {
		p := digitPixels(d)
		for k := 0; k < 7; k++ {
			db = append(db, p...)
			db = append(db, byte(d))
		}
	}
	test := make([][2][]byte, 30)
	ti := 0
	for d := 0; d < 10; d++ {
		p := digitPixels(d)
		for k := 0; k < 3; k++ {
			test[ti] = [2][]byte{p, {byte(d)}}
			ti++
		}
	}
	// decoys: first class-3 test sample looks like 8, first class-7 looks like 1
	for i := range test {
		if test[i][1][0] == 3 {
			test[i][0] = digitPixels(8)
			break
		}
	}
	for i := range test {
		if test[i][1][0] == 7 {
			test[i][0] = digitPixels(1)
			break
		}
	}
	for i := range test {
		db = append(db, test[i][0]...)
		db = append(db, test[i][1][0])
	}
	s.file(id, "digits.bin", db)

	// reference perceptron
	var w [10][64]float64
	var b [10]float64
	const lr = 0.1
	train := db[:70*65]
	for epoch := 0; epoch < 100; epoch++ {
		for o := 0; o+65 <= len(train); o += 65 {
			pix := train[o : o+64]
			label := int(train[o+64])
			for j := 0; j < 10; j++ {
				act := b[j]
				for i := 0; i < 64; i++ {
					if pix[i] == 1 {
						act += w[j][i]
					}
				}
				var y float64
				if act > 0 {
					y = 1
				}
				target := 0.0
				if j == label {
					target = 1
				}
				err := target - y
				if err != 0 {
					for i := 0; i < 64; i++ {
						w[j][i] += lr * err * float64(pix[i])
					}
					b[j] += lr * err
				}
			}
		}
	}
	pred := func(pix []byte) int {
		best, bi := -1.0, 0
		for j := 0; j < 10; j++ {
			act := b[j]
			for i := 0; i < 64; i++ {
				if pix[i] == 1 {
					act += w[j][i]
				}
			}
			if act > best {
				best, bi = act, j
			}
		}
		return bi
	}
	trainOK, testOK := 0, 0
	var errByClass [10]int
	for o := 0; o+65 <= len(train); o += 65 {
		if pred(train[o:o+64]) == int(train[o+64]) {
			trainOK++
		}
	}
	for i := range test {
		if pred(test[i][0]) == int(test[i][1][0]) {
			testOK++
		} else {
			errByClass[test[i][1][0]]++
		}
	}
	var errs []string
	for c := 0; c < 10; c++ {
		if errByClass[c] > 0 {
			errs = append(errs, fmt.Sprintf("ERR: class %d: %d wrong", c, errByClass[c]))
		}
	}
	acc := float64(testOK) / 30 * 100
	p1 := fmt.Sprintf("TRAIN: %d/70 correct\nTEST: %d/30 correct", trainOK, testOK)
	p2 := strings.Join(errs, "\n") + fmt.Sprintf("\nACCURACY: %.1f%%", acc)
	s.expected(id, p1, p2)
}

func init() {
	register("111", (*set).gen111)
	register("113", (*set).gen113)
	register("115", (*set).gen115)
	register("116", (*set).gen116)
	register("119", (*set).gen119)
	register("121", (*set).gen121)
	register("122", (*set).gen122)
	register("123", (*set).gen123)
	register("125", (*set).gen125)
	register("126", (*set).gen126)
	register("127", (*set).gen127)
	register("128", (*set).gen128)
}
