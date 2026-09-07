package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
)

// Generators for missions 104, 109, 112, 114, 117: converted from
// live-network/concurrency exercises to deterministic file-based missions
// (see AGENTS.md). All data is deterministic, no unseeded RNG, no wall-clock.

// ---------------------------------------------------------------- 104 mute protocol

func (s *set) gen104() {
	id := "104"
	// fuzz dictionary: "CMD ARGS -> RESPONSE"; "?" means the command is dead.
	dict := "AUTH crow -> OK crow\n" +
		"LOGIN crow -> DENY\n" +
		"PING 1234 -> PONG 1234\n" +
		"ECHO hello -> ?\n" +
		"GET flag -> ERR_NEED_KEY\n" +
		"TIME now -> ?\n" +
		"HELP -> ?\n"
	s.text(id, "protocol.txt", dict)

	var p1 []string
	for _, l := range strings.Split(strings.TrimSpace(dict), "\n") {
		cmd, resp, _ := strings.Cut(l, " -> ")
		if resp == "?" {
			continue
		}
		cmd = strings.Fields(cmd)[0]
		p1 = append(p1, fmt.Sprintf("%-5s %s", cmd+":", resp))
	}

	// handshake: token = first 8 hex chars of sha256("OMEGA:crow")
	h := sha256.Sum256([]byte("OMEGA:crow"))
	token := fmt.Sprintf("%x", h[:4])
	p2 := "AUTH crow -> OK crow\n" +
		fmt.Sprintf("KEY %s -> ACCEPTED\n", token) +
		"GET flag -> FLAG{MUTE_PROTOCOL_BROKEN}"
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 109 VM injection

func (s *set) gen109() {
	id := "109"
	// IronCore (mission 107) program: two forward jumps. JMP/JZ operand is an
	// absolute position from the start of the program (unsigned byte).
	prog := []byte{
		0x08, 0x07, // 0x00 JMP 0x07 (skip dead code)
		0x01, 0x13, 0x37, 0x00, 0x00, // 0x02 PUSH 0x1337 (dead code)
		0x0B, 0x00, 0x00, // 0x07 LOAD 0 (x0)
		0x01, 0x05, 0x00, 0x00, 0x00, // 0x0A PUSH 5
		0x04, // 0x0F ADD
		0x0C, 0x01, 0x00, // 0x10 STORE 1 (x1 = x0+5)
		0x0B, 0x01, 0x00, // 0x13 LOAD 1
		0x01, 0x00, 0x00, 0x00, 0x00, // 0x16 PUSH 0
		0x0A, // 0x1B CMP (x1 == 0)
		0x09, 0x1F, // 0x1C JZ 0x1F (jump if x1 == 0)
		0x0F, // 0x1E HLT
		0x0F, // 0x1F HLT
	}
	s.file(id, "target.bin", prog)

	// part 1: 6-byte needle (PUSH 0; HLT) at offset 0. The original's jumps
	// land in the shifted code, so their operands gain len(needle).
	needle := []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x0F}
	relocs := 0
	for _, b := range prog {
		if b == 0x08 || b == 0x09 {
			relocs++
		}
	}
	p1 := fmt.Sprintf("INJECTED %d bytes @ 0x0000\nJMP FIXED: %d relocations", len(needle), relocs)

	// part 2: trojan payload: for each char of "INFECTED" -> PUSH c;
	// STORE 0xFFFE (screen); then JMP to the shifted original.
	var payload []byte
	for _, c := range "INFECTED" {
		payload = append(payload, 0x01, byte(c), 0x00, 0x00, 0x00, 0x0C, 0xFE, 0xFF)
	}
	payload = append(payload, 0x08, byte(len(payload)+2)) // JMP to the original
	shift := len(payload)

	// relocation: JMP/JZ located inside the shifted original (position >=
	// shift) get their operand += shift.
	full := append(payload, prog...)
	for pc := 0; pc < len(full); pc++ {
		op := full[pc]
		if op != 0x08 && op != 0x09 {
			continue
		}
		if pc >= shift {
			full[pc+1] += byte(shift)
		}
	}

	screen := ironRun(full)
	p2 := "RUN:\n" + screen + "\n(оригинал продолжает работу)\nEXIT: 0"
	s.expected(id, p1, p2)
}

// ironRun executes an IronCore program (mission 107 opcodes). Variables are
// u32 (start 0, x0 = 0); STORE 0xFFFE appends the low byte to the output
// stream (screen); JMP/JZ take an absolute target position (unsigned byte);
// HLT stops with exit code 0.
func ironRun(prog []byte) string {
	var stack []uint32
	var vars [65536]uint32
	var screen []byte
	pc := 0
	for pc < len(prog) {
		op := prog[pc]
		pop := func() uint32 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			return v
		}
		switch op {
		case 0x01:
			stack = append(stack, binary.LittleEndian.Uint32(prog[pc+1:]))
			pc += 5
		case 0x04:
			a := pop()
			b := pop()
			stack = append(stack, a+b)
			pc++
		case 0x08:
			pc = int(prog[pc+1])
		case 0x09:
			a := pop()
			if a == 0 {
				pc = int(prog[pc+1])
			} else {
				pc += 2
			}
		case 0x0A:
			a := pop()
			b := pop()
			if b == a {
				stack = append(stack, 1)
			} else {
				stack = append(stack, 0)
			}
			pc++
		case 0x0B:
			stack = append(stack, vars[binary.LittleEndian.Uint16(prog[pc+1:])])
			pc += 3
		case 0x0C:
			v := pop()
			idx := binary.LittleEndian.Uint16(prog[pc+1:])
			if idx == 0xFFFE {
				screen = append(screen, byte(v))
			} else {
				vars[idx] = v
			}
			pc += 3
		case 0x0F:
			return string(screen)
		default:
			pc++
		}
	}
	return string(screen)
}

