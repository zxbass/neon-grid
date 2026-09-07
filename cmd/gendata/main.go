// Command gendata generates the input data files and expected outputs
// (expected.txt) for the LeetCode-style mission skeletons in solutions/.
//
// The reference logic below mirrors what a correct solution must produce;
// shared algorithms come from the tested sol/ package.
//
// Usage: go run ./cmd/gendata
package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"neon-grid/sol"
)

const dataRoot = "data"

// genRegistry maps a mission id ("021") to its generator. Generators
// self-register in init() of their group file.
var genRegistry = map[string]func(*set){}

func register(id string, fn func(*set)) {
	if _, dup := genRegistry[id]; dup {
		panic("duplicate generator for " + id)
	}
	genRegistry[id] = fn
}

func main() {
	ids := flag.String("ids", "", "comma-separated mission ids to generate (empty = all)")
	flag.Parse()
	want := func(id string) bool {
		if *ids == "" {
			return true
		}
		for _, part := range strings.Split(*ids, ",") {
			if lo, hi, ok := strings.Cut(part, "-"); ok {
				if id >= lo && id <= hi {
					return true
				}
				continue
			}
			if id == part {
				return true
			}
		}
		return false
	}
	mk := newSet()
	for id, fn := range genRegistry {
		if want(id) {
			fn(mk)
		}
	}
	fmt.Println("generated", len(mk.written), "files under", dataRoot)
}

func init() {
	register("001", (*set).gen001)
	register("002", (*set).gen002)
	register("003", (*set).gen003)
	register("004", (*set).gen004)
	register("005", (*set).gen005)
	register("006", (*set).gen006)
	register("007", (*set).gen007)
	register("008", (*set).gen008)
	register("009", (*set).gen009)
	register("010", (*set).gen010)
	register("011", (*set).gen011)
	register("012", (*set).gen012)
	register("013", (*set).gen013)
	register("014", (*set).gen014)
	register("015", (*set).gen015)
	register("016", (*set).gen016)
	register("017", (*set).gen017)
	register("018", (*set).gen018)
	register("019", (*set).gen019)
	register("020", (*set).gen020)
}

// ---------------------------------------------------------------- helpers

type set struct {
	written map[string]bool
}

func newSet() *set { return &set{written: map[string]bool{}} }

func (s *set) dir(id string) string {
	d := filepath.Join(dataRoot, id)
	if err := os.MkdirAll(d, 0o755); err != nil {
		panic(err)
	}
	return d
}

func (s *set) file(id, name string, data []byte) string {
	dir := s.dir(id)
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		panic(err)
	}
	s.written[p] = true
	return p
}

func (s *set) text(id, name, content string) string {
	return s.file(id, name, []byte(content))
}

func (s *set) expected(id, part1, part2 string) {
	body := strings.TrimRight(part1, "\n")
	if strings.TrimSpace(part2) != "" {
		body += "\n=== PART 2 ===\n" + strings.TrimRight(part2, "\n")
	}
	s.text(id, "expected.txt", body+"\n")
}

func u16le(b []byte, v uint16) { binary.LittleEndian.PutUint16(b, v) }
func u32le(b []byte, v uint32) { binary.LittleEndian.PutUint32(b, v) }
func u32be(b []byte, v uint32) { binary.BigEndian.PutUint32(b, v) }

func hexStr(b []byte) string {
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02X", x)
	}
	return strings.Join(parts, " ")
}

// classicDump renders mission-003 style: 16 bytes/line, 8+8 split, ASCII.
func classicDump(data []byte) string {
	var out strings.Builder
	for off := 0; off < len(data); off += 16 {
		end := off + 16
		if end > len(data) {
			end = len(data)
		}
		chunk := data[off:end]
		var hexp strings.Builder
		ascii := make([]byte, 16)
		for i := 0; i < 16; i++ {
			if i == 8 {
				hexp.WriteString(" ")
			}
			if i < len(chunk) {
				fmt.Fprintf(&hexp, "%02x ", chunk[i])
				if chunk[i] >= 0x20 && chunk[i] < 0x7f {
					ascii[i] = chunk[i]
				} else {
					ascii[i] = '.'
				}
			} else {
				hexp.WriteString("   ")
				ascii[i] = ' '
			}
		}
		fmt.Fprintf(&out, "%08x  %s |%s|\n", off, strings.TrimRight(hexp.String(), " "), ascii)
	}
	return out.String()
}

