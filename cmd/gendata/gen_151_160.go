package main

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// ---------------------------------------------------------------- 151 /proc

func (s *set) gen151() {
	id := "151"
	stat := "2049 (zen_agent) R 1 2049 2049 0 -1 4194304 128 0 0 0 12 34 5 7 20 19 5 0 0 0 12345678 1024 18446744073709551615"
	status := "Name:\tzen_agent\nState:\tR (running)\nPid:\t2049\nPPid:\t1\nThreads:\t5\nVmSize:\t12345678 kB\nVmRSS:\t1024 kB"
	s.text(id, "stat.txt", stat)
	s.text(id, "status.txt", status)
	p1 := "PID: 2049 NAME: zen_agent STATE: R THREADS: 5"
	p2 := "CPU TICKS: 46\nVSIZE: 12345678 RSS: 1024"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 152 ELF

func (s *set) gen152() {
	id := "152"
	textData := make([]byte, 16)
	for i := range textData {
		textData[i] = byte(i + 1)
	}
	shstrtab := []byte{0, '.', 't', 'e', 'x', 't', 0, '.', 's', 'h', 's', 't', 'r', 't', 'a', 'b', 0}

	ehdr := make([]byte, 64)
	copy(ehdr[0:4], "\x7fELF")
	ehdr[4] = 2 // 64-bit
	ehdr[5] = 1 // little endian
	ehdr[6] = 1
	ehdr[7] = 0
	u16le(ehdr[16:], 2) // EXEC
	u16le(ehdr[18:], 0x3E)
	u32le(ehdr[20:], 1)
	u32le(ehdr[24:], 0x401000) // entry
	u32le(ehdr[32:], 64)       // phoff
	u32le(ehdr[40:], 64+56)    // shoff
	u16le(ehdr[52:], 64)
	u16le(ehdr[54:], 56)
	u16le(ehdr[56:], 1)
	u16le(ehdr[58:], 64)
	u16le(ehdr[60:], 3)
	u16le(ehdr[62:], 2)

	phdr := make([]byte, 56)
	u32le(phdr[0:], 1)    // LOAD
	u32le(phdr[4:], 5)    // R+X
	u32le(phdr[8:], 0)    // offset
	u32le(phdr[16:], 0x400000)
	u32le(phdr[24:], 0x400000)
	u32le(phdr[32:], 0x1000)
	u32le(phdr[40:], 0x1000)
	u32le(phdr[48:], 0x1000)

	var elf []byte
	elf = append(elf, ehdr...)
	elf = append(elf, phdr...)
	shoff := len(elf)
	sh0 := make([]byte, 64)
	sh1 := make([]byte, 64)
	binary.LittleEndian.PutUint32(sh1[0:], 1) // sh_name = ".text"
	binary.LittleEndian.PutUint32(sh1[4:], 1) // sh_type = PROGBITS
	binary.LittleEndian.PutUint64(sh1[8:], 6) // sh_flags = SHF_ALLOC|SHF_EXECINSTR
	binary.LittleEndian.PutUint64(sh1[16:], 0x401000)
	binary.LittleEndian.PutUint64(sh1[24:], uint64(shoff+64*3))               // sh_offset (after 3 section headers)
	binary.LittleEndian.PutUint64(sh1[32:], uint64(len(textData)))
	binary.LittleEndian.PutUint64(sh1[48:], 16) // sh_addralign
	sh2 := make([]byte, 64)
	binary.LittleEndian.PutUint32(sh2[0:], 7) // sh_name = ".shstrtab"
	binary.LittleEndian.PutUint32(sh2[4:], 3) // sh_type = STRTAB
	binary.LittleEndian.PutUint64(sh2[24:], uint64(shoff+64*3+len(textData)))
	binary.LittleEndian.PutUint64(sh2[32:], uint64(len(shstrtab)))
	binary.LittleEndian.PutUint64(sh2[48:], 1) // sh_addralign
	elf = append(elf, sh0...)
	elf = append(elf, sh1...)
	elf = append(elf, sh2...)
	elf = append(elf, textData...)
	elf = append(elf, shstrtab...)
	s.file(id, "prog.elf", elf)

	p1 := "CLASS: ELF64 TYPE: EXEC MACHINE: x86-64 ENTRY: 0x401000"
	p2 := fmt.Sprintf("SECTIONS: 2 (.text, .shstrtab)\nTEXT SIZE: %d", len(textData))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 153 PLT/GOT

func (s *set) gen153() {
	id := "153"
	// GOT содержит либо адрес PLT-стаба (неразрешённый символ), либо адрес функции.
	// Игрок: если GOT == STUB — применить ленивую привязку и подставить адрес символа.
	got := map[string]uint32{"read": 0x404000, "write": 0x404008, "exit": 0x404010}
	plts := map[string]uint32{"read": 0x401050, "write": 0x401060, "exit": 0x401070}
	sym := map[string]uint32{"read": 0x7F0001, "write": 0x7F0012, "exit": 0x7F0000}
	var lines []string
	for _, name := range []string{"read", "write", "exit"} {
		// read/write ещё не привязаны (VALUE = PLT-стаб), exit уже привязан (VALUE = SYM)
		val := plts[name]
		if name == "exit" {
			val = sym[name]
		}
		lines = append(lines, fmt.Sprintf("%s SLOT=0x%08X VALUE=0x%08X STUB=0x%08X SYM=0x%08X",
			name, got[name], val, plts[name], sym[name]))
	}
	s.text(id, "got.txt", strings.Join(lines, "\n"))
	var p1, p2 []string
	for _, name := range []string{"read", "write", "exit"} {
		resolved := sym[name]
		p1 = append(p1, fmt.Sprintf("%s -> 0x%08X", name, resolved))
		p2 = append(p2, fmt.Sprintf("GOT[%s]=0x%08X", name, resolved))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 154 stack canary

func (s *set) gen154() {
	id := "154"
	s.text(id, "layout.txt", "BUFFER: 16\nCANARY: 8\nRET_OFFSET: 24\nSAVED_RBP: 16")
	canary := uint64(0xDEADBEEFCAFEBABE)
	ret := uint32(0x401337)
	payload := append([]byte{}, bytes16(0x41)...)
	var c [8]byte
	binary.LittleEndian.PutUint64(c[:], canary)
	payload = append(payload, c[:]...)
	var r [4]byte
	binary.LittleEndian.PutUint32(r[:], ret)
	payload = append(payload, r[:]...)
	s.text(id, "stack.txt", fmt.Sprintf("CANARY_AT: 16\nRET_AT: 24\nRBP_AT: 16"))

	p1 := fmt.Sprintf("CANARY: 0x%016X", canary)
	p2 := fmt.Sprintf("PAYLOAD: %X\nRIP: 0x%08X", payload, ret)
	s.expected(id, p1, p2)
}

func bytes16(b byte) []byte {
	out := make([]byte, 16)
	for i := range out {
		out[i] = b
	}
	return out
}

// ---------------------------------------------------------------- 155 format string

func (s *set) gen155() {
	id := "155"
	// stack: 8 words; word 6 (offset 7) points to secret string
	secret := "NEON_FMT_LEAKED"
	stack := []uint32{
		0x41424344, 0x20494545, 0x5A202020, 0x31313234,
		0x41414141, 0x42424242, 0xDEADBEEF, 0x00000000,
	}
	// leak line: %1$x %2$x ... %6$x then %7$s
	leak := make([]string, 6)
	for i := 0; i < 6; i++ {
		leak[i] = fmt.Sprintf("0x%08X", stack[i])
	}
	var lines []string
	for i, v := range stack {
		lines = append(lines, fmt.Sprintf("%d: 0x%08X", i+1, v))
	}
	s.text(id, "stack.txt", strings.Join(lines, "\n"))
	s.text(id, "secret.txt", secret)
	p1 := "LEAK: " + strings.Join(leak, " ")
	p2 := fmt.Sprintf("SECRET[7]: %s", secret)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 156 debugger

func cpuStepTrace(prog []byte) (steps []string, out []byte, finalA byte) {
	pc := 0
	var a byte
	mem := make([]byte, 256)
	opWidth := map[byte]int{0x00: 1, 0x01: 2, 0x02: 2, 0x03: 2, 0x04: 2, 0x05: 2, 0x06: 2, 0x07: 1, 0x08: 2}
	names := map[byte]string{0x00: "HLT", 0x01: "LDA", 0x02: "ADD", 0x03: "SUB", 0x04: "STA", 0x05: "JMP", 0x06: "JZ", 0x07: "PRN", 0x08: "INP"}
	for step := 0; pc < len(prog) && step < 10000; step++ {
		op := prog[pc]
		w := opWidth[op]
		if w == 0 || pc+w > len(prog) {
			break
		}
		var opd byte
		if w == 2 {
			opd = prog[pc+1]
		}
		before := a
		switch op {
		case 0x00:
			steps = append(steps, fmt.Sprintf("STEP %d: A=0x%02X PC=%d OP=HLT", step, a, pc))
			return steps, out, a
		case 0x01:
			a = opd
		case 0x02:
			a += opd
		case 0x03:
			a -= opd
		case 0x04:
			mem[opd] = a
		case 0x05:
			steps = append(steps, fmt.Sprintf("STEP %d: A=0x%02X PC=%d OP=JMP(%d)", step, before, pc, opd))
			pc = int(opd)
			continue
		case 0x06:
			steps = append(steps, fmt.Sprintf("STEP %d: A=0x%02X PC=%d OP=JZ(%d)", step, before, pc, opd))
			if a == 0 {
				pc = int(opd)
			} else {
				pc += w
			}
			continue
		case 0x07:
			out = append(out, a)
		case 0x08:
			a = mem[opd]
		}
		steps = append(steps, fmt.Sprintf("STEP %d: A=0x%02X PC=%d OP=%s(%d)", step, before, pc, names[op], opd))
		pc += w
	}
	return steps, out, a
}

func (s *set) gen156() {
	id := "156"
	// LDA #10; ADD #5; PRN; ADD #1; PRN; ADD #1; SUB #3; JZ $hlt; NOP; hlt
	prog := []byte{0x01, 0x0A, 0x02, 0x05, 0x07, 0x02, 0x01, 0x07, 0x02, 0x01, 0x03, 0x03, 0x06, 0x0E, 0x00}
	s.file(id, "prog.bin", prog)
	steps, out, a := cpuStepTrace(prog)
	p1 := strings.Join(steps, "\n")
	p2 := fmt.Sprintf("OUTPUT: %X\nFINAL A: 0x%02X", out, a)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 157 syscall tracer

func (s *set) gen157() {
	id := "157"
	trace := []string{
		"read(3, 0x7fff0010, 128) = 128",
		"write(1, 0x7fff0010, 128) = 128",
		"open(\"/etc/passwd\", 0) = 4",
		"read(4, 0x7fff0110, 4096) = 2048",
		"close(4) = 0",
		"mmap(0, 4096, 3) = 0x7f0000",
		"write(1, 0x7fff0120, 64) = -1 EAGAIN",
		"nanosleep(0x7fff0200) = 0",
		"read(3, 0x7fff0130, 4096) = 0",
		"exit_group(0) = ?",
	}
	s.text(id, "trace.txt", strings.Join(trace, "\n"))
	// "read(3, 0x7fff0010, 128) = 128" — syscall name before '('
	counts := map[string]int{}
	fails := 0
	for _, line := range trace {
		name := line[:strings.Index(line, "(")]
		counts[name]++
		if strings.Contains(line, "= -1") {
			fails++
		}
	}
	var names []string
	for n := range counts {
		names = append(names, n)
	}
	sort.Strings(names)
	var p1 []string
	for _, n := range names {
		p1 = append(p1, fmt.Sprintf("%s x%d", n, counts[n]))
	}
	p2 := fmt.Sprintf("TOTAL: %d FAILED: %d", len(trace), fails)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 158 zombies

func (s *set) gen158() {
	id := "158"
	procs := []struct {
		pid, ppid int
		state     string
	}{
		{1, 0, "S"}, {100, 1, "S"}, {101, 100, "Z"}, {102, 100, "R"},
		{200, 1, "S"}, {201, 200, "Z"}, {300, 999, "S"}, {301, 300, "S"},
	}
	var lines []string
	for _, p := range procs {
		lines = append(lines, fmt.Sprintf("%d %d %s", p.pid, p.ppid, p.state))
	}
	s.text(id, "procs.txt", strings.Join(lines, "\n"))
	inSet := map[int]bool{}
	for _, p := range procs {
		inSet[p.pid] = true
	}
	var zombies, orphans []string
	for _, p := range procs {
		if p.state == "Z" {
			zombies = append(zombies, fmt.Sprint(p.pid))
		}
		if p.ppid != 0 && !inSet[p.ppid] {
			orphans = append(orphans, fmt.Sprint(p.pid))
		}
	}
	p1 := fmt.Sprintf("ZOMBIES: %d [%s]\nORPHANS: %d [%s]", len(zombies), strings.Join(zombies, " "), len(orphans), strings.Join(orphans, " "))
	p2 := fmt.Sprintf("REPARENTED: %d (adopted by PID 1)", len(orphans))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 159 buddy allocator

type buddySim struct {
	pool     int
	free     map[int][]int // order -> list of block start addresses
	maxOrder int
}

func newBuddy(pool int) *buddySim {
	o := 0
	for (1 << o) < pool {
		o++
	}
	b := &buddySim{pool: pool, free: map[int][]int{}, maxOrder: o}
	b.free[o] = []int{0}
	return b
}

func (b *buddySim) alloc(size int) (int, bool) {
	order := 0
	for (1 << order) < size {
		order++
	}
	// find smallest available order >= requested
	found := -1
	for o := order; o <= b.maxOrder; o++ {
		if len(b.free[o]) > 0 {
			found = o
			break
		}
	}
	if found == -1 {
		return 0, false
	}
	addr := b.free[found][0]
	b.free[found] = b.free[found][1:]
	// split down
	for o := found; o > order; o-- {
		half := 1 << (o - 1)
		b.free[o-1] = append(b.free[o-1], addr+half)
	}
	return addr, true
}

func (b *buddySim) freeBlock(addr, size int) {
	order := 0
	for (1 << order) < size {
		order++
	}
	for {
		b.free[order] = append(b.free[order], addr)
		// buddy merge
		buddy := addr ^ (1 << order)
		found := -1
		for i, a := range b.free[order] {
			if a == buddy {
				found = i
				break
			}
		}
		if found == -1 || order == b.maxOrder {
			return
		}
		b.free[order] = append(b.free[order][:found], b.free[order][found+1:]...)
		b.free[order] = b.free[order][:len(b.free[order])-1]
		addr = addr &^ (1 << order)
		order++
	}
}

func (s *set) gen159() {
	id := "159"
	ops := []string{
		"ALLOC 300",
		"ALLOC 512",
		"ALLOC 64",
		"FREE 1",
		"ALLOC 1000",
		"ALLOC 128",
		"FREE 2",
	}
	s.text(id, "ops.txt", strings.Join(ops, "\n"))
	b := newBuddy(1024)
	type alloc struct {
		addr, size int
	}
	var allocs []alloc
	ok, fail := 0, 0
	for _, op := range ops {
		parts := strings.Fields(op)
		if parts[0] == "ALLOC" {
			var size int
			fmt.Sscanf(parts[1], "%d", &size)
			addr, good := b.alloc(size)
			if good {
				allocs = append(allocs, alloc{addr, size})
				ok++
			} else {
				fail++
			}
		} else {
			var idx int
			fmt.Sscanf(parts[1], "%d", &idx)
			b.freeBlock(allocs[idx].addr, allocs[idx].size)
		}
	}
	// free list summary
	var sizes []string
	for o := 0; o <= b.maxOrder; o++ {
		for _, a := range b.free[o] {
			sizes = append(sizes, fmt.Sprintf("%d@%d", 1<<o, a))
		}
	}
	sort.Strings(sizes)
	largest := 0
	for o := b.maxOrder; o >= 0; o-- {
		if len(b.free[o]) > 0 {
			largest = 1 << o
			break
		}
	}
	p2 := []string{
		fmt.Sprintf("ALLOC OK: %d FAIL: %d", ok, fail),
		fmt.Sprintf("LARGEST FREE: %d", largest),
	}
	s.expected(id, "FREE BLOCKS: "+strings.Join(sizes, " "), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 160 mmap LRU cache

func (s *set) gen160() {
	id := "160"
	pages := []int{0x0A, 0x1B, 0x0A, 0x2C, 0x3D, 0x1B, 0x0A, 0x4E, 0x2C, 0x5F, 0x0A, 0x1B}
	var lines []string
	for _, p := range pages {
		lines = append(lines, fmt.Sprintf("%02X", p))
	}
	s.text(id, "trace.txt", strings.Join(lines, "\n"))
	const cap4 = 4
	var cache []int
	hits, misses := 0, 0
	for _, p := range pages {
		found := -1
		for i, c := range cache {
			if c == p {
				found = i
				break
			}
		}
		if found >= 0 {
			hits++
			cache = append(cache[:found], cache[found+1:]...)
			cache = append(cache, p)
		} else {
			misses++
			cache = append(cache, p)
			if len(cache) > cap4 {
				cache = cache[1:]
			}
		}
	}
	var out []string
	for _, c := range cache {
		out = append(out, fmt.Sprintf("0x%02X", c))
	}
	p1 := fmt.Sprintf("HITS: %d MISSES: %d", hits, misses)
	p2 := "FINAL CACHE (LRU->MRU): " + strings.Join(out, " ")
	s.expected(id, p1, p2)
}

func init() {
	register("151", (*set).gen151)
	register("152", (*set).gen152)
	register("153", (*set).gen153)
	register("154", (*set).gen154)
	register("155", (*set).gen155)
	register("156", (*set).gen156)
	register("157", (*set).gen157)
	register("158", (*set).gen158)
	register("159", (*set).gen159)
	register("160", (*set).gen160)
}
