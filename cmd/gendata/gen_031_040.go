package main

// Generators for missions 031-040 (crypto & encoding part 2). All generators
// for this group self-register below.

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
	"time"

	"neon-grid/sol"
)

func (s *set) gen031() {
	id := "031"
	plain := []string{"MERCURY", "NEON", "M", "ME"}
	var in strings.Builder
	for _, p := range plain {
		fmt.Fprintf(&in, "encode: %s\n", p)
	}
	for _, p := range plain {
		fmt.Fprintf(&in, "decode: %s\n", sol.Base64Encode([]byte(p)))
	}
	s.text(id, "input.txt", in.String())

	var p1 strings.Builder
	for _, p := range plain {
		p1.WriteString(sol.Base64Encode([]byte(p)))
		p1.WriteByte('\n')
	}
	var p2 strings.Builder
	for _, p := range plain {
		fmt.Fprintf(&p2, "PLAINTEXT: %s\n", p)
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen032() {
	id := "032"
	var body []byte
	body = append(body, []byte("NEON GRID RLE BLIN")...)
	body = append(body, 0x07, 0x07, 0x07, 0x02, 0x02)
	body = append(body, bytes.Repeat([]byte{0xAA}, 4096)...)
	body = append(body, bytes.Repeat([]byte{0x00}, 70000)...)
	comp := sol.RLEPack(body)
	s.file(id, "compressed.bin", comp)

	first := body
	if len(first) > 64 {
		first = first[:64]
	}
	p1 := fmt.Sprintf("LEN: %d\nDATA: %s", len(body), hexStr(first))

	packed := sol.RLEPack(body)
	p2 := fmt.Sprintf("PACKED: %s\nROUNDTRIP OK", hexStr(packed))
	s.expected(id, p1, p2)
}

func (s *set) gen033() {
	id := "033"
	s.text(id, "freqs.txt", "freqs: a=45 b=13 c=12 d=16 e=9 f=5\n")
	freqs := map[byte]int{'a': 45, 'b': 13, 'c': 12, 'd': 16, 'e': 9, 'f': 5}
	codes := sol.HuffmanCodes(freqs)
	var p1 strings.Builder
	for _, ch := range []byte("abcdef") {
		fmt.Fprintf(&p1, "'%c': %s\n", ch, codes[ch])
	}

	msg := "RAVENS PACK THEIR BASES WITH HUFFMAN TREES"
	f2 := map[byte]int{}
	for i := 0; i < len(msg); i++ {
		f2[msg[i]]++
	}
	c2 := sol.HuffmanCodes(f2)
	lens := map[byte]int{}
	for ch, c := range c2 {
		lens[ch] = len(c)
	}
	canon := sol.CanonicalHuffmanCodes(lens)

	type pair struct {
		sym byte
		ln  int
	}
	var ps []pair
	for ch, l := range lens {
		ps = append(ps, pair{ch, l})
	}
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].ln != ps[j].ln {
			return ps[i].ln < ps[j].ln
		}
		return ps[i].sym < ps[j].sym
	})

	var arch []byte
	var nb [4]byte
	u32be(nb[:], uint32(len(ps)))
	arch = append(arch, nb[:]...)
	for _, p := range ps {
		arch = append(arch, p.sym, byte(p.ln))
	}
	u32be(nb[:], uint32(len(msg)))
	arch = append(arch, nb[:]...)
	var bw sol.BitWriter
	for i := 0; i < len(msg); i++ {
		c := canon[msg[i]]
		for j := 0; j < len(c); j++ {
			bw.Write(uint64(c[j]-'0'), 1)
		}
	}
	arch = append(arch, bw.Bytes()...)
	s.file(id, "archive.bin", arch)

	p2 := "DECODED: " + msg
	s.expected(id, strings.TrimRight(p1.String(), "\n"), p2)
}

func (s *set) gen034() {
	id := "034"
	plain := "TOBEORNOTTOBEORTOBEORNOT"
	codes := sol.LZWEncode([]byte(plain), 4096)
	var decLine strings.Builder
	for i, c := range codes {
		if i > 0 {
			decLine.WriteByte(' ')
		}
		fmt.Fprintf(&decLine, "%d", c)
	}
	s.text(id, "input.txt", fmt.Sprintf("encode: %s\ndecode: %s\n", plain, decLine.String()))

	var p1 strings.Builder
	p1.WriteString("CODES: ")
	for i, c := range codes {
		if i > 0 {
			p1.WriteByte(' ')
		}
		fmt.Fprintf(&p1, "%d", c)
	}
	dec := string(sol.LZWDecode(codes))
	p2 := "DECODED: " + dec
	s.expected(id, p1.String(), p2)
}