// ---------------------------------------------------------------- 112 bank cascade

func (s *set) gen112() {
	id := "112"
	// recorded race outcome: 2000 unlocked transfers A->B, one credit lost.
	race := "initial A=2000 B=0\n" +
		"THREAD A: 1000 transfers A->B (no lock)\n" +
		"THREAD B: 1000 transfers A->B (no lock)\n" +
		"FINAL: A=0 B=1999 total=1999 (expected 2000)\n"
	s.text(id, "race.txt", race)

	// journal: last tx (6) is uncommitted — the crash cuts it off.
	journal := "initial A=1000 B=0 C=0\n" +
		"begin 1\ntx 1 transfer A B 100\ncommit 1\n" +
		"begin 2\ntx 2 transfer B C 50\ncommit 2\n" +
		"begin 3\ntx 3 transfer C A 25\ncommit 3\n" +
		"begin 4\ntx 4 transfer A B 10\ncommit 4\n" +
		"begin 5\ntx 5 transfer B A 5\ncommit 5\n" +
		"begin 6\ntx 6 transfer B C 3\ncrash\n"
	s.text(id, "ops.log", journal)

	// part 1: diagnosis from race.txt
	p1 := "RACE DETECTED: total=1999 (expected 2000)"

	// part 2 reference: apply committed txs (lock order A<B<C), invariant,
	// then crash recovery back to the last committed tx (pending effects of
	// the uncommitted tx are discarded).
	bal := map[string]int{"A": 1000, "B": 0, "C": 0}
	lastCommitted, committed, rolled := 0, 0, 0
	pending := map[string]int{}
	inTx := false
	for _, l := range strings.Split(strings.TrimSpace(journal), "\n") {
		f := strings.Fields(l)
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "begin":
			inTx = true
		case "tx":
			var n, amt int
			var from, to string
			fmt.Sscanf(l, "tx %d transfer %s %s %d", &n, &from, &to, &amt)
			if inTx {
				pending[from] -= amt
				pending[to] += amt
				rolled++
			}
		case "commit":
			var n int
			fmt.Sscanf(f[1], "%d", &n)
			for k, v := range pending {
				bal[k] += v
			}
			pending = map[string]int{}
			lastCommitted = n
			committed = n
			rolled = 0
			inTx = false
		case "crash":
			inTx = false
		}
	}
	p2 := fmt.Sprintf("LOCK ORDER: ok (no deadlock)\n"+
		"INVARIANT: total=%d across %d tx\n"+
		"CRASH RECOVERY: restored to tx #%d (%d rolled back)",
		bal["A"]+bal["B"]+bal["C"], committed, lastCommitted, rolled)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 114 proxy tunnel

func (s *set) gen114() {
	id := "114"
	msg := "HELLO GRID, TUNNEL ECHO 5 BLOCKS OK"
	chunks := []int{7, 7, 7, 7, 7}
	var frames []byte
	off := 0
	for _, c := range chunks {
		frames = append(frames, 0x01)
		var l [2]byte
		binary.BigEndian.PutUint16(l[:], uint16(c))
		frames = append(frames, l[:]...)
		frames = append(frames, msg[off:off+c]...)
		off += c
	}
	frames = append(frames, 0x02) // end of stream
	s.file(id, "frames.bin", frames)

	p1 := "SENT 5 blocks  RECEIVED 5 blocks\nMIRROR OK"
	p2 := msg + "\nTUNNEL OVER HTTP: OK (5 frames)"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 117 botnet

func (s *set) gen117() {
	id := "117"
	nodes := "N=10\n" +
		"node 1 alive\nnode 3 dead\nnode 5 alive\nnode 7 alive\nnode 9 dead\n" +
		"node 11 alive\nnode 13 alive\nnode 15 dead\nnode 17 alive\nnode 19 alive\n"
	s.text(id, "nodes.txt", nodes)

	rounds := "round 1 yes 9\nround 2 yes 7\nround 3 yes 10\n"
	s.text(id, "rounds.txt", rounds)

	// part 1: leader = min alive id, confirmed by alive/N
	total, alive, minAlive := 0, 0, -1
	for _, l := range strings.Split(strings.TrimSpace(nodes), "\n") {
		f := strings.Fields(l)
		if strings.HasPrefix(l, "N=") {
			fmt.Sscanf(l, "N=%d", &total)
			continue
		}
		if f[2] == "alive" {
			var idn int
			fmt.Sscanf(f[1], "%d", &idn)
			alive++
			if minAlive < 0 || idn < minAlive {
				minAlive = idn
			}
		}
	}
	p1 := fmt.Sprintf("LEADER ELECTED: node %d (confirmed by %d/%d)", minAlive, alive, total)

	// part 2: 2PC — COMMIT only if all N vote YES
	var p2 []string
	for _, l := range strings.Split(strings.TrimSpace(rounds), "\n") {
		var i, x int
		fmt.Sscanf(l, "round %d yes %d", &i, &x)
		if x == total {
			p2 = append(p2, fmt.Sprintf("ROUND %d: %d/%d YES -> COMMIT (started)", i, x, total))
		} else {
			p2 = append(p2, fmt.Sprintf("ROUND %d: %d/%d YES -> ABORT", i, x, total))
		}
	}
	s.expected(id, p1, strings.Join(p2, "\n"))
}

func init() {
	register("104", (*set).gen104)
	register("109", (*set).gen109)
	register("112", (*set).gen112)
	register("114", (*set).gen114)
	register("117", (*set).gen117)
}