// ---------------------------------------------------------------- missions

func (s *set) gen001() {
	id := "001"
	boot := make([]byte, 512)
	copy(boot, []byte{0xEB, 0x3C, 0x90, 0x4D, 0x53, 0x44, 0x4F, 0x53, 0x35, 0x2E, 0x30, 0x00, 0x02, 0x01, 0x01, 0x00})
	boot[510], boot[511] = 0x55, 0xAA
	s.file(id, "boot.bin", boot)
	p1 := hexStr(boot[:16])
	var sum int
	for _, b := range boot {
		sum = (sum + int(b)) & 0xFF
	}
	p2 := fmt.Sprintf("BOOT OK\nSUM: 0x%02X", sum)
	s.expected(id, p1, p2)
}

func (s *set) gen002() {
	id := "002"
	s.text(id, "input.txt", "AA C1 3F 80 42 0E\n63 1 1\n42 0 1\n127 1 1\n")
	parse := func(b byte) string {
		size := b & 0x3F
		fast := "no"
		if b&0x40 != 0 {
			fast = "yes"
		}
		imp := "no"
		if b&0x80 != 0 {
			imp = "yes"
		}
		return fmt.Sprintf("SIZE=%d FAST=%s IMPORT=%s", size, fast, imp)
	}
	p1 := strings.Join([]string{
		parse(0xAA), parse(0xC1), parse(0x3F), parse(0x80), parse(0x42), parse(0x0E),
	}, "\n")
	p2 := "0xFF\n0xAA\nSIZE_OVERFLOW"
	s.expected(id, p1, p2)
}

func (s *set) gen003() {
	id := "003"
	disk := make([]byte, 0x1010)
	copy(disk[0:16], "NEONGRID1.0\x00\x00\x00\x00\x00")
	copy(disk[0x810:0x816], "*GRID*")
	copy(disk[0x816:0x81B], "ALLOY")
	copy(disk[0x1000:0x100D], "NEON GRID 2.0")
	s.file(id, "disk.img", disk)
	full := classicDump(disk)
	var filt strings.Builder
	prev := -1
	lines := strings.Split(strings.TrimRight(full, "\n"), "\n")
	for i, l := range lines {
		if !strings.Contains(l, "GRID") {
			continue
		}
		if prev >= 0 && i > prev+1 {
			filt.WriteString("....\n")
		}
		filt.WriteString(l + "\n")
		prev = i
	}
	s.expected(id, full, strings.TrimRight(filt.String(), "\n"))
}

