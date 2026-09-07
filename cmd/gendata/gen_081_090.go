package main

// Generators for missions 081-090 (forensics part 1). All generators for
// this group self-register below.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"neon-grid/sol"
)

func fat12Set(fat []byte, cl, v int) {
	idx := cl + cl/2
	if cl%2 == 0 {
		fat[idx] = byte(v & 0xFF)
		fat[idx+1] = (fat[idx+1] & 0xF0) | byte((v>>8)&0x0F)
	} else {
		fat[idx] = (fat[idx] & 0x0F) | byte((v<<4)&0xF0)
		fat[idx+1] = byte((v >> 4) & 0xFF)
	}
}

func (s *set) gen081() {
	id := "081"
	img := make([]byte, 2880*512)
	boot := img[:512]
	copy(boot[0:3], []byte{0xEB, 0x3C, 0x90})
	copy(boot[3:11], "NEONBOOT")
	binary.LittleEndian.PutUint16(boot[11:], 512)
	boot[13] = 1
	binary.LittleEndian.PutUint16(boot[14:], 1)
	boot[16] = 2
	binary.LittleEndian.PutUint16(boot[17:], 224)
	binary.LittleEndian.PutUint16(boot[19:], 2880)
	boot[21] = 0xF0
	binary.LittleEndian.PutUint16(boot[22:], 9)
	boot[510], boot[511] = 0x55, 0xAA

	const spf, reserved, numFAT = 9, 1, 2
	fat := make([]byte, spf*512)
	fat12Set(fat, 0, 0xFF0)
	fat12Set(fat, 1, 0xFFF)
	fat12Set(fat, 2, 0xFFF)
	copy(img[reserved*512:(reserved+spf)*512], fat)
	copy(img[(reserved+spf)*512:(reserved+2*spf)*512], fat)

	rootStart := reserved + numFAT*spf
	dir := img[rootStart*512:]
	entry := dir[:32]
	copy(entry[0:8], "SECRET")
	copy(entry[8:11], "TXT")
	entry[11] = 0x20
	binary.LittleEndian.PutUint16(entry[26:], 2)
	binary.LittleEndian.PutUint32(entry[28:], 512)

	dataStart := rootStart + 224*32/512
	content := "THE COURIER MOVES AT MIDNIGHT"
	cluster := img[dataStart*512 : dataStart*512+512]
	copy(cluster, content)

	s.file(id, "floppy.img", img)

	rootDirSec := (224*32 + 511) / 512
	clusters := (2880 - (reserved + numFAT*spf + rootDirSec)) / 1
	part1 := fmt.Sprintf("FAT12 1440KB  bytes/sector=512  clusters=%d  root=224", clusters)
	part2 := fmt.Sprintf("SECRET.TXT (512 bytes)\n%s", content)
	s.expected(id, part1, part2)
}

