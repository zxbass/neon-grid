package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"neon-grid/sol"
)

func be32(b []byte) uint32 { return binary.BigEndian.Uint32(b) }
func u16be(b []byte, v uint16) { binary.BigEndian.PutUint16(b, v) }

// ---------------------------------------------------------------- 141 ZIP

type zipFile struct {
	name string
	data []byte
}

func buildZip(files []zipFile) []byte {
	var out []byte
	var offs []int
	for _, f := range files {
		offs = append(offs, len(out))
		crc := sol.CRC32IEEE(f.data)
		name := []byte(f.name)
		hdr := make([]byte, 30)
		u32le(hdr[0:], 0x04034b50)
		u16le(hdr[4:], 20)
		u16le(hdr[8:], 0) // method STORE
		u32le(hdr[14:], crc)
		u32le(hdr[18:], uint32(len(f.data)))
		u32le(hdr[22:], uint32(len(f.data)))
		u16le(hdr[26:], uint16(len(name)))
		out = append(out, hdr...)
		out = append(out, name...)
		out = append(out, f.data...)
	}
	cdStart := len(out)
	var p1 []string
	for i, f := range files {
		name := []byte(f.name)
		crc := sol.CRC32IEEE(f.data)
		rec := make([]byte, 46)
		u32le(rec[0:], 0x02014b50)
		u16le(rec[4:], 20)
		u16le(rec[6:], 20)
		u32le(rec[16:], crc)
		u32le(rec[20:], uint32(len(f.data)))
		u32le(rec[24:], uint32(len(f.data)))
		u16le(rec[28:], uint16(len(name)))
		u32le(rec[42:], uint32(offs[i]))
		out = append(out, rec...)
		out = append(out, name...)
		p1 = append(p1, fmt.Sprintf("%s size=%d crc=0x%08X", f.name, len(f.data), crc))
	}
	cdSize := len(out) - cdStart
	eocd := make([]byte, 22)
	u32le(eocd[0:], 0x06054b50)
	u16le(eocd[8:], uint16(len(files)))
	u16le(eocd[10:], uint16(len(files)))
	u32le(eocd[12:], uint32(cdSize))
	u32le(eocd[16:], uint32(cdStart))
	out = append(out, eocd...)
	return out
}

func (s *set) gen141() {
	id := "141"
	zip := buildZip([]zipFile{
		{"README.TXT", []byte("NEON GRID ONLINE")},
		{"MAP.BIN", []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB}},
	})
	s.file(id, "archive.zip", zip)
	crc1 := sol.CRC32IEEE([]byte("NEON GRID ONLINE"))
	crc2 := sol.CRC32IEEE([]byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB})
	p1 := fmt.Sprintf("README.TXT size=16 crc=0x%08X\nMAP.BIN size=12 crc=0x%08X", crc1, crc2)
	s.expected(id, p1, "README.TXT: NEON GRID ONLINE")
}

// ---------------------------------------------------------------- 142 JPEG

func jpegSegment(out []byte, marker byte, payload []byte) []byte {
	out = append(out, 0xFF, marker)
	l := uint16(len(payload) + 2)
	out = append(out, byte(l>>8), byte(l))
	return append(out, payload...)
}