func (s *set) gen004() {
	id := "004"
	var nums []byte
	for _, v := range []uint32{65, 1, 0, 0x11223344, 0x47454d4f, 0x41445243} {
		var b [4]byte
		u32le(b[:], v)
		nums = append(nums, b[:]...)
	}
	s.file(id, "nums.bin", nums)
	var p1 strings.Builder
	var p2 strings.Builder
	for i := 0; i+4 <= len(nums); i += 4 {
		w := nums[i : i+4]
		v := binary.LittleEndian.Uint32(w)
		be := []byte{w[3], w[2], w[1], w[0]}
		fmt.Fprintf(&p1, "value=%-5d BE=0x%08X\n", v, binary.BigEndian.Uint32(be))
		printable := true
		for _, b := range w {
			if b < 0x20 || b > 0x7E {
				printable = false
			}
		}
		if printable {
			fmt.Fprintf(&p2, "%s\n", string(be))
		}
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen005() {
	id := "005"
	const codeLen = 512
	code := make([]byte, codeLen)
	for i := range code {
		code[i] = byte(i * 7)
	}
	var crc uint64
	for i := 0; i+4 <= len(code); i += 4 {
		crc += uint64(binary.LittleEndian.Uint32(code[i : i+4]))
	}
	crc %= 0xFFFFFFFF

	fw := make([]byte, 64+codeLen)
	copy(fw[0:2], "FW")
	fw[2], fw[3] = 0x01, 0x04
	copy(fw[4:12], "omega-b1")
	u32le(fw[12:], codeLen)
	u32le(fw[16:], uint32(crc))
	u16le(fw[20:], 0x0302)
	u16le(fw[22:], 0x0701)
	copy(fw[24:64], "Primary bootloader rev 2")
	copy(fw[64:], code)
	s.file(id, "firmware.bin", fw)

	p1 := fmt.Sprintf("magic    : FW v1.4\nname     : omega-b1\ncode_len : %d\ncrc      : 0x%08X\napi      : 3.2\nhw       : 7.1\ndesc     : Primary bootloader rev 2", codeLen, crc)
	p2 := fmt.Sprintf("CRC OK (0x%08X)", crc)
	s.expected(id, p1, p2)
}

func (s *set) gen006() {
	id := "006"
	s.text(id, "input.txt", "00000013 00000202 00000040\n")
	names := []string{"READ", "WRITE", "TRACE", "BREAK", "INJECT", "TUNNEL", "SCAN"}
	decode := func(w uint32) string {
		var namesOn []string
		for b := 0; b < 32; b++ {
			if w&(1<<b) == 0 {
				continue
			}
			if b < len(names) {
				namesOn = append(namesOn, names[b])
			} else {
				namesOn = append(namesOn, fmt.Sprintf("bit%d(reserved)", b))
			}
		}
		return strings.Join(namesOn, " ")
	}
	p1 := fmt.Sprintf("00000013: %s\n00000202: %s\n00000040: %s",
		decode(0x13), decode(0x202), decode(0x40))
	p2 := "KEY: 0x25"
	s.expected(id, p1, p2)
}

type rec struct {
	id       uint32
	nick     string
	lvl      uint8
	karma    uint8
	credits  uint32
	ip       string
}

func (s *set) gen007() {
	id := "007"
	records := []rec{
		{10, "CROW", 7, 99, 123456, "10.0.0.7"},
		{12, "shadow", 9, 5, 5000, "8.8.4.4"},
		{17, "gr1m", 6, 3, 40000, "1.2.3.4"},
		{7, "zen", 2, 50, 100, "0.0.0.1"},
		{100, "reave", 8, 77, 99999, "5.6.7.8"},
	}
	var dump []byte
	for _, r := range records {
		var b [26]byte
		u32le(b[0:], r.id)
		copy(b[4:12], r.nick)
		b[12], b[13] = r.lvl, r.karma
		u32le(b[14:], r.credits)
		copy(b[18:26], r.ip)
		dump = append(dump, b[:]...)
	}
	var end [4]byte
	u32le(end[:], 0xFFFFFFFF)
	dump = append(dump, end[:]...)
	s.file(id, "memdump.bin", dump)

	var p1 strings.Builder
	avg := 0
	for _, r := range records {
		fmt.Fprintf(&p1, "ID=%08X nick=%-8s lvl=%d karma=%d cred=%d ip=%s\n",
			r.id, r.nick, r.lvl, r.karma, r.credits, r.ip)
		avg += int(r.credits)
	}
	avg /= len(records)
	var suspects []string
	for _, r := range records {
		if r.lvl >= 5 && r.karma < 20 {
			suspects = append(suspects, fmt.Sprintf("SUSPECT: %08X %s", r.id, r.nick))
		}
	}
	p2 := strings.Join(suspects, "\n") + fmt.Sprintf("\nAVG=%d", avg)
	s.expected(id, strings.TrimRight(p1.String(), "\n"), p2)
}

func dosTS(year, month, day, hour, min, sec int) uint32 {
	return uint32(year-1980)<<25 | uint32(month)<<21 | uint32(day)<<16 |
		uint32(hour)<<11 | uint32(min)<<5 | uint32(sec/2)
}

func dosTime(ts uint32) time.Time {
	sec2 := int(ts & 0x1F)
	min := int((ts >> 5) & 0x3F)
	hour := int((ts >> 11) & 0x1F)
	day := int((ts >> 16) & 0x1F)
	month := int((ts >> 21) & 0xF)
	year := 1980 + int((ts>>25)&0x7F)
	return time.Date(year, time.Month(month), day, hour, min, sec2*2, 0, time.UTC)
}

func (s *set) gen008() {
	id := "008"
	t1 := dosTS(1980, 1, 1, 0, 0, 0)   // 0x00210000
	t2 := dosTS(1980, 1, 2, 0, 2, 0)   // 0x00220040
	t3 := dosTS(2099, 12, 31, 23, 59, 58)
	t4 := dosTS(2049, 12, 31, 23, 59, 58)
	s.text(id, "input.txt",
		fmt.Sprintf("%08X %08X %08X\n", t1, t2, t3)+
			fmt.Sprintf("%08X %08X %08X\n", t1, t2, t4))
	f := func(ts uint32) string {
		return dosTime(ts).Format("2006-01-02 15:04:05")
	}
	p1 := f(t1) + "\n" + f(t2) + "\n" + f(t3)
	latest := dosTime(t4)
	earliest := dosTime(t1)
	diff := latest.Unix() - earliest.Unix()
	p2 := fmt.Sprintf("LATEST: %s\nDIFF: %d", f(t4), diff)
	s.expected(id, p1, p2)
}

func (s *set) gen009() {
	id := "009"
	cargo := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	seal := func(data []byte) uint32 {
		acc := uint32(0x5A5A5A5A)
		for _, b := range data {
			acc = ((acc << 5) | (acc >> 27)) ^ uint32(b)
		}
		return acc
	}
	got := seal(cargo)
	s.file(id, "cargo.bin", cargo)

	batch := make([]byte, 4+len(cargo))
	u32le(batch[0:], 0x00000000)
	copy(batch[4:], cargo)
	s.file(id, "batch.bin", batch)

	p1 := fmt.Sprintf("SEAL: 0x%08X", got)
	p2 := fmt.Sprintf("SEAL BAD (expected 0x%08X, got 0x%08X)", 0x00000000, got)
	s.expected(id, p1, p2)
}

func (s *set) gen010() {
	id := "010"
	line0 := append([]byte("NEONGRID1.0"), 0, 0, 0, 0, 0)
	line1 := append([]byte("ALLOY-66"), 0, 0, 0, 0, 0, 0, 0, 0)
	line5 := append([]byte("NEONGRID2.0"), 0, 0, 0, 0, 0)
	var data []byte
	data = append(data, line0...)
	data = append(data, line1...)
	data = append(data, line1...)
	data = append(data, line1...)
	data = append(data, line1...)
	data = append(data, line5...)
	s.file(id, "dump.bin", data)

	full := strings.Join(sol.HexDump(data), "\n")

	lines := sol.HexDump(data)
	var dedup []string
	starred := false
	for i, l := range lines {
		if i > 0 && l[9:] == lines[i-1][9:] {
			if !starred {
				dedup = append(dedup, "*")
				starred = true
			}
			continue
		}
		starred = false
		dedup = append(dedup, l)
	}
	s.expected(id, full, strings.Join(dedup, "\n"))
}

func (s *set) gen011() {
	id := "011"
	w, h := 16, 8
	bpp := uint16(24)
	rowBytes := (w*int(bpp)/8 + 3) &^ 3
	imgSize := rowBytes * h
	bmp := make([]byte, 54+imgSize)
	copy(bmp[0:2], "BM")
	u32le(bmp[2:], uint32(len(bmp)))
	u32le(bmp[10:], 54)
	u32le(bmp[14:], 40)
	u32le(bmp[18:], uint32(w))
	u32le(bmp[22:], uint32(h))
	u16le(bmp[26:], 1)
	u16le(bmp[28:], bpp)
	u32le(bmp[34:], uint32(imgSize))
	// bottom-up rows; pixel (x, screenY)
	grays := " .:-=+*#%@"
	var art strings.Builder
	for screenY := 0; screenY < h; screenY++ {
		row := bmp[54+screenY*rowBytes:]
		for x := 0; x < w; x++ {
			R := uint32(40 + 10*x)
			G := uint32(30 + 12*screenY)
			B := uint32(20 + 8*(x+screenY))
			row[x*3+0] = byte(B)
			row[x*3+1] = byte(G)
			row[x*3+2] = byte(R)
			bright := (R*30 + G*59 + B*11) / 100
			idx := int(bright) * 10 / 100
			if idx > 9 {
				idx = 9
			}
			art.WriteByte(grays[idx])
		}
		art.WriteByte('\n')
	}
	s.file(id, "image.bmp", bmp)
	p1 := fmt.Sprintf("WIDTH=%d HEIGHT=%d BPP=%d DATA=%d", w, h, bpp, imgSize)
	s.expected(id, p1, strings.TrimRight(art.String(), "\n"))
}

func pngChunk(typ string, data []byte) []byte {
	var c []byte
	var l [4]byte
	u32be(l[:], uint32(len(data)))
	c = append(c, l[:]...)
	c = append(c, typ...)
	c = append(c, data...)
	var cr [4]byte
	u32be(cr[:], sol.CRC32IEEE(append([]byte(typ), data...)))
	c = append(c, cr[:]...)
	return c
}

func makePNG(w, h uint32) []byte {
	var png []byte
	png = append(png, 0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A)
	ihdr := make([]byte, 13)
	u32be(ihdr[0:], w)
	u32be(ihdr[4:], h)
	ihdr[8] = 8
	ihdr[9] = 2
	png = append(png, pngChunk("IHDR", ihdr)...)
	raw := make([]byte, 0, int(w)*int(h)*3+int(h))
	for y := uint32(0); y < h; y++ {
		raw = append(raw, 0)
		for x := uint32(0); x < w; x++ {
			raw = append(raw, byte(10*x+50), byte(20*y+30), byte(5*(x+y)+10))
		}
	}
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	zw.Write(raw)
	zw.Close()
	png = append(png, pngChunk("IDAT", zbuf.Bytes())...)
	png = append(png, pngChunk("IEND", nil)...)
	return png
}

func (s *set) gen012() {
	id := "012"
	good := makePNG(2, 2)
	s.file(id, "good.png", good)
	bad := append([]byte{}, good...)
	bad[8+8+13+3] ^= 0xFF // corrupt last byte of IHDR data before its CRC
	s.file(id, "bad.png", bad)

	// part 1: chunk list of good.png
	var p1 strings.Builder
	for off := 8; off < len(good); {
		l := int(binary.BigEndian.Uint32(good[off:]))
		typ := string(good[off+4 : off+8])
		data := good[off+8 : off+8+l]
		crc := good[off+8+l : off+12+l]
		want := sol.CRC32IEEE(append([]byte(typ), data...))
		status := "OK"
		if want != binary.BigEndian.Uint32(crc) {
			status = "CORRUPT"
		}
		fmt.Fprintf(&p1, "%s len=%d crc=%s\n", typ, l, status)
		off += 12 + l
	}
	// part 2: repair bad.png IHDR
	ihdr := bad[8+8 : 8+8+13]
	w := binary.BigEndian.Uint32(ihdr[0:])
	h := binary.BigEndian.Uint32(ihdr[4:])
	dep, col, comp, filt, ilace := ihdr[8], ihdr[9], ihdr[10], ihdr[11], ihdr[12]
	fixed := sol.CRC32IEEE(append([]byte("IHDR"), ihdr...))
	_ = comp
	_ = filt
	p2 := fmt.Sprintf("IHDR CORRUPT\nWIDTH=%d HEIGHT=%d DEPTH=%d COLOR=%d INTERLACE=%d\nFIXED_CRC: 0x%08X",
		w, h, dep, col, ilace, fixed)
	s.expected(id, strings.TrimRight(p1.String(), "\n"), p2)
}

func (s *set) gen013() {
	id := "013"
	const sr = 8000
	digits := "519180"
	sig := sol.SynthDTMF(digits, sr, 100)
	wav := make([]byte, 44+2*len(sig))
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
	u32le(wav[40:], uint32(2*len(sig)))
	for i, v := range sig {
		sv := int16(v * 10000)
		u16le(wav[44+2*i:], uint16(sv))
	}
	s.file(id, "dial.wav", wav)

	p1 := fmt.Sprintf("RATE=%d CH=1 BITS=16 SAMPLES=%d DURATION=%.3fs", sr, len(sig), float64(len(sig))/float64(sr))
	dec := sol.DecodeDTMF(sig, sr, 100)
	keys := strings.Join(strings.Split(dec, ""), " ")
	p2 := "KEYS: " + keys
	s.expected(id, p1, p2)
}

func (s *set) gen014() {
	id := "014"
	files := []struct {
		name  string
		size  int
		badCRC bool
	}{
		{"FLIGHT01.LOG", 128, false},
		{"AUDIO.CHUNK", 64, true},
	}
	var box []byte
	box = append(box, 'B', 'O', 'X', 'D')
	var ver [2]byte
	u16le(ver[:], 1)
	box = append(box, ver[:]...)
	var nf [2]byte
	u16le(nf[:], uint16(len(files)))
	box = append(box, nf[:]...)

	tableOff := len(box)
	_ = tableOff
	dataStart := 8 + 28*len(files)
	type entry struct {
		name string
		size int
		crc  uint32
		off  int
		body []byte
	}
	var entries []entry
	off := dataStart
	for _, f := range files {
		body := make([]byte, f.size)
		for i := range body {
			body[i] = byte((i*13 + 7) & 0xFF)
		}
		crc := sol.CRC32IEEE(body)
		if f.badCRC {
			crc++
		}
		entries = append(entries, entry{f.name, f.size, crc, off, body})
		off += f.size
	}
	for _, e := range entries {
		var nm [16]byte
		copy(nm[:], e.name)
		box = append(box, nm[:]...)
		var b [12]byte
		u32le(b[0:], uint32(e.size))
		u32le(b[4:], e.crc)
		u32le(b[8:], uint32(e.off))
		box = append(box, b[:]...)
	}
	for _, e := range entries {
		box = append(box, e.body...)
	}
	s.file(id, "box.bin", box)

	var p1, p2 strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&p1, "%-14s size=%d crc=0x%08X\n", e.name, e.size, e.crc)
		got := sol.CRC32IEEE(e.body)
		if got == e.crc {
			fmt.Fprintf(&p2, "%-14s EXTRACTED (CRC OK)\n", e.name)
		} else {
			fmt.Fprintf(&p2, "%-14s EXTRACTED (CRC MISMATCH expected=0x%08X got=0x%08X)\n", e.name, e.crc, got)
		}
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen015() {
	id := "015"
	tile0 := [][]int{
		{0, 0, 0, 0, 1, 1, 0, 0},
		{0, 0, 0, 1, 2, 1, 0, 0},
		{0, 0, 1, 2, 2, 1, 0, 0},
		{0, 1, 2, 2, 2, 1, 0, 0},
		{1, 2, 2, 2, 1, 0, 0, 0},
		{0, 1, 2, 1, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	sprites := make([]byte, 16*16*64)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			sprites[y*8+x] = byte(tile0[y][x])
		}
	}
	s.file(id, "sprites.bin", sprites)

	var p1 strings.Builder
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x > 0 {
				p1.WriteByte(' ')
			}
			p1.WriteString(fmt.Sprintf("%x", tile0[y][x]))
		}
		p1.WriteByte('\n')
	}
	chars := " .oO#@"
	var p2 strings.Builder
	for y := 0; y < 8; y++ {
		for x := 0; x < 16; x++ {
			var idx byte
			if x < 8 {
				idx = sprites[y*8+x]
			}
			ci := int(idx)
			if ci > 5 {
				ci = 5
			}
			p2.WriteByte(chars[ci])
		}
		p2.WriteByte('\n')
	}
	for y := 0; y < 16; y++ {
		p2.WriteString(strings.Repeat(" ", 16) + "\n")
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

func (s *set) gen016() {
	id := "016"
	// part 1 file: 3 blocks of 256 (255 data + 1 checksum byte)
	term := make([]byte, 3*256)
	for blk := 0; blk < 3; blk++ {
		base := blk * 256
		for i := 0; i < 255; i++ {
			term[base+i] = byte(blk*10 + i)
		}
		term[base+255] = ^byte(byte(term[base+0] + term[base+1] + term[base+2])) & 0xFF // placeholder, fixed below
	}
	for blk := 0; blk < 3; blk++ {
		base := blk * 256
		var sum byte
		for i := 0; i < 255; i++ {
			sum += term[base+i]
		}
		term[base+255] = ^sum & 0xFF
	}
	term[1*256+255] ^= 0x01 // corrupt block 1
	s.file(id, "term.bin", term)

	var p1 strings.Builder
	for blk := 0; blk < 3; blk++ {
		base := blk * 256
		var sum byte
		for i := 0; i < 255; i++ {
			sum += term[base+i]
		}
		want := ^sum & 0xFF
		if want == term[base+255] {
			fmt.Fprintf(&p1, "BLOCK %02d checksum OK\n", blk)
		} else {
			fmt.Fprintf(&p1, "BLOCK %02d checksum BAD (got=0x%02X want=0x%02X)\n", blk, term[base+255], want)
		}
	}

	// part 2 file: 3 blocks of 256 (4 addr + 250 data + 2 crc16)
	term2 := make([]byte, 3*256)
	for blk := 0; blk < 3; blk++ {
		base := blk * 256
		u32le(term2[base:], uint32(blk*256))
		for i := 0; i < 250; i++ {
			term2[base+4+i] = byte(blk*17 + i)
		}
		crc := sol.CRC16CCITT(term2[base : base+254])
		u16le(term2[base+254:], crc)
	}
	term2[2*256+254] ^= 0xFF // corrupt crc of block 2
	s.file(id, "term2.bin", term2)

	var p2 strings.Builder
	for blk := 0; blk < 3; blk++ {
		base := blk * 256
		crc := sol.CRC16CCITT(term2[base : base+254])
		got := binary.LittleEndian.Uint16(term2[base+254:])
		addr := binary.LittleEndian.Uint32(term2[base:])
		if crc == got {
			fmt.Fprintf(&p2, "GOOD: 0x%08X\n", addr)
		} else {
			fmt.Fprintf(&p2, "BAD: 0x%08X\n", addr)
		}
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

var glyphs = map[byte][]byte{
	'G': {0x70, 0x88, 0x80, 0xB8, 0x88, 0x88, 0x70, 0x00},
	'R': {0xF0, 0x88, 0x88, 0xF0, 0xA0, 0x90, 0x88, 0x00},
	'I': {0xFC, 0x10, 0x10, 0x10, 0x10, 0x10, 0xFC, 0x00},
	'D': {0xF0, 0x88, 0x88, 0x88, 0x88, 0x88, 0xF0, 0x00},
	'N': {0x80, 0xC0, 0xA0, 0x90, 0x88, 0x84, 0x82, 0x00},
	'E': {0xFC, 0x80, 0x80, 0xF0, 0x80, 0x80, 0xFC, 0x00},
	'O': {0x70, 0x88, 0x88, 0x88, 0x88, 0x88, 0x70, 0x00},
}

func (s *set) gen017() {
	id := "017"
	prgBlocks, chrBlocks := byte(1), byte(2)
	rom := make([]byte, 16+int(prgBlocks)*16384+int(chrBlocks)*8192)
	copy(rom[0:4], []byte{'N', 'E', 'S', 0x1A})
	rom[4], rom[5] = prgBlocks, chrBlocks
	rom[6] = 0x04 | 0x20 // mapper low 4, battery bit5
	rom[7] = 0x00        // NTSC
	// CHR bank 0: tiles 0..3 = "GRID"
	chr := rom[16+int(prgBlocks)*16384:]
	msg := "GRID"
	for i, ch := range []byte(msg) {
		glyph := glyphs[ch]
		tile := i * 16
		copy(chr[tile:tile+8], glyph)
	}
	s.file(id, "cart.nes", rom)

	p1 := "MAPPER=4 PRG=16KB*1 CHR=8KB*2 TRAINER=no BATTERY=yes REGION=NTSC"

	render := func(bank int) string {
		base := bank * 8192
		var sb strings.Builder
		for y := 0; y < 8; y++ {
			for t := 0; t < 16; t++ {
				row := chr[base+t*16+y]
				for b := 7; b >= 0; b-- {
					if row&(1<<b) != 0 {
						sb.WriteByte('#')
					} else {
						sb.WriteByte('.')
					}
				}
			}
			sb.WriteByte('\n')
		}
		return strings.TrimRight(sb.String(), "\n")
	}
	p2 := "BANK 0\n" + render(0) + "\nBANK 1\n" + render(1)
	s.expected(id, p1, p2)
}

func (s *set) gen018() {
	id := "018"
	font := make([]byte, 256*8)
	for ch, g := range glyphs {
		copy(font[int(ch)*8:], g)
	}
	screen := make([]byte, 4000)
	word := "NEON GRID"
	for i, ch := range []byte(word) {
		screen[i] = ch
	}
	for i := 0; i < 9; i++ {
		screen[2000+i] = 0x70 // bg 7 for the word area
	}
	for i := 160; i < 160+100; i++ {
		screen[2000+i] = 0x10 // bg 1 for a block
	}
	s.file(id, "screen.bin", screen)
	s.file(id, "cga_font.bin", font)

	var p1 strings.Builder
	for sy := 0; sy < 25; sy++ {
		for py := 0; py < 8; py++ {
			for cx := 0; cx < 80; cx++ {
				ch := screen[sy*80+cx]
				g := font[int(ch)*8+py]
				for b := 7; b >= 0; b-- {
					if g&(1<<b) != 0 {
						p1.WriteString("##")
					} else {
						p1.WriteString("..")
					}
				}
			}
			p1.WriteByte('\n')
		}
	}
	counts := map[int]int{}
	for i := 0; i < 2000; i++ {
		bg := int(screen[2000+i] >> 4 & 7)
		counts[bg]++
	}
	order := []int{0, 1, 7}
	var line strings.Builder
	line.WriteString("BG COLOR COUNTS:")
	for _, c := range order {
		fmt.Fprintf(&line, " %d:%d", c, counts[c])
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), line.String())
}

func (s *set) gen019() {
	id := "019"
	save := make([]byte, 32+5*4)
	u16le(save[0:], 1)
	copy(save[2:18], "CROW_001")
	u16le(save[18:], 100) // HP
	u16le(save[20:], 50)  // MP
	u32le(save[22:], 9999)
	save[26] = 7
	u32le(save[27:], 0x8039) // XP with bit15 -> cheater flag
	save[31] = 5
	items := [][2]uint16{{3, 1}, {77, 12}, {5, 2}, {1000, 1}, {200, 9999}}
	for i, it := range items {
		u16le(save[32+i*4:], it[0])
		u16le(save[34+i*4:], it[1])
	}
	s.file(id, "save.bin", save)

	p1 := "HERO=CROW_001 HP=100/100 MP=50/50 GOLD=9999 LVL=7 XP=32825\nITEMS: {3: 1, 77: 12, 5: 2, 1000: 1, 200: 9999}"
	p2 := "DIRTY: {1000: 1, 200: 9999}\nGOLD: 9999\nXP: 32825 CHEATER_FLAG"
	s.expected(id, p1, p2)
}

func makeBMP24(w, h int, seed byte) []byte {
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
	for i := range bmp[54:] {
		bmp[54+i] = byte(seed*7 + byte(i)*3)
	}
	return bmp
}

func (s *set) gen020() {
	id := "020"
	host := makeBMP24(16, 16, 1)
	msg := "MEET_AT_NEON_ALLEY_3AM"
	payload := append(make([]byte, 4), msg...)
	u32le(payload[0:], uint32(len(msg)))

	embed := func(bmp []byte, payload []byte) []byte {
		out := append([]byte{}, bmp...)
		off := 54
		bits := len(payload) * 8
		for i := 0; i < bits; i++ {
			byteIdx := i / 8
			bit := (payload[byteIdx] >> (7 - uint(i)%8)) & 1
			idx := off + i
			out[idx] = out[idx]&0xFE | bit
		}
		return out
	}
	extract := func(bmp []byte) []byte {
		off := 54
		var lenB [4]byte
		for i := 0; i < 32; i++ {
			bit := bmp[off+i] & 1
			lenB[i/8] |= bit << (7 - uint(i)%8)
		}
		n := int(binary.LittleEndian.Uint32(lenB[:]))
		out := make([]byte, n)
		for i := 0; i < n*8; i++ {
			bit := bmp[off+32+i] & 1
			out[i/8] |= bit << (7 - uint(i)%8)
		}
		return out
	}
	carrier := embed(host, payload)
	s.file(id, "carrier.bmp", carrier)
	s.file(id, "host.bmp", host)

	p1 := "PAYLOAD: " + msg
	_ = extract
	p2 := "WROTE stego.bmp (ok, 16x16)"
	s.expected(id, p1, p2)
}

// silence unused import warnings for math if not used elsewhere
var _ = math.Pi