func (s *set) gen082() {
	id := "082"
	var buf bytes.Buffer
	eh := make([]byte, 64)
	copy(eh[0:4], "\x7fELF")
	eh[4], eh[5], eh[6] = 2, 1, 1
	binary.LittleEndian.PutUint16(eh[16:], 3)
	binary.LittleEndian.PutUint16(eh[18:], 62)
	binary.LittleEndian.PutUint32(eh[20:], 1)
	binary.LittleEndian.PutUint64(eh[24:], 0x401000)
	binary.LittleEndian.PutUint64(eh[32:], 64)
	binary.LittleEndian.PutUint64(eh[40:], 0x1000)
	binary.LittleEndian.PutUint16(eh[52:], 64)
	binary.LittleEndian.PutUint16(eh[54:], 56)
	binary.LittleEndian.PutUint16(eh[56:], 1)
	binary.LittleEndian.PutUint16(eh[58:], 64)
	binary.LittleEndian.PutUint16(eh[60:], 13)
	binary.LittleEndian.PutUint16(eh[62:], 3)
	buf.Write(eh)

	ph := make([]byte, 56)
	binary.LittleEndian.PutUint32(ph[0:], 1)
	binary.LittleEndian.PutUint32(ph[4:], 5)
	binary.LittleEndian.PutUint64(ph[8:], 0)
	binary.LittleEndian.PutUint64(ph[16:], 0x400000)
	binary.LittleEndian.PutUint64(ph[24:], 0x400000)
	binary.LittleEndian.PutUint64(ph[32:], 0x1000)
	binary.LittleEndian.PutUint64(ph[40:], 0x1000)
	binary.LittleEndian.PutUint64(ph[48:], 0x1000)
	buf.Write(ph)
	buf.Write(make([]byte, 0x1000-buf.Len()))

	const numSections = 13
	shoff := buf.Len()
	for i := 0; i < numSections; i++ {
		buf.Write(make([]byte, 64))
	}
	nameBlob := "\x00.text\x00.rodata\x00.shstrtab\x00.symtab\x00.strtab\x00.data\x00.bss\x00.comment\x00.note.GNU-stack\x00.dynsym\x00.dynstr\x00.init\x00"
	nameOff := map[string]int{}
	names := []string{".text", ".rodata", ".shstrtab", ".symtab", ".strtab", ".data", ".bss", ".comment", ".note.GNU-stack", ".dynsym", ".dynstr", ".init"}
	pos := 0
	for _, n := range names {
		idx := strings.Index(nameBlob[pos:], n)
		nameOff[n] = pos + idx
		pos += idx + len(n) + 1
	}
	shstrOff := buf.Len()
	buf.Write([]byte(nameBlob))

	textOff := 0x3000
	rodataOff := 0x2000
	rodata := make([]byte, 0x80)
	copy(rodata, "FLAG: crow_was_here_2049\x00")
	text := make([]byte, 0x100)
	copy(text, []byte{0x48, 0x8b, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00})

	writeSection := func(i int, name string, typ uint32, flags uint64, addr, off, size uint64, align uint64) {
		sh := buf.Bytes()[shoff+i*64 : shoff+(i+1)*64]
		binary.LittleEndian.PutUint32(sh[0:], uint32(nameOff[name]))
		binary.LittleEndian.PutUint32(sh[4:], typ)
		binary.LittleEndian.PutUint64(sh[8:], flags)
		binary.LittleEndian.PutUint64(sh[16:], addr)
		binary.LittleEndian.PutUint64(sh[24:], off)
		binary.LittleEndian.PutUint64(sh[32:], size)
		binary.LittleEndian.PutUint64(sh[48:], align)
	}
	writeSection(1, ".text", 1, 6, 0x401000, uint64(textOff), 0x100, 16)
	writeSection(2, ".rodata", 1, 2, 0x402000, uint64(rodataOff), 0x80, 16)
	writeSection(3, ".shstrtab", 3, 0, 0, uint64(shstrOff), uint64(len(nameBlob)), 1)
	for i := 4; i < numSections; i++ {
		writeSection(i, names[i-1], 1, 0, 0, 0, 0, 0)
	}

	buf.Write(make([]byte, rodataOff-buf.Len()))
	buf.Write(rodata)
	buf.Write(make([]byte, textOff-buf.Len()))
	buf.Write(text)

	s.file(id, "omega.elf", buf.Bytes())

	part1 := "ELF64 LE  type=ET_DYN  machine=x86-64  entry=0x401000  shdr=13"
	part2 := "SECTION .rodata (offset=0x2000 size=0x80)\nFLAG: crow_was_here_2049\n.text entry bytes: 48 8b 05 00 00 00 00 00"
	s.expected(id, part1, part2)
}

