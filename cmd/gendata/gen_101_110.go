package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

// ---------------------------------------------------------------- 101 strings

func (s *set) gen101() {
	id := "101"
	blob := make([]byte, 0x200)
	copy(blob[0x000:], "NEON GRID\x00")
	copy(blob[0x010:], "admin\x00")
	copy(blob[0x020:], "FLAG{strings_are_easy}\x00")
	copy(blob[0x040:], "passwd_hash\x00")
	utf16 := []byte{}
	for _, ch := range []byte("session_key") {
		utf16 = append(utf16, ch, 0x00)
	}
	copy(blob[0x100:], utf16)
	copy(blob[0x140:], "enc:70757267655f746f6b656e\x00") // hex of "purge_token"
	s.file(id, "blob.bin", blob)

	printable := func(b byte) bool { return b >= 0x20 && b <= 0x7E }
	// part 1: printable ASCII runs of length >= 4
	var p1 []string
	for off := 0; off < len(blob); {
		if !printable(blob[off]) {
			off++
			continue
		}
		end := off
		for end < len(blob) && printable(blob[end]) {
			end++
		}
		if end-off >= 4 {
			p1 = append(p1, fmt.Sprintf("0x%08X: %s", off, blob[off:end]))
		}
		off = end
	}

	// part 2: filter FLAG/key/pass/token (case-insensitive), plus utf16/decoded
	keyword := func(text string) bool {
		t := strings.ToLower(text)
		for _, k := range []string{"flag", "key", "pass", "token"} {
			if strings.Contains(t, k) {
				return true
			}
		}
		return false
	}
	var p2 []string
	for off := 0; off < len(blob); {
		b := blob[off]
		// UTF-16LE first: printable byte at even offset followed by 0x00
		if b >= 0x20 && b <= 0x7E && off+1 < len(blob) && blob[off+1] == 0x00 {
			end := off
			for end+1 < len(blob) && blob[end] >= 0x20 && blob[end] <= 0x7E && blob[end+1] == 0x00 {
				end += 2
			}
			var text []byte
			for i := off; i < end; i += 2 {
				text = append(text, blob[i])
			}
			if len(text) >= 4 && keyword(string(text)) {
				p2 = append(p2, fmt.Sprintf("0x%08X: utf16: %s", off, text))
			}
			off = end
			continue
		}
		if printable(b) {
			end := off
			for end < len(blob) && printable(blob[end]) {
				end++
			}
			if end-off >= 4 {
				text := string(blob[off:end])
				if strings.HasPrefix(text, "enc:") {
					if payload, err := hex.DecodeString(text[4:]); err == nil && keyword(string(payload)) {
						p2 = append(p2, fmt.Sprintf("0x%08X: decoded: %s", off, payload))
					}
				} else if keyword(text) {
					p2 = append(p2, fmt.Sprintf("0x%08X: %s", off, text))
				}
			}
			off = end
			continue
		}
		off++
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 102 disassembler

func (s *set) gen102() {
	id := "102"
	prog := []byte{0x00, 0x01, 0x00, 0x03, 0x01, 0x34, 0x12, 0x04, 0x02, 0x05, 0x03,
		0x02, 0x03, 0x00, 0x02, 0xFF, 0x00, 0x06}
	s.file(id, "code.bin", prog)
	regs := "ABXY"
	var p1 []string
	var jmps []string
	pc := 0
	for pc < len(prog) {
		op := prog[pc]
		switch op {
		case 0x00:
			p1 = append(p1, fmt.Sprintf("0x%04X: NOP", pc))
			pc += 1
		case 0x01:
			p1 = append(p1, fmt.Sprintf("0x%04X: LDA #%c", pc, regs[prog[pc+1]&0x0F]))
			pc += 2
		case 0x02:
			addr := binary.LittleEndian.Uint16(prog[pc+1:])
			p1 = append(p1, fmt.Sprintf("0x%04X: JMP 0x%04X", pc, addr))
			jmps = append(jmps, fmt.Sprintf("0x%04X", addr))
			pc += 3
		case 0x03:
			addr := binary.LittleEndian.Uint16(prog[pc+2:])
			p1 = append(p1, fmt.Sprintf("0x%04X: LD %c, 0x%04X", pc, regs[prog[pc+1]&0x0F], addr))
			pc += 4
		case 0x04:
			p1 = append(p1, fmt.Sprintf("0x%04X: INC %c", pc, regs[prog[pc+1]&0x0F]))
			pc += 2
		case 0x05:
			p1 = append(p1, fmt.Sprintf("0x%04X: DEC %c", pc, regs[prog[pc+1]&0x0F]))
			pc += 2
		case 0x06:
			p1 = append(p1, fmt.Sprintf("0x%04X: HLT", pc))
			pc += 1
		case 0xFF:
			p1 = append(p1, fmt.Sprintf("0x%04X: RET", pc))
			pc += 1
		default:
			p1 = append(p1, fmt.Sprintf("0x%04X: TRUNCATED", pc))
			break
		}
	}
	// jumps into the code (between ENTRY and END)
	var inCode []string
	for _, j := range jmps {
		var a uint32
		fmt.Sscanf(j, "0x%X", &a)
		if a >= 0 && a < uint32(len(prog)) {
			inCode = append(inCode, j)
		}
	}
	p2 := fmt.Sprintf("ENTRY: 0x0000\nJUMPS: %s\nEND: 0x%04X", strings.Join(inCode, " "), len(prog))
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 103 crackme

// crackmePassword satisfies check() from task 103: 8 printable chars,
// v0=0x5A5A5A5A -> v_{i+1}=rotl32(v_i ^ (c << 3i), 7) -> v8=0xCAFEBABE.
func crackmePassword() string {
	pw := "390M38aZ"
	v := uint32(0x5A5A5A5A)
	for i := 0; i < len(pw); i++ {
		v = (v ^ uint32(pw[i])<<uint(3*i))<<7 | (v^uint32(pw[i])<<uint(3*i))>>25
	}
	if v != 0xCAFEBABE {
		panic("crackme password no longer verifies")
	}
	return pw
}

func (s *set) gen103() {
	id := "103"
	pw := crackmePassword()
	s.expected(id, "PASSWORD: "+pw, "PASSWORD: "+pw+"\nCHECK: true")
}

// ---------------------------------------------------------------- 105 bitflip

func (s *set) gen105() {
	id := "105"
	lic := make([]byte, 256)
	copy(lic[0x00:], "LICENSE VALID\x00")
	for i := 0x0D; i < 0x80; i++ {
		lic[i] = 0x90
	}
	lic[0x80] = 0x74 // je +5
	lic[0x81] = 0x05
	for i := 0x82; i < 0x88; i++ {
		lic[i] = 0x90
	}
	copy(lic[0x88:], "LICENSE INVALID\x00")
	copy(lic[0x97:], []byte{0x74, 0x05, 0x90, 0x90})
	for i := 0x9B; i < len(lic); i++ {
		lic[i] = 0x90
	}
	s.file(id, "license.bin", lic)

	patch := append([]byte{}, lic...)
	patch[0x80] = 0xEB

	origSum := sha256.Sum256(lic)
	patchSum := sha256.Sum256(patch)
	p1 := "INVALID_STR @ 0x00000088\nAFTER: 74 05 90 90"
	p2 := fmt.Sprintf("PATCHED @ 0x00000080: 74 -> EB\nSHA256 ORIG: %x\nSHA256 PATCH: %x", origSum, patchSum)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 106 symtab

func (s *set) gen106() {
	id := "106"
	// .text: init (0x00, 0x40), validate_key (0x40, 0x200), decrypt (0x240, 0x100)
	text := make([]byte, 0x340)
	for i := range text {
		text[i] = 0x90
	}
	copy(text[0x00:], []byte{0xE8, 0x3B, 0x00, 0x00, 0x00}) // call -> 0x40
	copy(text[0x40:], []byte{0xE8, 0xFB, 0x01, 0x00, 0x00}) // call -> 0x240
	copy(text[0x240:], []byte{0xE8, 0xBB, 0x0D, 0x00, 0x00}) // call -> 0x1000 (unknown)

	syms := []struct {
		name  string
		off   uint32
		size  uint32
	}{
		{"init", 0x00, 0x40},
		{"validate_key", 0x40, 0x200},
		{"decrypt", 0x240, 0x100},
	}
	strtab := "\x00"
	for _, sym := range syms {
		strtab += sym.name + "\x00"
	}
	shstrtab := "\x00.text\x00.symtab\x00.strtab\x00.shstrtab\x00"

	textOff := 0x1000
	symOff := textOff + len(text)
	strOff := symOff + 4*24
	shstrOff := strOff + len(strtab)
	shOff := (shstrOff + len(shstrtab) + 7) &^ 7
	total := shOff + 5*0x40

	elf := make([]byte, total)
	copy(elf[0:4], "\x7FELF")
	elf[4], elf[5], elf[6], elf[7] = 2, 1, 1, 0 // 64-bit, LE, SYSV
	binary.LittleEndian.PutUint16(elf[16:], 2)  // ET_EXEC
	binary.LittleEndian.PutUint16(elf[18:], 0x3E)
	binary.LittleEndian.PutUint32(elf[20:], 1)
	binary.LittleEndian.PutUint64(elf[24:], 0x401000)
	binary.LittleEndian.PutUint64(elf[32:], 0x40) // phoff
	binary.LittleEndian.PutUint64(elf[40:], uint64(shOff))
	binary.LittleEndian.PutUint16(elf[52:], 0x40) // ehsize
	binary.LittleEndian.PutUint16(elf[54:], 0x38) // phentsize
	binary.LittleEndian.PutUint16(elf[56:], 1)    // phnum
	binary.LittleEndian.PutUint16(elf[58:], 0x40) // shentsize
	binary.LittleEndian.PutUint16(elf[60:], 5)    // shnum
	binary.LittleEndian.PutUint16(elf[62:], 4)    // shstrndx

	// one PT_LOAD: R+X, covers whole file, VA base 0x400000
	binary.LittleEndian.PutUint32(elf[0x40:], 1)
	binary.LittleEndian.PutUint32(elf[0x44:], 5)
	binary.LittleEndian.PutUint64(elf[0x48:], 0)
	binary.LittleEndian.PutUint64(elf[0x50:], 0x400000)
	binary.LittleEndian.PutUint64(elf[0x58:], 0x400000)
	binary.LittleEndian.PutUint64(elf[0x60:], uint64(total))
	binary.LittleEndian.PutUint64(elf[0x68:], uint64(total))
	binary.LittleEndian.PutUint64(elf[0x70:], 0x1000)

	copy(elf[textOff:], text)
	// symtab: null entry + 3 funcs
	binary.LittleEndian.PutUint32(elf[symOff+24:], 1) // init: name
	elf[symOff+24+4] = 0x12                          // STB_GLOBAL | STT_FUNC
	binary.LittleEndian.PutUint16(elf[symOff+24+6:], 1)
	binary.LittleEndian.PutUint64(elf[symOff+24+8:], 0x401000)
	binary.LittleEndian.PutUint64(elf[symOff+24+16:], 0x40)
	binary.LittleEndian.PutUint32(elf[symOff+48:], 6) // validate_key
	elf[symOff+48+4] = 0x12
	binary.LittleEndian.PutUint16(elf[symOff+48+6:], 1)
	binary.LittleEndian.PutUint64(elf[symOff+48+8:], 0x401040)
	binary.LittleEndian.PutUint64(elf[symOff+48+16:], 0x200)
	binary.LittleEndian.PutUint32(elf[symOff+72:], 19) // decrypt
	elf[symOff+72+4] = 0x12
	binary.LittleEndian.PutUint16(elf[symOff+72+6:], 1)
	binary.LittleEndian.PutUint64(elf[symOff+72+8:], 0x401240)
	binary.LittleEndian.PutUint64(elf[symOff+72+16:], 0x100)
	copy(elf[strOff:], strtab)
	copy(elf[shstrOff:], shstrtab)

	// section headers
	sh := func(idx int, name, typ, flags uint32, addr, off, size uint64, link, info, align, entsize uint32) {
		base := shOff + idx*0x40
		binary.LittleEndian.PutUint32(elf[base:], name)
		binary.LittleEndian.PutUint32(elf[base+4:], typ)
		binary.LittleEndian.PutUint64(elf[base+8:], uint64(flags))
		binary.LittleEndian.PutUint64(elf[base+16:], addr)
		binary.LittleEndian.PutUint64(elf[base+24:], off)
		binary.LittleEndian.PutUint64(elf[base+32:], size)
		binary.LittleEndian.PutUint32(elf[base+40:], link)
		binary.LittleEndian.PutUint32(elf[base+44:], info)
		binary.LittleEndian.PutUint64(elf[base+48:], uint64(align))
		binary.LittleEndian.PutUint64(elf[base+56:], uint64(entsize))
	}
	sh(1, 1, 1, 0x6, 0x401000, uint64(textOff), uint64(len(text)), 0, 0, 16, 0)          // .text
	sh(2, 7, 2, 0, 0, uint64(symOff), 4*24, 3, 1, 8, 24)                                  // .symtab
	sh(3, 15, 3, 0, 0, uint64(strOff), uint64(len(strtab)), 0, 0, 1, 0)                   // .strtab
	sh(4, 23, 3, 0, 0, uint64(shstrOff), uint64(len(shstrtab)), 0, 0, 1, 0)               // .shstrtab
	s.file(id, "elf.bin", elf)

	p1 := strings.Join([]string{
		"FUNCS:",
		"  0x401000 init         size=64",
		"  0x401040 validate_key size=512",
		"  0x401240 decrypt      size=256",
	}, "\n")
	p2 := strings.Join([]string{
		"init -> validate_key",
		"validate_key -> decrypt",
		"decrypt -> (неизвестно)",
	}, "\n")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 107 VM decompiler

var ironCoreOps = map[byte]string{
	0x01: "PUSH", 0x02: "POP", 0x03: "DUP", 0x04: "ADD", 0x05: "SUB", 0x06: "MUL",
	0x07: "AND", 0x08: "JMP", 0x09: "JZ", 0x0A: "CMP", 0x0B: "LOAD", 0x0C: "STORE",
	0x0D: "CALL", 0x0E: "RET", 0x0F: "HLT",
}

func (s *set) gen107() {
	id := "107"
	prog := []byte{
		0x01, 0x00, 0x00, 0x00, 0x00, // PUSH 0
		0x0B, 0x00, 0x00, // LOAD 0 (x0)
		0x04,       // ADD
		0x03,       // DUP
		0x04,       // ADD
		0x01, 0x03, 0x00, 0x00, 0x00, // PUSH 3
		0x04,       // ADD
		0x0C, 0x01, 0x00, // STORE 1 (x1)
		0x0F, // HLT
	}
	s.file(id, "vm.bin", prog)
	var p1 []string
	pc := 0
	for pc < len(prog) {
		op := prog[pc]
		switch op {
		case 0x01:
			p1 = append(p1, fmt.Sprintf("0x%04X: PUSH %d", pc, binary.LittleEndian.Uint32(prog[pc+1:])))
			pc += 5
		case 0x0B:
			p1 = append(p1, fmt.Sprintf("0x%04X: LOAD %d", pc, binary.LittleEndian.Uint16(prog[pc+1:])))
			pc += 3
		case 0x0C:
			p1 = append(p1, fmt.Sprintf("0x%04X: STORE %d", pc, binary.LittleEndian.Uint16(prog[pc+1:])))
			pc += 3
		case 0x0D:
			p1 = append(p1, fmt.Sprintf("0x%04X: CALL %d", pc, binary.LittleEndian.Uint16(prog[pc+1:])))
			pc += 3
		case 0x08, 0x09:
			p1 = append(p1, fmt.Sprintf("0x%04X: %s %d", pc, ironCoreOps[op], int8(prog[pc+1])))
			pc += 2
		default:
			p1 = append(p1, fmt.Sprintf("0x%04X: %s", pc, ironCoreOps[op]))
			pc += 1
		}
	}
	p2 := "x1 = 2*x0 + 3\nRESULT(5) = 13"
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 108 obfuscator

func (s *set) gen108() {
	id := "108"
	fog := "var _0x1 = (0x1F + 0x20) * 3 - 0x3C + 0x7;\n" +
		"var _0x2 = \"\\x4e\\x45\\x4f\\x4e\";\n" +
		"if (_0x2 === \"NEON\") { return true; } else { return false; }\n"
	s.text(id, "fog.js", fog)
	p1 := "var _0x1 = 136;"
	p2 := "a = 136;\nb = \"NEON\";\nreturn b === \"NEON\";"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 110 unpacker

func (s *set) gen110() {
	id := "110"
	// inner VM program (107): x1 = x0*3 + 4, RESULT(7) = 25
	prog := []byte{
		0x0B, 0x00, 0x00, // LOAD 0 (x0)
		0x01, 0x03, 0x00, 0x00, 0x00, // PUSH 3
		0x06,       // MUL
		0x01, 0x04, 0x00, 0x00, 0x00, // PUSH 4
		0x04,       // ADD
		0x0C, 0x01, 0x00, // STORE 1 (x1)
		0x0F, // HLT
	}
	const unpackedSize = 32768
	const key = "CELLOPHANE"
	unpacked := make([]byte, unpackedSize)
	copy(unpacked, prog)

	packed := make([]byte, 28+unpackedSize)
	copy(packed[0:4], "PKUP")
	binary.LittleEndian.PutUint32(packed[4:], unpackedSize)
	binary.LittleEndian.PutUint32(packed[8:], 28)
	copy(packed[12:28], key)
	for i := 0; i < unpackedSize; i++ {
		packed[28+i] = unpacked[i] ^ key[i%len(key)]
	}
	// pad to 64 KB
	file := append(packed, make([]byte, 65536-len(packed))...)
	s.file(id, "packed.bin", file)

	p1 := fmt.Sprintf("PKUP size=%d key=%s\nUNPACKED: %d bytes", unpackedSize, key, unpackedSize)
	p2 := "x1 = x0 * 3 + 4\nRESULT(7) = 25"
	s.expected(id, p1, p2)
}

func init() {
	register("101", (*set).gen101)
	register("102", (*set).gen102)
	register("103", (*set).gen103)
	register("105", (*set).gen105)
	register("106", (*set).gen106)
	register("107", (*set).gen107)
	register("108", (*set).gen108)
	register("110", (*set).gen110)
}