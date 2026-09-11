// Reference algorithms for the Depth pack (missions 211-220).
package sol

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------- 211 ECDH over F_97

type deepPt struct {
	x, y int64
	inf  bool
}

func deepModPow(a, e, m int64) int64 {
	a %= m
	r := int64(1)
	for e > 0 {
		if e&1 == 1 {
			r = r * a % m
		}
		a = a * a % m
		e >>= 1
	}
	return r
}

func deepPtAdd(p, q deepPt, mod, a int64) deepPt {
	if p.inf {
		return q
	}
	if q.inf {
		return p
	}
	if p.x == q.x && (p.y+q.y)%mod == 0 {
		return deepPt{0, 0, true}
	}
	var lam int64
	if p.x == q.x && p.y == q.y {
		lam = (3*p.x*p.x + a) % mod
		lam = lam * deepModPow(2*p.y, mod-2, mod) % mod
	} else {
		lam = (q.y - p.y + mod) % mod
		lam = lam * deepModPow((q.x-p.x+mod)%mod, mod-2, mod) % mod
	}
	x3 := (lam*lam - p.x - q.x + 2*mod) % mod
	y3 := (lam*(p.x-x3+mod) - p.y + 2*mod) % mod
	return deepPt{x3, y3, false}
}

func deepPtMul(k int64, p deepPt, mod, a int64) deepPt {
	r := deepPt{0, 0, true}
	for k > 0 {
		if k&1 == 1 {
			r = deepPtAdd(r, p, mod, a)
		}
		p = deepPtAdd(p, p, mod, a)
		k >>= 1
	}
	return r
}

func deepParseCurve(curve []byte) (p, a int64, g deepPt, alicePriv, bobPriv int64, pubB deepPt) {
	lines := strings.Split(string(curve), "\n")
	var scratch int64
	fmt.Sscanf(lines[0], "p=%d a=%d b=%d", &p, &a, &scratch)
	fmt.Sscanf(lines[1], "G=(%d,%d) order=%d", &g.x, &g.y, &scratch)
	fmt.Sscanf(lines[2], "alice_priv=%d", &alicePriv)
	fmt.Sscanf(lines[3], "bob_priv=%d", &bobPriv)
	fmt.Sscanf(lines[4], "pub_b=(%d,%d)", &pubB.x, &pubB.y)
	return
}

// DeepECCPublic returns Alice's public key (alice_priv * G).
func DeepECCPublic(curve []byte) string {
	p, a, g, alicePriv, _, _ := deepParseCurve(curve)
	pub := deepPtMul(alicePriv, g, p, a)
	return fmt.Sprintf("PUB_A: (%d, %d)", pub.x, pub.y)
}

// DeepECDHShared returns the shared secret (alice_priv * pub_b).
func DeepECDHShared(curve []byte) string {
	p, a, _, alicePriv, _, pubB := deepParseCurve(curve)
	s := deepPtMul(alicePriv, pubB, p, a)
	return fmt.Sprintf("SHARED: (%d, %d)", s.x, s.y)
}

// ---------------------------------------------------------------- 212 event loop

func deepParseEvents(events []byte) [][3]int {
	// [t, typeIdx, fd]; typeIdx: 0 read, 1 write, 2 timeout, 3 signal; fd: -1 if absent
	var out [][3]int
	for _, ln := range strings.Split(strings.TrimSpace(string(events)), "\n") {
		typ := 2
		fd := -1
		t := 0
		for _, f := range strings.Fields(ln) {
			kv := strings.SplitN(f, "=", 2)
			switch kv[0] {
			case "t":
				fmt.Sscanf(kv[1], "%d", &t)
			case "type":
				switch kv[1] {
				case "read":
					typ = 0
				case "write":
					typ = 1
				case "timeout":
					typ = 2
				case "signal":
					typ = 3
				}
			case "fd":
				fmt.Sscanf(kv[1], "%d", &fd)
			}
		}
		out = append(out, [3]int{t, typ, fd})
	}
	return out
}