func (s *set) gen142() {
	id := "142"
	var jpg []byte
	jpg = append(jpg, 0xFF, 0xD8) // SOI
	app0 := []byte{'J', 'F', 'I', 'F', 0, 1, 2, 0, 0, 1, 0, 1, 0, 0}
	jpg = jpegSegment(jpg, 0xE0, app0)
	q := []byte{0x00, 16, 11, 10, 16, 24, 40, 51, 61, 12, 12, 14, 19, 26, 58, 60, 55,
		14, 13, 16, 24, 40, 57, 69, 56, 14, 17, 22, 29, 51, 87, 80, 62,
		18, 22, 37, 56, 68, 109, 103, 77, 24, 35, 55, 64, 81, 104, 113, 92,
		49, 64, 78, 87, 103, 121, 120, 101, 72, 92, 95, 98, 112, 100, 103, 99}
	jpg = jpegSegment(jpg, 0xDB, q)
	sof0 := []byte{8, 0, 16, 0, 16, 1, 1, 0x11, 0}
	jpg = jpegSegment(jpg, 0xC0, sof0)
	sos := []byte{1, 1, 0x00, 0x00, 0x3F, 0x00}
	jpg = jpegSegment(jpg, 0xDA, sos)
	jpg = append(jpg, 0x33, 0x77, 0xFF, 0xD9) // some scan data + EOI
	s.file(id, "photo.jpg", jpg)
	var sum uint32
	for _, v := range q[1:] {
		sum += uint32(v)
	}
	first8 := make([]string, 8)
	for i := 0; i < 8; i++ {
		first8[i] = fmt.Sprint(q[1+i])
	}
	p1 := "WIDTH=16 HEIGHT=16\nCOMPONENTS=1 PRECISION=8\nSEGMENTS: APP0 DQT SOF0 SOS"
	p2 := fmt.Sprintf("QTABLE[0]: %s\nQSUM: %d", strings.Join(first8, " "), sum)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 143 MIDI

func midiVLQ(v int) []byte {
	if v < 0x80 {
		return []byte{byte(v)}
	}
	var out []byte
	for v > 0 {
		out = append([]byte{byte(v & 0x7F)}, out...)
		v >>= 7
	}
	for i := 0; i < len(out)-1; i++ {
		out[i] |= 0x80
	}
	return out
}

func midiNoteName(n int) string {
	names := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	return fmt.Sprintf("%s%d", names[n%12], n/12-1)
}

func (s *set) gen143() {
	id := "143"
	notes := []struct {
		note, dur int
	}{
		{60, 480}, {62, 240}, {64, 960},
	}
	var ev []byte
	ev = append(ev, 0x00, 0xFF, 0x51, 0x03, 0x07, 0xA1, 0x20) // tempo 500000
	ev = append(ev, 0x00, 0xFF, 0x58, 0x04, 0x04, 0x02, 0x18, 0x08)
	for _, n := range notes {
		ev = append(ev, 0x00, 0x90, byte(n.note), 100)
		ev = append(ev, midiVLQ(n.dur)...)
		ev = append(ev, 0x80, byte(n.note), 0)
	}
	ev = append(ev, 0x00, 0xFF, 0x2F, 0x00)
	trk := append([]byte{'M', 'T', 'r', 'k'}, 0, 0, 0, 0)
	trk = append(trk, ev...)
	u32be(trk[4:], uint32(len(ev)))

	mid := []byte{'M', 'T', 'h', 'd', 0, 0, 0, 6, 0, 0, 0, 1, 0x01, 0xE0}
	mid = append(mid, trk...)
	s.file(id, "song.mid", mid)

	var p1 []string
	p1 = append(p1, fmt.Sprintf("TRACKS: 1 TEMPO: 500000 DIVISION: 480 NOTES: %d", len(notes)))
	var p2 []string
	for i, n := range notes {
		p2 = append(p2, fmt.Sprintf("N%d: %s DUR=%d", i, midiNoteName(n.note), n.dur))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 144 TTF

func ttChecksum(data []byte) uint32 {
	var sum uint32
	for i := 0; i+4 <= len(data); i += 4 {
		sum += be32(data[i:])
	}
	return sum
}

func (s *set) gen144() {
	id := "144"
	// cmap format 4 with segment [0x41..0x41] -> glyph 0 and 0xFFFF sentinel
	var cmap []byte
	cmap = append(cmap, 0, 0, 0, 1, 0, 3, 0, 1, 0, 0, 0, 12) // header
	seg := make([]byte, 32)
	u16be := binary.BigEndian.PutUint16
	u16be(seg[0:], 4)  // format
	u16be(seg[2:], 32) // length
	u16be(seg[6:], 4)  // segCountX2
	u16be(seg[8:], 4)  // searchRange
	u16be(seg[10:], 1) // entrySelector
	u16be(seg[12:], 0) // rangeShift
	u16be(seg[14:], 0x0041)
	u16be(seg[16:], 0xFFFF)
	u16be(seg[18:], 0) // reservedPad
	u16be(seg[20:], 0x0041)
	u16be(seg[22:], 0xFFFF)
	u16be(seg[24:], 0xFFBF) // idDelta: (0x41 + 0xFFBF) & 0xFFFF = 0
	u16be(seg[26:], 1)
	u16be(seg[28:], 0)
	u16be(seg[30:], 0)
	cmap = append(cmap, seg...)

	head := make([]byte, 54)
	u32be(head[0:], 0x00010000)
	u32be(head[4:], 0x00010000)
	u32be(head[12:], 0x5F0F3CF5)
	u16be(head[18:], 2048) // unitsPerEm
	u16be(head[44:], 8)    // lowestRecPPEM
	u16be(head[46:], 2)    // fontDirectionHint
	u16be(head[50:], 0)    // indexToLocFormat

	hhea := make([]byte, 36)
	u32be(hhea[0:], 0x00010000)
	u16be(hhea[32:], 1) // numberOfHMetrics

	maxp := make([]byte, 6)
	u32be(maxp[0:], 0x00010000)
	u16be(maxp[4:], 1) // numGlyphs

	hmtx := make([]byte, 4)
	u16be(hmtx[0:], 1000)

	glyf := make([]byte, 10) // one empty glyph

	loca := make([]byte, 4)
	u16be(loca[0:], 0)
	u16be(loca[2:], 10)

	tables := []struct {
		tag  string
		data []byte
	}{
		{"cmap", cmap}, {"glyf", glyf}, {"head", head}, {"hhea", hhea},
		{"hmtx", hmtx}, {"loca", loca}, {"maxp", maxp},
	}

	order := []string{"cmap", "glyf", "head", "hhea", "hmtx", "loca", "maxp"}
	var out []byte
	// offset table placeholder: 12 + 16*7 = 124 bytes
	offsetTable := make([]byte, 12+16*len(tables))
	u32be(offsetTable[0:], 0x00010000)
	u16be(offsetTable[4:], uint16(len(tables)))
	u16be(offsetTable[6:], 64) // searchRange
	u16be(offsetTable[8:], 2)  // entrySelector
	u16be(offsetTable[10:], 48)
	out = append(out, offsetTable...)
	var tmap = map[string][]byte{}
	for _, t := range tables {
		tmap[t.tag] = t.data
	}
	off := len(out)
	for _, tag := range order {
		data := tmap[tag]
		rec := out[12+16*indexOfString(order, tag):]
		copy(rec[0:4], tag)
		u32be(rec[4:], ttChecksum(data))
		u32be(rec[8:], uint32(off))
		u32be(rec[12:], uint32(len(data)))
		out = append(out, data...)
		off += len(data)
	}
	// head checksumAdjustment
	whole := ttChecksum(out)
	u32be(out[8:], 0xB1B0AFBA-whole)
	s.file(id, "font.ttf", out)

	names := strings.Join(order, " ")
	s.expected(id, "TABLES: "+names, "UNITS: 2048 GLYPHS: 1\nA -> GLYPH 0")
}

func indexOfString(a []string, v string) int {
	for i, x := range a {
		if x == v {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------- 145 WASM

func uleb128(n uint32) []byte {
	var out []byte
	for {
		b := byte(n & 0x7F)
		n >>= 7
		if n == 0 {
			out = append(out, b)
			return out
		}
		out = append(out, b|0x80)
	}
}

func (s *set) gen145() {
	id := "145"
	var m []byte
	m = append(m, 0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00) // \0asm v1

	// type section: 1 func type (i32)->(i32)
	payload := []byte{0x01, 0x60, 0x01, 0x7F, 0x01, 0x7F}
	m = append(m, 0x01)
	m = append(m, uleb128(uint32(len(payload)))...)
	m = append(m, payload...)

	// function section: 1 func, type 0
	m = append(m, 0x03)
	m = append(m, uleb128(2)...)
	m = append(m, 0x01, 0x00)

	// export section: "zen" func 0
	var ex []byte
	ex = append(ex, 0x01)
	ex = append(ex, 0x04)
	ex = append(ex, 'z', 'e', 'n')
	ex = append(ex, 0x00, 0x00)
	m = append(m, 0x07)
	m = append(m, uleb128(uint32(len(ex)))...)
	m = append(m, ex...)

	// code section: body = i32.const 42; end
	body := []byte{0x00, 0x41, 0x2A, 0x0B}
	var cs []byte
	cs = append(cs, 0x01)
	cs = append(cs, uleb128(uint32(len(body)))...)
	cs = append(cs, body...)
	m = append(m, 0x0A)
	m = append(m, uleb128(uint32(len(cs)))...)
	m = append(m, cs...)
	s.file(id, "mod.wasm", m)

	s.expected(id, "SECTIONS: TYPE FUNC EXPORT CODE", "EXPORT: zen FUNC 0\nBODY BYTES: 3")
}

// ---------------------------------------------------------------- 146 PDF

func (s *set) gen146() {
	id := "146"
	content := "CROW WAS HERE 2049"
	var buf bytes.Buffer
	write := func(b []byte) int { n := len(b); buf.Write(b); return n }
	buf.WriteString("%PDF-1.4\n")
	var offs []int
	offs = append(offs, buf.Len())
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	offs = append(offs, buf.Len())
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	offs = append(offs, buf.Len())
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /Contents 4 0 R >>\nendobj\n")
	offs = append(offs, buf.Len())
	fmt.Fprintf(&buf, "4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(content), content)
	xrefOff := buf.Len()
	buf.WriteString("xref\n0 5\n")
	write([]byte("0000000000 65535 f \n"))
	for _, o := range offs {
		fmt.Fprintf(&buf, "%010d 00000 n \n", o)
	}
	buf.WriteString("trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n")
	fmt.Fprintf(&buf, "%d\n%%%%EOF\n", xrefOff)
	write(nil)
	s.file(id, "doc.pdf", buf.Bytes())

	p1 := "CATALOG: obj1\nPAGES: obj2 COUNT=1\nPAGE: obj3\nSTREAM: obj4"
	s.expected(id, p1, fmt.Sprintf("CONTENT: %s", content))
}

// ---------------------------------------------------------------- 147 MP3

func (s *set) gen147() {
	id := "147"
	header := []byte{0xFF, 0xFB, 0x90, 0x00} // MPEG1 L3, 128kbps, 44100, stereo
	frameSize := 144*128000/44100 + 0        // 417
	const frames = 10
	var mp3 []byte
	for i := 0; i < frames; i++ {
		mp3 = append(mp3, header...)
		mp3 = append(mp3, make([]byte, frameSize-4)...)
	}
	mp3 = append(mp3, 0, 0, 0, 0)
	s.file(id, "clip.mp3", mp3)

	rows := []string{"FRAME 0: 128kbps 44100Hz STEREO", "FRAME 1: 128kbps 44100Hz STEREO", "FRAME 2: 128kbps 44100Hz STEREO"}
	dur := float64(frames*1152) / 44100.0
	p2 := fmt.Sprintf("FRAMES: %d DURATION: %.3fs", frames, dur)
	s.expected(id, strings.Join(rows, "\n"), p2)
}

// ---------------------------------------------------------------- 148 UTF-8

type utf8Case struct {
	name string
	data []byte
}

func utf8Valid(b []byte) (bool, string) {
	bad := ""
	i := 0
	for i < len(b) {
		c := b[i]
		switch {
		case c < 0x80:
			i++
		case c >= 0xC2 && c <= 0xDF:
			if i+1 >= len(b) || b[i+1]&0xC0 != 0x80 {
				bad = "TRUNCATED"
				return false, bad
			}
			i += 2
		case c >= 0xE0 && c <= 0xEF:
			if i+2 >= len(b) || b[i+1]&0xC0 != 0x80 || b[i+2]&0xC0 != 0x80 {
				bad = "TRUNCATED"
				return false, bad
			}
			if c == 0xE0 && b[i+1] < 0xA0 {
				bad = "OVERLONG"
				return false, bad
			}
			if c == 0xED && b[i+1] >= 0xA0 {
				bad = "SURROGATE"
				return false, bad
			}
			i += 3
		case c >= 0xF0 && c <= 0xF4:
			if i+3 >= len(b) || b[i+1]&0xC0 != 0x80 || b[i+2]&0xC0 != 0x80 || b[i+3]&0xC0 != 0x80 {
				bad = "TRUNCATED"
				return false, bad
			}
			if c == 0xF0 && b[i+1] < 0x90 {
				bad = "OVERLONG"
				return false, bad
			}
			if c == 0xF4 && b[i+1] >= 0x90 {
				bad = "OUT_OF_RANGE"
				return false, bad
			}
			i += 4
		default:
			bad = "INVALID"
			return false, bad
		}
	}
	return true, ""
}

func utf8Count(b []byte) int {
	cp := 0
	i := 0
	for i < len(b) {
		switch {
		case b[i] < 0x80:
			cp++
			i++
		case b[i] < 0xE0:
			cp++
			i += 2
		case b[i] < 0xF0:
			cp++
			i += 3
		default:
			cp++
			i += 4
		}
	}
	return cp
}

func (s *set) gen148() {
	id := "148"
	cases := []utf8Case{
		{"ZEN", []byte("ZEN")},
		{"GRID_CYRILLIC", []byte("ГРИД")},
		{"E2_82_AC", []byte{0xE2, 0x82, 0xAC}},
		{"OVERLONG_C0_AF", []byte{0xC0, 0xAF}},
		{"SURROGATE_ED_A0_80", []byte{0xED, 0xA0, 0x80}},
		{"F4_90_80_80", []byte{0xF4, 0x90, 0x80, 0x80}},
		{"F8_80_80_80_80", []byte{0xF8, 0x80, 0x80, 0x80, 0x80}},
		{"LONE_80", []byte{0x80}},
		{"TRUNCATED_E2_82", []byte{0xE2, 0x82}},
	}
	var blob []byte
	for i, c := range cases {
		if i > 0 {
			blob = append(blob, 0xFF) // separator
		}
		blob = append(blob, c.data...)
	}
	s.file(id, "utf8.bin", blob)

	var p1 []string
	total := 0
	for _, c := range cases {
		ok, reason := utf8Valid(c.data)
		if ok {
			cp := utf8Count(c.data)
			total += cp
			p1 = append(p1, fmt.Sprintf("%s: OK %d", c.name, cp))
		} else {
			p1 = append(p1, fmt.Sprintf("%s: BAD %s", c.name, reason))
		}
	}
	p2 := fmt.Sprintf("TOTAL OK: %d", total)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 149 ISO9660

func (s *set) gen149() {
	id := "149"
	const sector = 2048
	pvd := make([]byte, sector)
	pvd[0] = 1
	copy(pvd[1:6], "CD001")
	pvd[6] = 1
	copy(pvd[8:40], "NEON_GRID")
	copy(pvd[40:72], "NEON_ISO_2049")
	u32le(pvd[80:], 400)
	u32be(pvd[84:], 400)
	u32le(pvd[120:], 1)
	u32be(pvd[124:], 1)
	u32le(pvd[128:], 1)
	u32be(pvd[132:], 1)
	u32le(pvd[136:], sector)
	u32be(pvd[140:], sector)
	// root directory record
	root := make([]byte, 34)
	root[0] = 34
	u32le(root[2:], 20)
	u32be(root[6:], 20)
	u32le(root[10:], sector)
	u32be(root[14:], sector)
	root[18] = 0
	root[19] = 0
	root[20] = 0
	root[21] = 2 // directory flag
	u16le(root[28:], 1)
	u16be(root[30:], 1)
	root[32] = 1
	root[33] = '.'
	copy(pvd[156:], root)

	// root dir sector 20: ".", "..", HELLO.TXT
	content := "NEON GRID WELCOMES YOU"
	rootDir := make([]byte, sector)
	dot := make([]byte, 34)
	dot[0] = 34
	u32le(dot[2:], 20)
	u32be(dot[6:], 20)
	u32le(dot[10:], sector)
	u32be(dot[14:], sector)
	dot[21] = 2
	u16le(dot[28:], 1)
	u16be(dot[30:], 1)
	dot[32] = 1
	dot[33] = '.'
	copy(rootDir[0:], dot)
	dotdot := make([]byte, 34)
	copy(dotdot, dot)
	dotdot[33] = '.'
	copy(rootDir[34:], dotdot)
	file := make([]byte, 42)
	file[0] = 42
	u32le(file[2:], 21)
	u32be(file[6:], 21)
	u32le(file[10:], uint32(len(content)))
	u32be(file[14:], uint32(len(content)))
	file[18] = 0
	u16le(file[28:], 1)
	u16be(file[30:], 1)
	file[32] = 9
	copy(file[33:42], "HELLO.TXT")
	copy(rootDir[68:], file)

	fileSec := make([]byte, sector)
	copy(fileSec, content)

	iso := append([]byte{}, pvd...)
	iso = append(iso, make([]byte, sector*19)...) // sectors 1..19 empty
	iso = append(iso, rootDir...)
	iso = append(iso, fileSec...)
	s.file(id, "disk.iso", iso)

	p1 := fmt.Sprintf("VOLUME: NEON_ISO_2049\nROOT EXTENT: 20 SIZE: %d\nFILES: 1", sector)
	p2 := fmt.Sprintf("HELLO.TXT SIZE=%d\nCONTENT: %s", len(content), content)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 150 OLE/CFB

func (s *set) gen150() {
	id := "150"
	const sector = 512
	// header
	hdr := make([]byte, sector)
	copy(hdr[0:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
	u16le(hdr[24:], 0x003E) // minor
	u16le(hdr[26:], 0x0003) // major
	u16le(hdr[28:], 0xFFFE) // byte order
	u16le(hdr[30:], 9)      // sector shift
	u16le(hdr[32:], 6)      // mini shift
	u32le(hdr[44:], 0)      // num dir sectors
	u32le(hdr[48:], 1)      // num FAT sectors
	u32le(hdr[52:], 2)      // first dir sector
	u32le(hdr[56:], 0)      // transaction
	u32le(hdr[60:], 4096)   // mini cutoff
	u32le(hdr[64:], 0xFFFFFFFE)
	u32le(hdr[68:], 0)
	u32le(hdr[72:], 0xFFFFFFFE)
	u32le(hdr[76:], 0) // first FAT sector in DIFAT
	for i := 1; i < 109; i++ {
		u32le(hdr[76+i*4:], 0xFFFFFFFF)
	}

	// sector 1: FAT
	fat := make([]byte, sector)
	// sector 2 = dir, chain end; sector 3 = README, end; 4->5->end (SECRET.BIN)
	u32le(fat[1*4:], 0xFFFFFFFE)
	u32le(fat[3*4:], 0xFFFFFFFE)
	u32le(fat[4*4:], 5)
	u32le(fat[5*4:], 0xFFFFFFFE)
	for i := 6; i < 128; i++ {
		u32le(fat[i*4:], 0xFFFFFFFF)
	}

	// sector 2: directory, 4 entries x 128 bytes
	dirS := make([]byte, sector)
	rootEntry := make([]byte, 128)
	name := "Root Entry"
	u16le(rootEntry[0:], uint16(len(name)*2+2))
	copy(rootEntry[2:], utf16le(name))
	rootEntry[64] = 5 // root storage
	rootEntry[66] = 1
	u32le(rootEntry[68:], 0xFFFFFFFF)
	u32le(rootEntry[72:], 0xFFFFFFFF)
	u32le(rootEntry[76:], 1) // child = entry 1
	copy(dirS[0:], rootEntry)

	readme := make([]byte, 128)
	name = "README"
	u16le(readme[0:], uint16(len(name)*2+2))
	copy(readme[2:], utf16le(name))
	readme[64] = 2 // stream
	readme[66] = 0
	u32le(readme[68:], 0xFFFFFFFF)
	u32le(readme[72:], 0xFFFFFFFF)
	u32le(readme[76:], 0xFFFFFFFF)
	u32le(readme[116:], 3) // start sector
	u32le(readme[120:], 4096)
	copy(dirS[128:], readme)

	secret := make([]byte, 128)
	name = "SECRET.BIN"
	u16le(secret[0:], uint16(len(name)*2+2))
	copy(secret[2:], utf16le(name))
	secret[64] = 2
	u32le(secret[68:], 0xFFFFFFFF)
	u32le(secret[72:], 0xFFFFFFFF)
	u32le(secret[76:], 0xFFFFFFFF)
	u32le(secret[116:], 4)
	u32le(secret[120:], 8192)
	copy(dirS[256:], secret)

	// sector 3: README data
	readmeData := make([]byte, 4096)
	copy(readmeData, "OLE_STREAM_READY:NEON_GRID:2049\n")
	// sector 4-5: SECRET.BIN
	secretData := make([]byte, 8192)
	copy(secretData, "SECRET:THE_BLACK_GATE_IS_OPEN\n")

	ole := append([]byte{}, hdr...)
	ole = append(ole, fat...)
	ole = append(ole, dirS...)
	ole = append(ole, readmeData...)
	ole = append(ole, secretData...)
	s.file(id, "doc.ole", ole)

	p1 := "MAGIC: OK SECTOR_SIZE: 512\nENTRIES: Root Entry, README, SECRET.BIN"
	p2 := "README SIZE: 4096\nCONTENT: OLE_STREAM_READY:NEON_GRID:2049"
	s.expected(id, p1, p2)
}

func utf16le(s string) []byte {
	var out []byte
	for _, r := range s {
		out = append(out, byte(r), byte(r>>8))
	}
	return out
}

func init() {
	register("141", (*set).gen141)
	register("142", (*set).gen142)
	register("143", (*set).gen143)
	register("144", (*set).gen144)
	register("145", (*set).gen145)
	register("146", (*set).gen146)
	register("147", (*set).gen147)
	register("148", (*set).gen148)
	register("149", (*set).gen149)
	register("150", (*set).gen150)
}