func (s *set) gen083() {
	id := "083"
	img := make([]byte, 0x7100)
	copy(img[0:2], "MZ")
	binary.LittleEndian.PutUint32(img[0x3C:], 0x80)
	pe := 0x80
	copy(img[pe:pe+4], "PE\x00\x00")
	binary.LittleEndian.PutUint16(img[pe+4:], 0x14C)
	binary.LittleEndian.PutUint16(img[pe+6:], 5)
	binary.LittleEndian.PutUint16(img[pe+14:], 0xE0)
	binary.LittleEndian.PutUint16(img[pe+16:], 0x0102)
	opt := pe + 24
	binary.LittleEndian.PutUint16(img[opt:], 0x10B)
	binary.LittleEndian.PutUint32(img[opt+0x10:], 0x401000)
	binary.LittleEndian.PutUint16(img[opt+0x68:], 2)
	sect := opt + 0xE0
	sections := [][]string{{".text", "0x800", "0x1000", "0x800", "0x1000"}, {".data", "0x200", "0x3000", "0x200", "0x3000"}, {".rdata", "0x100", "0x5000", "0x100", "0x5000"}, {".pdata", "0x100", "0x6000", "0x100", "0x6000"}, {".reloc", "0x100", "0x7000", "0x100", "0x7000"}}
	for i, sec := range sections {
		e := sect + i*40
		copy(img[e:e+8], sec[0])
		binary.LittleEndian.PutUint32(img[e+8:], uint32(parseHexVal(sec[1])))
		binary.LittleEndian.PutUint32(img[e+12:], uint32(parseHexVal(sec[2])))
		binary.LittleEndian.PutUint32(img[e+16:], uint32(parseHexVal(sec[3])))
		binary.LittleEndian.PutUint32(img[e+20:], uint32(parseHexVal(sec[4])))
	}
	rdata := img[0x5000:]
	copy(rdata, "pw_let_the_raven_out\x00pass_is_not_here\x00OMEGA-DYNE\x00")
	s.file(id, "camera.exe", img)

	part1 := "PE32 x86  sections=5  entry=0x401000  subsys=GUI"
	part2 := ".rdata strings:\n  pw_let_the_raven_out\n  pass_is_not_here\n  OMEGA-DYNE\nFOUND: pw_let_the_raven_out"
	s.expected(id, part1, part2)
}

func parseHexVal(s string) uint64 {
	var v uint64
	fmt.Sscanf(strings.TrimPrefix(s, "0x"), "%X", &v)
	return v
}

func (s *set) gen084() {
	id := "084"
	const sec = 512
	img := make([]byte, 4097*sec)
	mbr := img[:sec]
	mbr[446+0] = 0x80
	mbr[446+4] = 0x0C
	binary.LittleEndian.PutUint32(mbr[446+8:], 2048)
	binary.LittleEndian.PutUint32(mbr[446+12:], 614400)
	mbr[446+16] = 0x00
	mbr[446+20] = 0x83
	binary.LittleEndian.PutUint32(mbr[446+24:], 616448)
	binary.LittleEndian.PutUint32(mbr[446+28:], 204800)
	mbr[510], mbr[511] = 0x55, 0xAA

	gpt := img[sec : 2*sec]
	copy(gpt[0:8], "EFI PART")
	binary.LittleEndian.PutUint32(gpt[8:], 0x00010000)
	binary.LittleEndian.PutUint32(gpt[12:], 92)
	binary.LittleEndian.PutUint64(gpt[24:], 1)
	binary.LittleEndian.PutUint64(gpt[32:], 4096)
	binary.LittleEndian.PutUint64(gpt[40:], 3)
	binary.LittleEndian.PutUint64(gpt[48:], 4095)
	binary.LittleEndian.PutUint64(gpt[72:], 2)
	binary.LittleEndian.PutUint32(gpt[80:], 4)
	binary.LittleEndian.PutUint32(gpt[84:], 128)
	for i := 0; i < 4; i++ {
		ent := img[(2+i)*sec : (3+i)*sec]
		if i == 2 {
			binary.LittleEndian.PutUint64(ent[32:], 2048)
			binary.LittleEndian.PutUint64(ent[40:], 4096)
			u16 := utf16.Encode([]rune("MERCURY"))
			for j, r := range u16 {
				binary.LittleEndian.PutUint16(ent[56+j*2:], r)
			}
		}
	}
	s.file(id, "disk.img", img)

	part1 := "PT 0: type=0x0C (FAT32 LBA) start=2048 sectors=614400 size=300.0MB\n" +
		"PT 1: type=0x83 (Linux)     start=616448 sectors=204800 size=100.0MB"
	part2 := "PARTITION MERCURY: LBA 2048..4096 (1.0MB)"
	s.expected(id, part1, part2)
}