func (s *set) gen035() {
	id := "035"
	plain := "NEON NEON NEON GRID GRID GRID LZ77 LZ77 LZ77 VACUUM VACUUM VACUUM"
	packed := lz77Enc([]byte(plain))
	s.file(id, "packed.bin", packed)
	p1 := "DECODED: " + plain

	pattern := []byte("neon grid lz77 vacuum ")
	var raw []byte
	for len(raw) < 1234 {
		raw = append(raw, pattern...)
	}
	raw = append(raw, 'O', 'K')
	raw = raw[:1234]
	s.file(id, "raw.bin", raw)
	pack2 := lz77Enc(raw)
	p2 := fmt.Sprintf("PACKED: %s\nROUNDTRIP OK (input=%d output=%d)",
		hexStr(pack2), len(raw), len(lz77Dec(pack2)))
	s.expected(id, p1, p2)
}

func (s *set) gen036() {
	id := "036"
	hdrs := [][4]uint64{
		{7, 1, 0xA5, 0x0102},
		{5, 13, 0x80, 0xBEEF},
		{1, 31, 0x01, 0x0000},
	}
	var packet []byte
	var p1 strings.Builder
	for _, h := range hdrs {
		packet = append(packet, packHeader(h[0], h[1], h[2], h[3])...)
		br := sol.NewBitReader(packet[len(packet)-4:])
		v, _ := br.Read(3)
		t, _ := br.Read(5)
		f, _ := br.Read(8)
		l, _ := br.Read(16)
		fmt.Fprintf(&p1, "VERSION=%d TYPE=%d FLAGS=0x%02X LENGTH=%d\n", v, t, f, l)
	}
	s.file(id, "packet.bin", packet)

	var p2 strings.Builder
	for i := 0; i+4 <= len(packet); i += 4 {
		fmt.Fprintf(&p2, "WRITTEN: %s\n", hexStr(packet[i:i+4]))
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen037() {
	id := "037"
	s.text(id, "input.txt", "NEON\nGRID\n")
	p1 := fmt.Sprintf("NEON: 0x%08X\nGRID: 0x%08X",
		sol.CRC32IEEE([]byte("NEON")), sol.CRC32IEEE([]byte("GRID")))

	blob := make([]byte, 65536)
	for i := range blob {
		blob[i] = byte(i*13 + 7)
	}
	s.file(id, "blob.bin", blob)
	corrupted := append([]byte{}, blob...)
	corrupted[len(corrupted)-1] ^= 0x01
	p2 := fmt.Sprintf("TABLE CRC: 0x%08X\nBITWISE MATCH: OK\nCORRUPTED CRC: 0x%08X",
		sol.CRC32IEEE(blob), sol.CRC32IEEE(corrupted))
	s.expected(id, p1, p2)
}

func (s *set) gen038() {
	id := "038"
	okFrame := []byte{0x7E, 0x03, 0x00, 0x41, 0x42, 0x43, 0x40, 0x7E}
	badFrame := []byte{0x7E, 0x02, 0x00, 0x41, 0x42, 0x55, 0x7E}
	s.file(id, "frame.bin", append(append([]byte{}, okFrame...), badFrame...))

	var p1 strings.Builder
	for _, f := range [][]byte{okFrame, badFrame} {
		if _, _, ok := sol.Frame(f); ok {
			p1.WriteString("FRAME OK\n")
		} else {
			p1.WriteString("FRAME BAD\n")
		}
	}

	frames := [][]byte{
		{0x7E, 0x06, 0x00, 0x48, 0x45, 0x4C, 0x4C, 0x4F, 0xFF, 0x42, 0x7E},
		{0x7E, 0x04, 0x00, 0x47, 0x52, 0x49, 0x44, 0x18, 0x7E},
		{0x7E, 0x04, 0x00, 0x41, 0x42, 0x43, 0x44, 0x55, 0x7E},
	}
	var all []byte
	for _, f := range frames {
		all = append(all, f...)
	}
	s.file(id, "frames.bin", all)

	var p2 strings.Builder
	for _, f := range frames {
		data, fixed, ok := sol.Frame(f)
		switch {
		case ok && fixed:
			fmt.Fprintf(&p2, "FIXED: %s\n", data)
		case ok:
			fmt.Fprintf(&p2, "OK: %s\n", data)
		default:
			fmt.Fprintf(&p2, "CORRUPT: %s\n", hexStr(data))
		}
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen039() {
	id := "039"
	pal := [][3]int{
		{0, 0, 0}, {255, 0, 0}, {0, 255, 0}, {0, 0, 255},
		{255, 255, 0}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
		{128, 128, 128}, {128, 0, 0}, {0, 128, 0}, {0, 0, 128},
		{128, 128, 0}, {128, 0, 128}, {0, 128, 128}, {64, 64, 64},
	}
	nearest := func(r, g, b int) int {
		best, bestD := 0, int(^uint(0)>>1)
		for i, c := range pal {
			d := (r-c[0])*(r-c[0]) + (g-c[1])*(g-c[1]) + (b-c[2])*(b-c[2])
			if d < bestD {
				best, bestD = i, d
			}
		}
		return best
	}
	inputs := [][3]int{{255, 0, 128}, {0, 255, 0}, {12, 12, 12}, {0, 128, 0}}
	var in strings.Builder
	for _, c := range inputs {
		fmt.Fprintf(&in, "%d %d %d\n", c[0], c[1], c[2])
	}
	s.text(id, "colors.txt", in.String())

	var p1 strings.Builder
	for _, c := range inputs {
		fmt.Fprintf(&p1, "%d %d %d  -> %d\n", c[0], c[1], c[2], nearest(c[0], c[1], c[2]))
	}

	w, h := 32, 16
	rowBytes := (w*3 + 3) &^ 3
	bmp := make([]byte, 54+rowBytes*h)
	copy(bmp[0:2], "BM")
	u32le(bmp[2:], uint32(len(bmp)))
	u32le(bmp[10:], 54)
	u32le(bmp[14:], 40)
	u32le(bmp[18:], uint32(w))
	u32le(bmp[22:], uint32(h))
	u16le(bmp[26:], 1)
	u16le(bmp[28:], 24)
	u32le(bmp[34:], uint32(rowBytes*h))
	pix := func(x, y int) (int, int, int) { return 40 + 6*x, 30 + 8*y, 20 + 4*(x+y) }
	for y := 0; y < h; y++ {
		row := bmp[54+y*rowBytes:]
		for x := 0; x < w; x++ {
			r, g, b := pix(x, y)
			row[x*3+0] = byte(b)
			row[x*3+1] = byte(g)
			row[x*3+2] = byte(r)
		}
	}
	s.file(id, "image.bmp", bmp)

	var q []byte
	var sum int64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x += 2 {
			r1, g1, b1 := pix(x, y)
			idx1 := nearest(r1, g1, b1)
			r2, g2, b2 := pix(x+1, y)
			idx2 := nearest(r2, g2, b2)
			q = append(q, byte(idx1<<4|idx2))
			for _, px := range [][3]int{{r1, g1, b1}, {r2, g2, b2}} {
				idx := nearest(px[0], px[1], px[2])
				dr := px[0] - pal[idx][0]
				dg := px[1] - pal[idx][1]
				db := px[2] - pal[idx][2]
				sum += int64(dr*dr + dg*dg + db*db)
			}
		}
	}
	s.file(id, "quant.raw", q)
	mse := float64(sum) / float64(w*h)
	p2 := fmt.Sprintf("MSE: %.2f\nWROTE quant.raw (0x%X байт)", mse, len(q))
	s.expected(id, strings.TrimRight(p1.String(), "\n"), p2)
}

func (s *set) gen040() {
	id := "040"
	mt := time.Unix(1000000000, 0).UTC()
	entries := []struct {
		name string
		mode int64
		size int
		typ  byte
	}{
		{"HEIST_PLAN.TXT", 0644, 4096, tar.TypeReg},
		{"TRACKING.DB", 0644, 128, tar.TypeReg},
		{"VAULT", 0755, 0, tar.TypeDir},
	}
	dataFor := func(name string, size int) []byte {
		b := make([]byte, size)
		for i := range b {
			if name == "TRACKING.DB" {
				b[i] = byte(i*13 + 5)
			} else {
				b[i] = byte(i*7 + 3)
			}
		}
		return b
	}
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range entries {
		hdr := &tar.Header{Name: e.name, Mode: e.mode, Size: int64(e.size), ModTime: mt, Typeflag: e.typ}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		if e.size > 0 {
			tw.Write(dataFor(e.name, e.size))
		}
	}
	tw.Close()
	tb := buf.Bytes()

	// corrupt TRACKING.DB checksum: header 2 starts at 512 + 8*512
	off := 512 + 8*512
	cf := tb[off+148 : off+156]
	for i := 0; i < 8; i++ {
		if cf[i] >= '0' && cf[i] <= '9' {
			cf[i]++
			break
		}
	}
	s.file(id, "archive.tar", tb)

	recs := parseTar(tb)
	var p1 strings.Builder
	for _, r := range recs {
		fmt.Fprintf(&p1, "name=%s size=%d mode=%04o type=%d mtime=%d\n",
			r.name, r.size, r.mode, r.typ, r.mtime)
	}
	var p2 strings.Builder
	for _, r := range recs {
		if r.typ == 5 {
			fmt.Fprintf(&p2, "%s: DIR (size=%d)\n", r.name, r.size)
			continue
		}
		if r.cksumOK {
			fmt.Fprintf(&p2, "%s: EXTRACTED (cksum OK)\n", r.name)
		} else {
			fmt.Fprintf(&p2, "%s: EXTRACTED (cksum BAD)\n", r.name)
		}
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

// ---------------------------------------------------------------- helpers

func lz77Enc(data []byte) []byte {
	var out []byte
	pos := 0
	for pos < len(data) {
		bestLen, bestOff := 0, 0
		maxOff := pos
		if maxOff > 4096 {
			maxOff = 4096
		}
		for off := 1; off <= maxOff; off++ {
			l := 0
			for l < 255 && pos+l < len(data) && data[pos-off+l] == data[pos+l] {
				l++
			}
			if l > bestLen {
				bestLen, bestOff = l, off
			}
		}
		if bestLen >= 3 {
			out = append(out, 0x00)
			out = binary.LittleEndian.AppendUint16(out, uint16(bestOff))
			out = append(out, byte(bestLen))
			pos += bestLen
		} else {
			out = append(out, 0x80, data[pos])
			pos++
		}
	}
	return out
}

func lz77Dec(p []byte) []byte {
	var out []byte
	i := 0
	for i < len(p) {
		flag := p[i]
		i++
		if flag&0x80 != 0 {
			out = append(out, p[i])
			i++
			continue
		}
		off := int(binary.LittleEndian.Uint16(p[i : i+2]))
		l := int(p[i+2])
		i += 3
		for k := 0; k < l; k++ {
			out = append(out, out[len(out)-off])
		}
	}
	return out
}

func packHeader(version, typ, flags, length uint64) []byte {
	var w sol.BitWriter
	w.Write(version, 3)
	w.Write(typ, 5)
	w.Write(flags, 8)
	w.Write(length, 16)
	return w.Bytes()
}

type tarRec struct {
	name    string
	mode    int
	size    int
	mtime   int
	typ     int
	cksumOK bool
}

func parseTar(b []byte) []tarRec {
	var recs []tarRec
	off := 0
	for off+512 <= len(b) {
		h := b[off : off+512]
		name := trimNUL(h[0:100])
		if name == "" {
			break
		}
		mode := oct(h[100:108])
		size := oct(h[124:136])
		mtime := oct(h[136:148])
		cksum := oct(h[148:156])
		typ := 0
		switch h[156] {
		case '5':
			typ = 5
		}
		var sum int
		for i, x := range h {
			if i >= 148 && i < 156 {
				x = 0x20
			}
			sum += int(x)
		}
		sum &= 0xFFFF
		recs = append(recs, tarRec{name, mode, size, mtime, typ, sum == cksum})
		off += 512 + (size+511)/512*512
	}
	return recs
}

func trimNUL(b []byte) string {
	i := 0
	for i < len(b) && b[i] != 0 {
		i++
	}
	return string(b[:i])
}

func oct(b []byte) int {
	var s []byte
	for _, c := range b {
		if c == 0 || c == ' ' {
			continue
		}
		s = append(s, c)
	}
	v := 0
	for _, c := range s {
		if c < '0' || c > '7' {
			break
		}
		v = v*8 + int(c-'0')
	}
	return v
}

func init() {
	register("031", (*set).gen031)
	register("032", (*set).gen032)
	register("033", (*set).gen033)
	register("034", (*set).gen034)
	register("035", (*set).gen035)
	register("036", (*set).gen036)
	register("037", (*set).gen037)
	register("038", (*set).gen038)
	register("039", (*set).gen039)
	register("040", (*set).gen040)
}
