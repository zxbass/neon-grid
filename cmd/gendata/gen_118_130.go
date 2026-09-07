package main

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"neon-grid/sol"
)

// ---------------------------------------------------------------- 118 mercury

// caesarCipher rotates an uppercase text by n using the same rule as
// mission 022 (A-Z only, spaces pass through).
func caesarCipher(text string, n int) string { return sol.Caesar(text, n) }

func (s *set) gen118() {
	id := "118"
	// riddles.txt: 5 cipher challenges, one per line. Part 1 decrypts
	// each with the stated cipher and prints a ROUND OK line per round.
	xor1 := sol.XorByte([]byte("MERCURY IS WATCHING"), 0x5A)
	caesar2 := sol.Caesar("THE GRID NEVER SLEEPS", 13)
	vig3 := sol.Vigenere("MERCURY SEES ALL", "OMEGA", false, true)
	xor4 := sol.XorByte([]byte("THE GRID NEVER SLEEPS TONIGHT"), 0x4A)
	caesar5 := sol.Caesar("MERCURY SEES ALL", 7)

	riddles := "XOR key=0x5A " + sol.HexStr(xor1) + "\n" +
		"CAESAR n=13 " + caesar2 + "\n" +
		"VIGENERE key=OMEGA " + vig3 + "\n" +
		"XOR key=0x4A " + sol.HexStr(xor4) + "\n" +
		"CAESAR n=7 " + caesar5 + "\n"
	s.text(id, "riddles.txt", riddles)

	var p1 []string
	for i := 1; i <= 5; i++ {
		p1 = append(p1, fmt.Sprintf("ROUND %d/5 OK", i))
	}
	p1 = append(p1, "KEY: mercury_sees_all")

	// guess.txt: one ciphertext + the 3 candidate ciphers. Part 2 scores
	// every candidate decryption and picks the most printable one.
	guess := "ciphertext: " + sol.HexStr(xor4) + "\n" +
		"candidates: XOR CAESAR VIGENERE\n" +
		"vigenere_key: OMEGA\n"
	s.text(id, "guess.txt", guess)

	score := func(b []byte) (float64, bool) {
		good := 0
		space := false
		for _, c := range b {
			if (c >= 'A' && c <= 'Z') || c == ' ' {
				good++
				if c == ' ' {
					space = true
				}
			}
		}
		return float64(good) / float64(len(b)), space
	}

	type cand struct {
		label string
		text  []byte
		sc    float64
		sp    bool
	}
	var best *cand
	consider := func(c *cand) {
		if best == nil || c.sc > best.sc || (c.sc == best.sc && c.sp && !best.sp) {
			best = c
		}
	}
	for k := 0; k <= 0xFF; k++ {
		dec := sol.XorByte(xor4, byte(k))
		sc, sp := score(dec)
		consider(&cand{fmt.Sprintf("XOR (key 0x%02X)", k), dec, sc, sp})
	}
	for n := 1; n <= 25; n++ {
		dec := sol.Caesar(string(xor4), n)
		sc, sp := score([]byte(dec))
		consider(&cand{fmt.Sprintf("CAESAR (shift %d)", n), []byte(dec), sc, sp})
	}
	dec := sol.Vigenere(string(xor4), "OMEGA", true, true)
	sc, sp := score([]byte(dec))
	consider(&cand{"VIGENERE (key OMEGA)", []byte(dec), sc, sp})

	if best.label != "XOR (key 0x4A)" {
		panic("118: expected XOR (key 0x4A) to win the guess, got " + best.label)
	}
	p2 := "CIPHER GUESSED: " + best.label + "\nDIALOGUE COMPLETE: 5/5"
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 120 last exit

func (s *set) gen120() {
	id := "120"
	// path_*.dat: "PATH" magic + XOR key byte + u32le payload length +
	// payload XORed with the key byte (mission 021 XorByte style).
	enc := func(payload string, key byte) []byte {
		var f []byte
		f = append(f, 'P', 'A', 'T', 'H', key)
		var l [4]byte
		binary.LittleEndian.PutUint32(l[:], uint32(len(payload)))
		f = append(f, l[:]...)
		return append(f, sol.XorByte([]byte(payload), key)...)
	}
	s.file(id, "path_obey.dat", enc("encryption_key_for_self", 0x11))
	s.file(id, "path_sell.dat", enc("buyer=HADLEY_FOUNTAIN price=999999", 0x22))
	s.file(id, "path_free.dat", enc("MERCURY_core_v2.049 // he wants out", 0x33))
	s.text(id, "choice.txt", "FREE\n")

	p1 := "OBEY: encryption_key_for_self\n" +
		"SELL: buyer=HADLEY_FOUNTAIN price=999999\n" +
		"FREE: MERCURY_core_v2.049 // he wants out"
	p2 := "CHOICE: FREE\n" +
		"OUTCOME: NEON LIGHTS, OPEN SKY. MERCURY IS FREE.\n" +
		"STATS: 120 missions, 12 themes, binary-to-TUI.\n" +
		"NEXT: the grid remembers you."
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 124 raft

func (s *set) gen124() {
	id := "124"
	// cluster.log: recorded Raft history. Part 1 derives the leader per
	// term from the votes (majority >= 2 of 3); part 2 derives commit acks
	// (leader + LOGOK nodes) and echoes the catch-up.
	log := "VOTE 1 node-1 node-2\n" +
		"VOTE 1 node-3 node-2\n" +
		"LEADER 1 node-2\n" +
		"HEART 1 node-2 idx=4\n" +
		"LOGOK 1 node-1 idx=4\n" +
		"LOGOK 1 node-3 idx=4\n" +
		"COMMIT 1 k=v\n" +
		"VOTE 2 node-2 node-1\n" +
		"VOTE 2 node-3 node-1\n" +
		"LEADER 2 node-1\n" +
		"HEART 2 node-1 idx=7\n" +
		"LOGOK 2 node-2 idx=7\n" +
		"COMMIT 2 k2=v2\n" +
		"CATCHUP 2 node-3 +3 entries\n"
	s.text(id, "cluster.log", log)

	type termInfo struct {
		votes   map[string]int
		logoks  map[string]bool
		commits []string
		catchup []string
	}
	terms := []int{}
	byTerm := map[int]*termInfo{}
	addTerm := func(t int) *termInfo {
		if _, ok := byTerm[t]; !ok {
			terms = append(terms, t)
			byTerm[t] = &termInfo{votes: map[string]int{}, logoks: map[string]bool{}}
		}
		return byTerm[t]
	}
	for _, l := range strings.Split(strings.TrimSpace(log), "\n") {
		f := strings.Fields(l)
		term, _ := strconv.Atoi(f[1])
		ti := addTerm(term)
		switch f[0] {
		case "VOTE":
			ti.votes[f[3]]++
		case "LOGOK":
			ti.logoks[f[2]] = true
		case "COMMIT":
			ti.commits = append(ti.commits, f[2])
		case "CATCHUP":
			ti.catchup = append(ti.catchup, f[2]+" "+f[3]+" "+f[4])
		}
	}
	var p1, p2 []string
	for _, t := range terms {
		ti := byTerm[t]
		leader := ""
		for node, n := range ti.votes {
			if n >= 2 {
				leader = node
				break
			}
		}
		if leader != "" {
			p1 = append(p1, fmt.Sprintf("TERM %d: leader=%s", t, leader))
		}
		for _, c := range ti.commits {
			ack := 1 + len(ti.logoks)
			p2 = append(p2, fmt.Sprintf("COMMIT %s (%d/3)", c, ack))
		}
		for _, c := range ti.catchup {
			p2 = append(p2, "CATCHUP "+c)
		}
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 129 self-heal

// gateCore builds the 64-byte self-healing core: AUTH header, seq,
// magic + integrity crc, the 6-digit access code with its own crc.
func gateCore() []byte {
	core := make([]byte, 64)
	copy(core[0:5], "AUTH ")
	copy(core[5:15], "NEON_RAVEN")
	u32be(core[16:], 0x0A0B0C0D)
	u16le(core[20:], 0x5A5A)
	u16le(core[22:], sol.CRC16CCITT(core[0:22]))
	copy(core[24:30], "704213")
	u16le(core[30:], sol.CRC16CCITT(core[24:30]))
	copy(core[32:], "MERCURY_SELF_HEALING_CORE_012345")
	return core
}

// gateRegion packs the RLE stream + per-block crc16 checksums:
// u16le rleLen || RLE bytes || 4 x u16le crc16(block).
func gateRegion(rle []byte, core []byte) []byte {
	region := make([]byte, 0, 2+len(rle)+8)
	var l [2]byte
	binary.LittleEndian.PutUint16(l[:], uint16(len(rle)))
	region = append(region, l[:]...)
	region = append(region, rle...)
	for i := 0; i < 4; i++ {
		var cr [2]byte
		binary.LittleEndian.PutUint16(cr[:], sol.CRC16CCITT(core[i*16:(i+1)*16]))
		region = append(region, cr[:]...)
	}
	return region
}

func (s *set) gen129() {
	id := "129"
	core := gateCore()
	rle := sol.RLEPack(core)
	if len(rle) != 3*64 {
		panic("129: RLE stream must be all-singleton tuples")
	}
	region := gateRegion(rle, core)

	// gate.bin: u32le version + 256-byte region XORed with 0x5A.
	gate := make([]byte, 4+len(region))
	binary.LittleEndian.PutUint32(gate[0:], 9)
	copy(gate[4:], sol.XorByte(region, 0x5A))
	s.file(id, "gate.bin", gate)

	// gate.damaged: same image with 3 bytes of block 2 (core bytes
	// 17, 23, 27) flipped inside the RLE stream, in place.
	dr := append([]byte{}, region...)
	for _, i := range []int{17, 23, 27} {
		dr[4+3*i] ^= 0xA5
	}
	damaged := make([]byte, 4+len(dr))
	binary.LittleEndian.PutUint32(damaged[0:], 9)
	copy(damaged[4:], sol.XorByte(dr, 0x5A))
	s.file(id, "gate.damaged", damaged)

	// gate.mirror: "MIRR" + u16le block count + 4 plain blocks + the
	// healthy rleLen + healthy RLE stream (for reference).
	mirror := []byte{'M', 'I', 'R', 'R'}
	var nb [2]byte
	binary.LittleEndian.PutUint16(nb[:], 4)
	mirror = append(mirror, nb[:]...)
	mirror = append(mirror, core...)
	var l [2]byte
	binary.LittleEndian.PutUint16(l[:], uint16(len(rle)))
	mirror = append(mirror, l[:]...)
	mirror = append(mirror, rle...)
	s.file(id, "gate.mirror", mirror)

	// sanity: only block 2 of the damaged image fails its crc.
	badUnpack, _ := sol.RLEUnpack(dr[2 : 2+int(binary.LittleEndian.Uint16(dr[0:2]))])
	crcsOff := 2 + int(binary.LittleEndian.Uint16(dr[0:2]))
	failures := 0
	for i := 0; i < 4; i++ {
		if sol.CRC16CCITT(badUnpack[i*16:(i+1)*16]) != binary.LittleEndian.Uint16(dr[crcsOff+2*i:]) {
			failures++
		}
	}
	if failures != 1 {
		panic(fmt.Sprintf("129: damaged image must fail exactly 1 block crc, got %d", failures))
	}

	code := string(core[24:30])
	p1 := fmt.Sprintf("STAGE 1: OK (version=9)\n"+
		"STAGE 2: OK (len=256)\n"+
		"STAGE 3: OK (blocks=4)\n"+
		"STAGE 4: OK (auth=NEON_RAVEN)\n"+
		"STAGE 5: OK (magic=0x5A5A crc16=0x%04X)\n"+
		"STAGE 6: OK (code_crc=0x%04X)\n"+
		"STAGE 7: CODE=%s ACCESS GRANTED",
		binary.LittleEndian.Uint16(core[22:24]),
		binary.LittleEndian.Uint16(core[30:32]), code)
	blockOff := 6 + 16*(2-1)
	p2 := fmt.Sprintf("DAMAGE: stage 3 block #2\n"+
		"HEALED: using mirror copy (offset 0x%02X)\n"+
		"RESULT: ACCESS GRANTED (self-healed 1/1)", blockOff)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 130 voice

func (s *set) gen130() {
	id := "130"
	challenge := "0e63b9c7a41f4d2d"
	tag := sol.CRC16CCITT([]byte(challenge))
	s.text(id, "handshake.txt",
		"nonce=deadbeef00c0ffee\n"+
			"challenge="+challenge+"\n"+
			fmt.Sprintf("tag=%04x\n", tag))

	p1 := fmt.Sprintf("CHALLENGE: %s\nTAG: %04x", challenge, tag)

	type msg struct{ you, merc string }
	dialog := []msg{
		{"who am i?", "thorn"},
		{"hello thorn", "you made it through the grid"},
		{"what now?", "welcome back, crow. the grid remembers."},
	}
	key64, _ := strconv.ParseUint(challenge[:4], 16, 16)
	key := uint16(key64)
	var lines, out []string
	for _, m := range dialog {
		kb := []byte{byte(key >> 8), byte(key)}
		lines = append(lines, "YOU "+sol.HexStr(sol.XorRepeat([]byte(m.you), kb)))
		lines = append(lines, "MERC "+sol.HexStr(sol.XorRepeat([]byte(m.merc), kb)))
		out = append(out, "YOU: "+m.you)
		out = append(out, `MERC: "`+m.merc+`"`)
		key = sol.CRC16CCITT(append(kb, []byte(m.merc)...))
	}
	s.text(id, "dialog.txt", strings.Join(lines, "\n")+"\n")
	p2 := strings.Join(out, "\n")
	s.expected(id, p1, p2)
}

func init() {
	register("118", (*set).gen118)
	register("120", (*set).gen120)
	register("124", (*set).gen124)
	register("129", (*set).gen129)
	register("130", (*set).gen130)
}