func (s *set) gen085() {
	id := "085"
	dump := make([]byte, 4*1024*1024)
	for i := range dump {
		dump[i] = 0xAA
	}
	// JPEG at 0x00000000, 89222 bytes.
	jpeg := dump[:89222]
	copy(jpeg[0:3], []byte{0xFF, 0xD8, 0xFF})
	copy(jpeg[len(jpeg)-2:], []byte{0xFF, 0xD9})
	for i := 3; i < len(jpeg)-2; i++ {
		jpeg[i] = 0xAB
	}
	// PNG at 0x00123ABC, 12345 bytes (8 sig + filler + IEND chunk).
	pngStart := 0x00123ABC
	png := dump[pngStart : pngStart+12345]
	copy(png[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	copy(png[len(png)-12:], []byte{0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82})
	for i := 8; i < len(png)-12; i++ {
		png[i] = 0xCD
	}
	// ZIP at 0x00300000, carved 4051 bytes (PK0304 + filler + PK0102 + 22 EOCD).
	zipStart := 0x00300000
	zipEnd := zipStart + 4055
	z := dump[zipStart:zipEnd]
	copy(z[0:4], []byte{0x50, 0x4B, 0x03, 0x04})
	for i := 4; i < 4+4025; i++ {
		z[i] = 0xEF
	}
	copy(z[4+4025:4+4025+4], []byte{0x50, 0x4B, 0x01, 0x02})
	copy(z[4+4025+4:], []byte{0x50, 0x4B, 0x05, 0x06, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	s.file(id, "dump.bin", dump)

	part1 := "JPEG @ 0x00000000\nPNG  @ 0x00123ABC\nZIP  @ 0x00300000"
	part2 := "carved_0.jpg 89222 bytes\ncarved_1.png 12345 bytes\ncarved_2.zip 4051 bytes"
	s.expected(id, part1, part2)
}

func (s *set) gen086() {
	id := "086"
	ops := []string{"SET a=1", "SET b=2", "SET c=9", "SET c=4", "SET c=7", "DEL c", "SET a=1", "SET b=2", "SET c=5", "DEL c"}
	var wal []byte
	for _, op := range ops {
		rec := make([]byte, 6+len(op))
		binary.LittleEndian.PutUint16(rec[0:], 0x4A4B)
		binary.LittleEndian.PutUint16(rec[2:], uint16(len(op)))
		binary.LittleEndian.PutUint16(rec[4:], sol.CRC16CCITT([]byte(op)))
		copy(rec[6:], op)
		wal = append(wal, rec...)
	}
	trunc := make([]byte, 6+3)
	binary.LittleEndian.PutUint16(trunc[0:], 0x4A4B)
	binary.LittleEndian.PutUint16(trunc[2:], 8)
	binary.LittleEndian.PutUint16(trunc[4:], 0)
	copy(trunc[6:], "SET")
	wal = append(wal, trunc...)
	s.file(id, "wal.bin", wal)

	part1 := "WAL: 10 ok, 1 corrupted\nFINAL: a=1 b=2 c=DELETED"
	part2 := "APPLIED: 10  SKIPPED: 1 (truncated tail)\nSNAPSHOT: a=1 b=2 c=DELETED (3 entries)"
	s.expected(id, part1, part2)
}

func (s *set) gen087() {
	id := "087"
	const stripe = 4096
	const stripes = 8
	const msg = "THE_PURGE_BEGINS_AT_DUSK"
	disks := make([][]byte, 4)
	for d := 0; d < 4; d++ {
		disks[d] = make([]byte, stripe*stripes)
	}
	live := func(d int) byte { return byte(0x30 + d) }
	// live disks carry a per-disk filler pattern.
	for d := 0; d < 4; d++ {
		if d == 2 {
			continue
		}
		for i := range disks[d] {
			disks[d][i] = live(d)
		}
	}
	// disk1 block of stripe0 carries header+message.
	blk := disks[1][0:stripe]
	binary.LittleEndian.PutUint32(blk[0:], uint32(len(msg)))
	copy(blk[4:], msg)
	// original disk2[0]=0x4D.
	disks[2][0] = 0x4D
	// compute parity per stripe on disk s%4.
	for s := 0; s < stripes; s++ {
		p := s % 4
		par := make([]byte, stripe)
		for d := 0; d < 4; d++ {
			if d == p {
				continue
			}
			blk := disks[d][s*stripe : (s+1)*stripe]
			for i := 0; i < stripe; i++ {
				par[i] ^= blk[i]
			}
		}
		copy(disks[p][s*stripe:(s+1)*stripe], par)
	}
	// disk2 is dead: zeroed.
	for i := range disks[2] {
		disks[2][i] = 0
	}
	for d := 0; d < 4; d++ {
		s.file(id, fmt.Sprintf("disk%d.bin", d), disks[d])
	}

	part1 := "DEAD: disk2\nRESTORED: disk2[0]=0x4D"
	part2 := fmt.Sprintf("RAID.TXT: %s", msg)
	s.expected(id, part1, part2)
}

func (s *set) gen088() {
	id := "088"
	var tb bytes.Buffer
	tw := tar.NewWriter(&tb)
	notes := "  The archive is the key.\n  CRC of manifest: 0x1234ABCD\n"
	f2 := bytes.Repeat([]byte{0xF2}, 10000)
	f3 := bytes.Repeat([]byte{0xF3}, 2289)
	add := func(name string, typ byte, data []byte) {
		hdr := &tar.Header{Name: name, Mode: 0o644, Typeflag: typ, Size: int64(len(data))}
		if err := tw.WriteHeader(hdr); err != nil {
			panic(err)
		}
		tw.Write(data)
	}
	add("VAULT/", tar.TypeDir, nil)
	add("NOTES.TXT", tar.TypeReg, []byte(notes))
	add("F2.BIN", tar.TypeReg, f2)
	add("F3.BIN", tar.TypeReg, f3)
	if err := tw.Close(); err != nil {
		panic(err)
	}
	tarBytes := tb.Bytes()

	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	zw.ModTime = time.Unix(0, 0) // fixed mtime: deterministic output
	zw.Write(tarBytes)
	if err := zw.Close(); err != nil {
		panic(err)
	}
	s.file(id, "payload.tar.gz", gz.Bytes())

	part1 := fmt.Sprintf("PAYLOAD: %d bytes (deflate ok)", len(tarBytes))
	part2 := "NOTES.TXT:\n" + strings.TrimRight(notes, "\n") + "\nEXTRACTED: 3 files, 12345 bytes total"
	s.expected(id, part1, part2)
}

type diffBlock struct{ off, length int }

func diffBlocks(old, new []byte) []diffBlock {
	var blocks []diffBlock
	cur := -1
	for i := 0; i < len(old); i++ {
		if old[i] != new[i] {
			if cur == -1 {
				cur = i
			}
			continue
		}
		if cur != -1 {
			blocks = append(blocks, diffBlock{cur, i - cur})
			cur = -1
		}
	}
	if cur != -1 {
		blocks = append(blocks, diffBlock{cur, len(old) - cur})
	}
	var merged []diffBlock
	for _, b := range blocks {
		if len(merged) > 0 && b.off-(merged[len(merged)-1].off+merged[len(merged)-1].length) < 4 {
			last := &merged[len(merged)-1]
			last.length = b.off + b.length - last.off
		} else {
			merged = append(merged, b)
		}
	}
	return merged
}

func (s *set) gen089() {
	id := "089"
	old := make([]byte, 48)
	for i := range old {
		old[i] = byte(0x40 + i)
	}
	newb := append([]byte{}, old...)
	copy(newb[0:4], []byte{0x01, 0x02, 0x03, 0x04})
	copy(newb[16:18], []byte{0xDE, 0xAD})
	copy(newb[32:40], bytes.Repeat([]byte{0xAA}, 8))
	s.file(id, "old.bin", old)
	s.file(id, "new.bin", newb)

	blocks := diffBlocks(old, newb)
	var parts []string
	for _, b := range blocks {
		parts = append(parts, fmt.Sprintf("offset=%d len=%d", b.off, b.length))
	}
	part1 := fmt.Sprintf("PATCH: %d blocks (%s)\nROUNDTRIP OK", len(blocks), strings.Join(parts, ", "))
	part2 := "PATCHES: 5  TOTAL_BYTES: 2048 (base=4096)\nRECONSTRUCT v5: OK (identical)"
	s.expected(id, part1, part2)

	base := bytes.Repeat([]byte{0xFF}, 4096)
	s.file(id, "base.bin", base)
	prev := base
	for i := 1; i <= 5; i++ {
		v := append([]byte{}, prev...)
		lo, hi := (i-1)*410, i*410
		if hi > 2048 {
			hi = 2048
		}
		for j := lo; j < hi; j++ {
			v[j] = byte(j + 1)
		}
		s.file(id, fmt.Sprintf("v%d.bin", i), v)
		prev = v
	}
}

func (s *set) gen090() {
	id := "090"
	stringsList := []string{
		"ssh crow@10.0.0.7 -p 2200",
		"password = \"l33t_r4v3n\"",
		"/home/raven/drop/keys.pem",
	}
	mem := make([]byte, 0x4000)
	off := 0
	for _, str := range stringsList {
		copy(mem[off:], str)
		off += len(str) + 1
	}
	// non-matching long strings to test the filter.
	other := "Reallocating heap pages"
	copy(mem[off:], other)
	s.file(id, "mem.dmp", mem)
	s.text(id, "keys.pem", "ssh-rsa AAAAB3NzaC1yc2E phoenix@raven")

	logs := []string{
		"2039-01-02 03:04:05 login user=phoenix ip=10.0.0.7 key=AAAAB3NzaC1yc2E",
		"2039-01-02 03:45:00 login user=phoenix ip=10.0.0.7 key=AAAAB3NzaC1yc2E",
		"2039-01-03 00:12:33 logout user=phoenix ip=10.0.0.7 key=AAAAB3NzaC1yc2E",
		"2039-01-03 01:00:00 login user=crow ip=10.0.0.9 key=ZZZZZZZZZZZZ",
		"2039-01-03 02:00:00 logout user=crow ip=10.0.0.9 key=ZZZZZZZZZZZZ",
	}
	s.text(id, "auth.log", strings.Join(logs, "\n"))

	part1 := "FOUND:\n  ssh crow@10.0.0.7 -p 2200\n  password = \"l33t_r4v3n\"\n  /home/raven/drop/keys.pem"
	part2 := "2039-01-02 03:04:05 login 10.0.0.7\n2039-01-02 03:45:00 login 10.0.0.7\n2039-01-03 00:12:33 logout 10.0.0.7"
	s.expected(id, part1, part2)
}

func init() {
	register("081", (*set).gen081)
	register("082", (*set).gen082)
	register("083", (*set).gen083)
	register("084", (*set).gen084)
	register("085", (*set).gen085)
	register("086", (*set).gen086)
	register("087", (*set).gen087)
	register("088", (*set).gen088)
	register("089", (*set).gen089)
	register("090", (*set).gen090)
}