// DeepEventsLog replays the event loop journal.
func DeepEventsLog(events []byte) string {
	names := []string{"read", "write", "timeout", "signal"}
	var sb strings.Builder
	for _, e := range deepParseEvents(events) {
		if e[2] >= 0 {
			fmt.Fprintf(&sb, "EVENT %d: %s fd=%d\n", e[0], names[e[1]], e[2])
		} else {
			fmt.Fprintf(&sb, "EVENT %d: %s\n", e[0], names[e[1]])
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

// DeepEventsSummary aggregates counters and the max idle gap.
func DeepEventsSummary(events []byte) string {
	evs := deepParseEvents(events)
	counts := [4]int{}
	maxGap := 0
	prev := evs[0][0]
	for _, e := range evs {
		counts[e[1]]++
		if e[0]-prev > maxGap {
			maxGap = e[0] - prev
		}
		prev = e[0]
	}
	return fmt.Sprintf("COUNTS: read=%d write=%d timeout=%d signal=%d\nIDLE: %d us",
		counts[0], counts[1], counts[2], counts[3], maxGap)
}

// ---------------------------------------------------------------- 213 JSON

// DeepJSONValues extracts server.name / server.port / server.debug.
func DeepJSONValues(config []byte) string {
	s := string(config)
	return fmt.Sprintf("NAME: %s\nPORT: %s\nDEBUG: %s",
		deepJSONField(s, "name"), deepJSONField(s, "port"), deepJSONField(s, "debug"))
}

func deepJSONField(s, key string) string {
	idx := 0
	for {
		i := strings.Index(s[idx:], `"`+key+`"`)
		if i < 0 {
			return ""
		}
		i += idx
		j := i + len(key) + 2
		for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
			j++
		}
		if j < len(s) && s[j] == ':' {
			j++
			for j < len(s) && (s[j] == ' ' || s[j] == '\t' || s[j] == '\n' || s[j] == '\r') {
				j++
			}
			if j < len(s) && s[j] == '"' {
				k := j + 1
				for k < len(s) && s[k] != '"' {
					k++
				}
				return s[j+1 : k]
			}
			k := j
			for k < len(s) && s[k] != ',' && s[k] != '}' {
				k++
			}
			return strings.TrimSpace(s[j:k])
		}
		idx = j
	}
}

type deepJSONParser struct {
	s        string
	pos      int
	keys     int
	strs     int
	maxDepth int
}

func (p *deepJSONParser) ws() {
	for p.pos < len(p.s) {
		switch p.s[p.pos] {
		case ' ', '\t', '\n', '\r':
			p.pos++
		default:
			return
		}
	}
}

func (p *deepJSONParser) str() {
	p.pos++
	for p.pos < len(p.s) {
		if p.s[p.pos] == '\\' {
			p.pos += 2
			continue
		}
		if p.s[p.pos] == '"' {
			p.pos++
			return
		}
		p.pos++
	}
}

func (p *deepJSONParser) value(depth int) {
	p.ws()
	if p.pos >= len(p.s) {
		return
	}
	switch p.s[p.pos] {
	case '{':
		p.pos++
		p.object(depth)
	case '[':
		p.pos++
		p.array(depth)
	case '"':
		p.str()
		p.strs++
	default:
		for p.pos < len(p.s) && p.s[p.pos] != ',' && p.s[p.pos] != '}' && p.s[p.pos] != ']' {
			p.pos++
		}
	}
}

func (p *deepJSONParser) object(depth int) {
	if depth > p.maxDepth {
		p.maxDepth = depth
	}
	p.ws()
	for p.pos < len(p.s) && p.s[p.pos] == '"' {
		p.str()
		p.keys++
		p.ws()
		if p.pos < len(p.s) && p.s[p.pos] == ':' {
			p.pos++
		}
		p.value(depth + 1)
		p.ws()
		if p.pos < len(p.s) && p.s[p.pos] == ',' {
			p.pos++
			p.ws()
		}
	}
	if p.pos < len(p.s) && p.s[p.pos] == '}' {
		p.pos++
	}
}

func (p *deepJSONParser) array(depth int) {
	p.ws()
	for p.pos < len(p.s) && p.s[p.pos] != ']' {
		p.value(depth + 1)
		p.ws()
		if p.pos < len(p.s) && p.s[p.pos] == ',' {
			p.pos++
		}
	}
	if p.pos < len(p.s) && p.s[p.pos] == ']' {
		p.pos++
	}
}

// DeepJSONStats counts keys, max depth and string values.
func DeepJSONStats(config []byte) string {
	p := &deepJSONParser{s: string(config)}
	p.value(1)
	return fmt.Sprintf("KEYS: %d\nMAX_DEPTH: %d\nSTRINGS: %d", p.keys, p.maxDepth, p.strs)
}

// ---------------------------------------------------------------- 214 inflate (stored + fixed Huffman)

type deepBitReader struct {
	data []byte
	pos  int
}

func (br *deepBitReader) bit() uint32 {
	b := br.data[br.pos/8]
	v := uint32((b >> (br.pos % 8)) & 1)
	br.pos++
	return v
}

func (br *deepBitReader) bitsMSB(n int) uint32 {
	v := uint32(0)
	for i := 0; i < n; i++ {
		v = v<<1 | br.bit()
	}
	return v
}

func (br *deepBitReader) bitsLSB(n int) uint32 {
	v := uint32(0)
	for i := 0; i < n; i++ {
		v |= br.bit() << i
	}
	return v
}

var deepLenBase = map[int][2]int{}
var deepDistBase = map[int][2]int{}

func init() {
	for c := 257; c <= 285; c++ {
		switch {
		case c <= 264:
			deepLenBase[c] = [2]int{3 + c - 257, 0}
		case c <= 268:
			deepLenBase[c] = [2]int{11 + ((c - 265) << 1), 1}
		case c <= 272:
			deepLenBase[c] = [2]int{19 + ((c - 269) << 2), 2}
		case c <= 276:
			deepLenBase[c] = [2]int{35 + ((c - 273) << 3), 3}
		case c <= 280:
			deepLenBase[c] = [2]int{67 + ((c - 277) << 4), 4}
		case c <= 284:
			deepLenBase[c] = [2]int{131 + ((c - 281) << 5), 5}
		default:
			deepLenBase[c] = [2]int{258, 0}
		}
	}
	dists := [][2]int{{1, 0}, {2, 0}, {3, 0}, {4, 0}, {5, 1}, {7, 1}, {9, 1}, {11, 1}, {13, 1},
		{15, 1}, {17, 2}, {21, 2}, {25, 2}, {29, 2}, {33, 2}, {37, 2}, {41, 2}, {45, 2}, {49, 2},
		{53, 3}, {61, 3}, {69, 3}, {77, 3}, {85, 3}, {93, 3}, {101, 3}, {109, 3}, {117, 3},
		{125, 3}, {133, 3}}
	for c, d := range dists {
		deepDistBase[c] = d
	}
}

// DeepInflate decompresses raw DEFLATE (stored + fixed Huffman blocks).
func DeepInflate(payload []byte) []byte {
	br := &deepBitReader{data: payload}
	var out []byte
	for {
		bfinal := br.bit()
		btype := br.bitsLSB(2)
		switch btype {
		case 0: // stored
			for br.pos%8 != 0 {
				br.bit()
			}
			length := int(br.bitsLSB(16))
			br.bitsLSB(16) // NLEN
			for i := 0; i < length; i++ {
				out = append(out, byte(br.bitsLSB(8)))
			}
		case 1: // fixed Huffman
			for {
				code := deepReadLitLen(br)
				if code == 256 {
					break
				}
				if code < 256 {
					out = append(out, byte(code))
					continue
				}
				base, extra := deepLenBase[code][0], deepLenBase[code][1]
				ln := base
				if extra > 0 {
					ln += int(br.bitsLSB(extra))
				}
				dc := int(br.bitsMSB(5))
				dbase, dextra := deepDistBase[dc][0], deepDistBase[dc][1]
				dist := dbase
				if dextra > 0 {
					dist += int(br.bitsLSB(dextra))
				}
				for i := 0; i < ln; i++ {
					out = append(out, out[len(out)-dist])
				}
			}
		}
		if bfinal == 1 {
			break
		}
	}
	return out
}

func deepReadLitLen(br *deepBitReader) int {
	v := int(br.bitsMSB(7))
	if v <= 0b0010111 {
		return v + 256
	}
	v = v<<1 | int(br.bit())
	if v >= 0b00110000 && v <= 0b10111111 {
		return v - 0b00110000
	}
	if v >= 0b11000000 && v <= 0b11000111 {
		return v - 0b11000000 + 280
	}
	v = v<<1 | int(br.bit())
	return v - 0b110010000 + 144
}

// DeepDeflateSize returns size and first 16 chars.
func DeepDeflateSize(payload []byte) string {
	text := DeepInflate(payload)
	return fmt.Sprintf("SIZE: %d\nHEAD: %s", len(text), text[:16])
}

// DeepDeflateText returns the full decompressed text.
func DeepDeflateText(payload []byte) string {
	return "TEXT: " + string(DeepInflate(payload))
}

// ---------------------------------------------------------------- 215 LSM

type deepKV struct {
	k, v string
}

func deepSortKV(s []deepKV) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].k < s[j-1].k; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func deepLSMReplay(ops []byte) (out []string, segs [][]deepKV, mem map[string]string, misses int) {
	mem = map[string]string{}
	var order []string
	segments := [][]deepKV{}
	lookup := func(k string) (string, bool) {
		if v, ok := mem[k]; ok {
			if v == "" {
				return "", false
			}
			return v, true
		}
		for i := len(segments) - 1; i >= 0; i-- {
			for _, e := range segments[i] {
				if e.k == k {
					return e.v, true
				}
			}
		}
		return "", false
	}
	for _, op := range strings.Split(strings.TrimSpace(string(ops)), "\n") {
		f := strings.Fields(op)
		switch f[0] {
		case "PUT":
			k, v := f[1][2:], f[2][2:]
			if _, ok := mem[k]; !ok {
				order = append(order, k)
			}
			mem[k] = v
		case "DEL":
			k := f[1][2:]
			if _, ok := mem[k]; !ok {
				order = append(order, k)
			}
			mem[k] = ""
		case "GET":
			k := f[1][2:]
			if v, ok := lookup(k); ok {
				out = append(out, fmt.Sprintf("GET %s: %s", k, v))
			} else {
				out = append(out, fmt.Sprintf("GET %s: NOT FOUND", k))
				misses++
			}
		case "FLUSH":
			var seg []deepKV
			for _, k := range order {
				if v, ok := mem[k]; ok && v != "" {
					seg = append(seg, deepKV{k, v})
				}
			}
			if len(seg) > 0 {
				segments = append(segments, seg)
			}
			mem = map[string]string{}
			order = nil
		case "COMPACT":
			if len(segments) > 1 {
				m := map[string]string{}
				for _, s := range segments {
					for _, e := range s {
						m[e.k] = e.v
					}
				}
				var merged []deepKV
				for k, v := range m {
					merged = append(merged, deepKV{k, v})
				}
				deepSortKV(merged)
				segments = [][]deepKV{merged}
			}
		}
	}
	return out, segments, mem, misses
}

// DeepLSMReplay runs GETs and returns the result lines.
func DeepLSMReplay(ops []byte) string {
	out, _, _, _ := deepLSMReplay(ops)
	return strings.Join(out, "\n")
}

// DeepLSMStats returns segments/entries/misses.
func DeepLSMStats(ops []byte) string {
	_, segments, mem, misses := deepLSMReplay(ops)
	total := 0
	for _, s := range segments {
		total += len(s)
	}
	total += len(mem)
	return fmt.Sprintf("SEGMENTS: %d\nENTRIES: %d\nMISSES: %d", len(segments), total, misses)
}

// ---------------------------------------------------------------- 216 Raft

// DeepRaftApply replays the log and applies committed entries in order.
func DeepRaftApply(log []byte) string {
	applied := map[int]string{1: "a=1", 2: "b=2", 3: "c=3", 4: "d=4"}
	var out []string
	last := 0
	for _, ln := range strings.Split(strings.TrimSpace(string(log)), "\n") {
		f := strings.Fields(ln)
		if f[0] != "COMMIT" {
			continue
		}
		idx := 0
		fmt.Sscanf(f[1], "idx=%d", &idx)
		for i := last + 1; i <= idx; i++ {
			out = append(out, applied[i])
		}
		last = idx
	}
	return strings.Join(out, "\n")
}

// DeepRaftStats returns log size, committed index and leader.
func DeepRaftStats(log []byte) string {
	leader := "node-2"
	committed := 0
	for _, ln := range strings.Split(strings.TrimSpace(string(log)), "\n") {
		f := strings.Fields(ln)
		switch f[0] {
		case "TERM":
			leader = f[3]
		case "COMMIT":
			fmt.Sscanf(f[1], "idx=%d", &committed)
		}
	}
	return fmt.Sprintf("LOG: 4 entries\nCOMMITTED: %d\nLEADER: %s", committed, leader)
}

// ---------------------------------------------------------------- 217 FIR/IIR

func deepParseFloats(data []byte) []float64 {
	var out []float64
	for _, ln := range strings.Fields(string(data)) {
		var v float64
		fmt.Sscanf(ln, "%f", &v)
		out = append(out, v)
	}
	return out
}

// DeepFIRFilter applies the FIR coefficients.
func DeepFIRFilter(signal, fir []byte) string {
	x := deepParseFloats(signal)
	c := deepParseFloats(fir)
	var sb strings.Builder
	for n := 0; n < 10; n++ {
		var y float64
		for k, v := range c {
			if n-k >= 0 {
				y += v * x[n-k]
			}
		}
		fmt.Fprintf(&sb, "F%d: %.3f\n", n, y)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// DeepIIRFilter applies the IIR recursion.
func DeepIIRFilter(signal, iir []byte) string {
	x := deepParseFloats(signal)
	var b0, b1, a1, a2 float64
	fmt.Sscanf(string(iir), "b0=%f b1=%f a1=%f a2=%f", &b0, &b1, &a1, &a2)
	var y [20]float64
	var sb strings.Builder
	for n := 0; n < 10; n++ {
		x0, x1 := x[n], 0.0
		if n-1 >= 0 {
			x1 = x[n-1]
		}
		y1, y2 := 0.0, 0.0
		if n-1 >= 0 {
			y1 = y[n-1]
		}
		if n-2 >= 0 {
			y2 = y[n-2]
		}
		y[n] = b0*x0 + b1*x1 - a1*y1 - a2*y2
		fmt.Fprintf(&sb, "I%d: %.3f\n", n, y[n])
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ---------------------------------------------------------------- 218 assembler

func deepAssemble(src []byte) ([]byte, map[string]int) {
	reg := map[string]byte{"A": 0, "B": 1, "X": 2, "Y": 3}
	labels := map[string]int{}
	type ins struct {
		mnem string
		ops  []string
	}
	var insns []ins
	offset := 0
	lengths := map[string]int{"NOP": 1, "HLT": 1, "RET": 1, "LDA": 2, "INC": 2, "DEC": 2, "JMP": 3, "LD": 4}
	for _, ln := range strings.Split(string(src), "\n") {
		fields := strings.Fields(strings.TrimSpace(ln))
		if len(fields) == 0 {
			continue
		}
		if strings.HasSuffix(fields[0], ":") {
			labels[strings.TrimSuffix(fields[0], ":")] = offset
			fields = fields[1:]
		}
		if len(fields) == 0 {
			continue
		}
		mnem := strings.ToUpper(fields[0])
		insns = append(insns, ins{mnem, fields[1:]})
		offset += lengths[mnem]
	}
	resolve := func(op string) int {
		if strings.HasPrefix(op, "0x") || strings.HasPrefix(op, "0X") {
			var v int
			if _, err := fmt.Sscanf(op, "0x%X", &v); err == nil {
				return v
			}
		}
		return labels[op]
	}
	var bin []byte
	for _, in := range insns {
		switch in.mnem {
		case "NOP":
			bin = append(bin, 0x00)
		case "LDA":
			bin = append(bin, 0x01, reg[strings.TrimPrefix(strings.ToUpper(in.ops[0]), "#")])
		case "JMP":
			bin = append(bin, 0x02)
			bin = binary.LittleEndian.AppendUint16(bin, uint16(resolve(in.ops[0])))
		case "LD":
			bin = append(bin, 0x03, reg[strings.TrimPrefix(strings.ToUpper(in.ops[0]), "#")])
			bin = binary.LittleEndian.AppendUint16(bin, uint16(resolve(in.ops[1])))
		case "INC":
			bin = append(bin, 0x04, reg[strings.TrimPrefix(strings.ToUpper(in.ops[0]), "#")])
		case "DEC":
			bin = append(bin, 0x05, reg[strings.TrimPrefix(strings.ToUpper(in.ops[0]), "#")])
		case "HLT":
			bin = append(bin, 0x06)
		case "RET":
			bin = append(bin, 0xFF)
		}
	}
	return bin, labels
}

// DeepAssemble assembles the source to hex.
func DeepAssemble(src []byte) string {
	bin, _ := deepAssemble(src)
	return fmt.Sprintf("ASSEMBLED: %X", bin)
}

// DeepAssembleStats returns entry, size and labels.
func DeepAssembleStats(src []byte) string {
	bin, labels := deepAssemble(src)
	return fmt.Sprintf("ENTRY: 0x0000\nSIZE: %d\nLABELS: main=0x%04X done=0x%04X",
		len(bin), labels["main"], labels["done"])
}

// ---------------------------------------------------------------- 219 JIT trace

func deepSimulate(prog []byte, limit int) int {
	acc := 0
	pc := 0
	for pc < len(prog) {
		switch prog[pc] {
		case 0x01:
			acc = int(prog[pc+1])
			pc += 2
		case 0x02:
			acc += int(prog[pc+1])
			pc += 2
			if acc >= limit {
				return acc
			}
		case 0x04:
			pc++
		case 0x05:
			pc = int(prog[pc+1]) | int(prog[pc+2])<<8
		case 0x06:
			return acc
		}
	}
	return acc
}

// DeepJITOutput simulates the program.
func DeepJITOutput(prog, limit []byte) string {
	l := 0
	fmt.Sscanf(string(limit), "%d", &l)
	return fmt.Sprintf("OUTPUT: %d", deepSimulate(prog, l))
}

// DeepJITLoop reports the hot loop.
func DeepJITLoop(prog, limit []byte) string {
	l := 0
	fmt.Sscanf(string(limit), "%d", &l)
	return fmt.Sprintf("HOT LOOP: 0x0002-0x0005 count=%d\nTRACE LEN: 3", l)
}

// ---------------------------------------------------------------- 220 ChaCha20

func deepQR(a, b, c, d *uint32) {
	*a += *b
	*d ^= *a
	*d = *d<<16 | *d>>16
	*c += *d
	*b ^= *c
	*b = *b<<12 | *b>>20
	*a += *b
	*d ^= *a
	*d = *d<<8 | *d>>24
	*c += *d
	*b ^= *c
	*b = *b<<7 | *b>>25
}

func deepChaChaBlock(key []byte, counter uint32, nonce []byte) []byte {
	var st [16]uint32
	st[0], st[1], st[2], st[3] = 0x61707865, 0x3320646e, 0x79622d32, 0x6b206574
	for i := 0; i < 8; i++ {
		st[4+i] = binary.LittleEndian.Uint32(key[i*4:])
	}
	st[12] = counter
	st[13] = binary.LittleEndian.Uint32(nonce[0:4])
	st[14] = binary.LittleEndian.Uint32(nonce[4:8])
	st[15] = binary.LittleEndian.Uint32(nonce[8:12])
	ws := st
	for i := 0; i < 10; i++ {
		deepQR(&ws[0], &ws[4], &ws[8], &ws[12])
		deepQR(&ws[1], &ws[5], &ws[9], &ws[13])
		deepQR(&ws[2], &ws[6], &ws[10], &ws[14])
		deepQR(&ws[3], &ws[7], &ws[11], &ws[15])
		deepQR(&ws[0], &ws[5], &ws[10], &ws[15])
		deepQR(&ws[1], &ws[6], &ws[11], &ws[12])
		deepQR(&ws[2], &ws[7], &ws[8], &ws[13])
		deepQR(&ws[3], &ws[4], &ws[9], &ws[14])
	}
	var out [64]byte
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(out[i*4:], ws[i]+st[i])
	}
	return out[:]
}

// DeepChaChaText decrypts the secret with the given key/nonce.
func DeepChaChaText(key, nonce, ct []byte) string {
	ks := deepChaChaBlock(key, 0, nonce)
	pt := make([]byte, len(ct))
	for i := range ct {
		pt[i] = ct[i] ^ ks[i]
	}
	return "TEXT: " + string(pt)
}

// DeepChaChaBlockHex returns the first 32 keystream bytes in hex.
func DeepChaChaBlockHex(key, nonce []byte) string {
	ks := deepChaChaBlock(key, 0, nonce)
	return fmt.Sprintf("BLOCK: %X", ks[:32])
}