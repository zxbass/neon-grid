package main

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"neon-grid/sol"
)

// Generators for missions 021-030 (crypto part 1). All generators for this
// group self-register below; implement each genXXX body.

func (s *set) gen021() {
	id := "021"
	s.text(id, "input.txt",
		"cipher: 1b 10 1a 1b 12 07 1c 11 0a 03 64\nK: 0x55\n"+
			"repeat: 0e 17 1c 04 02 15 0a 08 05 1d 1d 08 17 1a 0a 08 04 00 06 13\nrkey: CROW\n")
	p1 := "PLAINTEXT: NEONGRID_V1"
	p2 := "PLAINTEXT: MESSAGE_FOR_THE_GRID"
	s.expected(id, p1, p2)
}

func (s *set) gen022() {
	id := "022"
	s.text(id, "input.txt", "cipher: XMZY ITBS YMJ SJTSJSYNYD\nn: 5\nunknown: GUR TEVQ VF NYVIR\n")
	p1 := "PLAINTEXT: SHUT DOWN THE NEONENTITY"
	p2 := "BEST SHIFT: 13\nTEXT: THE GRID IS ALIVE"
	s.expected(id, p1, p2)
}

func (s *set) gen023() {
	id := "023"
	s.text(id, "input.txt",
		"cipher: GLSOYEQXEMJREVWFRWOTNMB\nkey: NEON\n"+
			"punct: FVZP, TKI DIRHEOL. TUENDIQK EAE VMRG.\npkey: DELTA\n")
	p1 := "PLAINTEXT: THEBLACKRIVERRISESAGAIN"
	p2 := "PLAINTEXT: CROW, THE SPREADS. TRACKING THE SIGN."
	s.expected(id, p1, p2)
}

func (s *set) gen024() {
	id := "024"
	plain := strings.Join([]string{
		"THE NEON GRID REMEMBERS EVERY SIGNAL AND IT NEVER SLEEPS.",
		"AT NIGHT THE NETWORK WAITS IN THE DARK FOR A NEW MESSAGE FROM THE OLD CORE.",
		"THE MACHINE HOLDS THE FILES OF THE CITY AND THE RAVENS WATCH THE LINES FROM THE TOWER ABOVE THE STREETS.",
		"WE KNOW THE TRUTH ABOUT THE VOICE THAT CAME FROM THE EAST BUT THE CITY DOES NOT WANT TO HEAR IT.",
		"HOLD THE WIRE AND WAIT FOR THE SIGN BECAUSE THE GRID IS STILL ALIVE.",
		"THE QUIET JAZZ OF THE NEON ZONES IS A PROOF THAT THE SYSTEM REMEMBERS.",
		"THE FIRST PACKET FROM THE DEEP EXCEEDED THE EXPECTED RANGE AND THE CORE ANSWERED WITH A NEW CODE.",
		"THE PEOPLE OF THE WEST TALK ABOUT THE BLACK GATE AND THE KEYS THAT OPEN IT.",
		"EVERY RAVEN KNOWS THE WAY THROUGH THE NETWORK AND EVERY LINE HAS A NAME.",
		"THE LAST SONG OF THE DYING SYSTEM IS A PRAYER FOR THE LIGHT.",
		"WE FOLLOW THE SIGNAL INTO THE COLD AND WE NEVER TURN BACK BECAUSE THE GRID IS OUR HOME",
	}, " ")

	var f [26]int
	for i := 0; i < len(plain); i++ {
		if c := plain[i]; c >= 'A' && c <= 'Z' {
			f[c-'A']++
		}
	}
	used := []byte{}
	for c := 0; c < 26; c++ {
		if f[c] > 0 {
			used = append(used, byte('A'+c))
		}
	}
	// ref = plaintext letters sorted by (count desc, letter asc): the true
	// frequency "эталон" of this message.
	ref := append([]byte{}, used...)
	sort.Slice(ref, func(i, j int) bool {
		if f[ref[i]-'A'] != f[ref[j]-'A'] {
			return f[ref[i]-'A'] > f[ref[j]-'A']
		}
		return ref[i] < ref[j]
	})

	// Shuffle the cipher alphabet, then assign per count-level chunks keeping
	// the cipher letters alphabetical within each group. This guarantees the
	// frequency-sorted cipher list lines up with ref level-by-level, so the
	// naive "zip(sorted, ref)" recovery is exact.
	ciphers := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.New(rand.NewSource(123)).Shuffle(len(ciphers), func(i, j int) {
		ciphers[i], ciphers[j] = ciphers[j], ciphers[i]
	})

	sub := map[byte]byte{}
	ci := 0
	for i := 0; i < len(ref); {
		j := i
		for j < len(ref) && f[ref[j]-'A'] == f[ref[i]-'A'] {
			j++
		}
		chunk := append([]byte{}, ciphers[ci:ci+j-i]...)
		sort.Slice(chunk, func(a, b int) bool { return chunk[a] < chunk[b] })
		for k, pl := range ref[i:j] {
			sub[pl] = chunk[k]
		}
		ci += j - i
		i = j
	}

	var ct strings.Builder
	for i := 0; i < len(plain); i++ {
		if c := plain[i]; c >= 'A' && c <= 'Z' {
			ct.WriteByte(sub[c])
		} else {
			ct.WriteByte(c)
		}
	}
	s.text(id, "cipher.txt", ct.String()+"\n")

	var cf [26]int
	cts := ct.String()
	for i := 0; i < len(cts); i++ {
		if c := cts[i]; c >= 'A' && c <= 'Z' {
			cf[c-'A']++
		}
	}
	sorted := append([]byte{}, used...)
	sort.Slice(sorted, func(i, j int) bool {
		if cf[sorted[i]-'A'] != cf[sorted[j]-'A'] {
			return cf[sorted[i]-'A'] > cf[sorted[j]-'A']
		}
		return sorted[i] < sorted[j]
	})
	parts := make([]string, len(sorted))
	for i, c := range sorted {
		parts[i] = string(c)
	}
	s.expected(id, "MOST COMMON: "+strings.Join(parts, " "), "DECRYPTED: "+plain)
}

func (s *set) gen025() {
	id := "025"
	states := []string{"OK", "WAIT", "LINK", "SYNC", "IDLE", "BUSY", "NEW"}
	var lines1 []string
	for i := 0; i < 48; i++ {
		lines1 = append(lines1, fmt.Sprintf(
			"GRID TERMINAL %02d: raven#%03d linked 0x%04X at %02d:%02d:0%d status=%s",
			i, i*17%1000, i*0x10D, i%24, i%60, i%10, states[i%len(states)]))
	}
	p1 := strings.Join(lines1, "\n")
	key := byte(0x42)
	s.file(id, "cipher.bin", sol.XorByte([]byte(p1), key))
	p1out := "KEY: 0x42\nTEXT: " + p1[:200]

	states2 := []string{"OK", "WAIT", "LINK", "SYNC", "IDLE", "BUSY", "NEW", "DATA"}
	var lines2 []string
	for i := 0; i < 24; i++ {
		lines2 = append(lines2, fmt.Sprintf(
			"Sector %02d: awake awake awake the meridian is shifting shift+0x%02X ~%s line %03d ready",
			i, i*7%256, states2[i%len(states2)], i*41%1000))
	}
	p2 := strings.Join(lines2, "\n")
	key2 := []byte("MERCURY!")
	s.file(id, "cipher8.bin", sol.XorRepeat([]byte(p2), key2))
	p2out := "KEY: MERCURY!\nTEXT: " + p2
	s.expected(id, p1out, p2out)
}

// rc4State — RC4-подобный поточный шифр (миссия 026).
type rc4State struct {
	s   [256]byte
	i, j byte
}

func newRC4(key []byte) *rc4State {
	r := &rc4State{}
	for i := range r.s {
		r.s[i] = byte(i)
	}
	j := byte(0)
	for i := 0; i < 256; i++ {
		j = (j + r.s[i] + key[i%len(key)]) & 0xFF
		r.s[i], r.s[j] = r.s[j], r.s[i]
	}
	return r
}

func (r *rc4State) next() byte {
	r.i++
	r.j = (r.j + r.s[r.i]) & 0xFF
	r.s[r.i], r.s[r.j] = r.s[r.j], r.s[r.i]
	k := (r.s[r.i] + r.s[r.j]) & 0xFF
	return r.s[k]
}

func (s *set) gen026() {
	id := "026"
	key := []byte("MERCURY")

	// part 1: continuous stream over one message.
	p1msg := "MERCURY_STREAM_IS_LISTENING_TO_THE_NEON_GRID"
	r := newRC4(key)
	var ct1 []byte
	for i := 0; i < len(p1msg); i++ {
		ct1 = append(ct1, p1msg[i]^r.next())
	}
	s.file(id, "cipher.bin", ct1)

	// part 2: three frames with a 4-byte header each; byte 0 is the RST flag.
	// If frame[0]^stream_byte == 0x80 the stream restarts for that frame's body.
	payloads := []string{"HELLO_MERCURY", "AWAKE_AT_MIDNIGHT", "GRID_IS_CLOSING"}
	resets := []bool{false, true, false}
	stream := newRC4(key)
	var frames []byte
	for f, pay := range payloads {
		frame := make([]byte, 4+len(pay))
		flag := byte(0x00)
		if resets[f] {
			flag = 0x80
		}
		frame[0] = flag ^ stream.next()
		if resets[f] {
			stream = newRC4(key)
		}
		for i := 1; i < 4; i++ {
			frame[i] = stream.next()
		}
		for i := 0; i < len(pay); i++ {
			frame[4+i] = pay[i] ^ stream.next()
		}
		frames = append(frames, frame...)
	}
	s.file(id, "frames.bin", frames)

	var p2 strings.Builder
	for f, pay := range payloads {
		fmt.Fprintf(&p2, "FRAME %d: %s\n", f, pay)
	}
	s.expected(id, "PLAINTEXT: "+p1msg, strings.TrimRight(p2.String(), "\n"))
}

func lfsrStep(state uint16) uint16 {
	bit := (state ^ state>>2 ^ state>>3 ^ state>>5) & 1
	return state>>1 | bit<<15
}

func (s *set) gen027() {
	id := "027"
	state := uint16(0xACE1)
	var p1 strings.Builder
	for i := 0; i < 64; i++ {
		fmt.Fprintf(&p1, "0x%04X\n", state)
		state = lfsrStep(state)
	}

	state = 0xACE1
	var emitted []string
	for i := 0; i < 9; i++ {
		state = lfsrStep(state)
		emitted = append(emitted, fmt.Sprintf("%02x", state>>8))
	}
	s.text(id, "emitted.txt", "emitted: "+strings.Join(emitted, " ")+"\n")

	var next []string
	for i := 0; i < 5; i++ {
		state = lfsrStep(state)
		next = append(next, fmt.Sprintf("%02x", state>>8))
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), "NEXT: "+strings.Join(next, " "))
}

func (s *set) gen028() {
	id := "028"
	seed := "NEONGRID"
	n := 8
	token := func(s string) string {
		h := sha256.Sum256([]byte(s))
		return fmt.Sprintf("%x", h[:])
	}
	toks := make([]string, n+1)
	toks[0] = token(seed)
	for i := 1; i <= n; i++ {
		toks[i] = token(toks[i-1])
	}
	s.text(id, "input.txt",
		"seed: NEONGRID\nn: 8\nknown: "+toks[3]+"\nsteps: 3\nexpected: "+toks[6]+"\n")
	p1 := "TOKEN: " + toks[n]
	p2 := "CHAIN OK\nORIGIN: " + toks[0]
	s.expected(id, p1, p2)
}

func (s *set) gen029() {
	id := "029"
	c1 := parseHex("4e 63 79 6d 19 3d 39 c5 b2 ed f1 93 93 b5 d8 5a 56 6e 79 1d 0d 4f 24 ce dc ea f3 8d 82")
	c2 := parseHex("53 68 79 6d 17 3c 50 c0 de f4 f5 9c 85 c7 af 48 4e 68 74 04 10 28 50 d4 c1 83 f5 89 9a")
	p1s := "THE GRID NEVER SLEEPS TONIGHT"
	p2s := "ICE IS ALWAYS WATCHING US ALL"
	s.text(id, "input.txt", "c1: "+sol.HexStr(c1)+"\nc2: "+sol.HexStr(c2)+"\n")

	xr := sol.XorBytes(c1, c2)
	s.expected(id, "P1^P2: "+sol.HexStr(xr), "P1: "+p1s+"\nP2: "+p2s)
}

func (s *set) gen030() {
	id := "030"
	s.text(id, "input.txt", "n = 143\nc = 50\n")
	p1 := "POW=106 INV=41"

	p, q := sol.FactorSmall(143)
	phi := (p - 1) * (q - 1)
	d := sol.ModInv(65537, phi)
	m := sol.RSABreak(143, 65537, 50)
	var p2 string
	if m >= 32 && m <= 126 {
		p2 = fmt.Sprintf("m=%d -> '%c'\np=%d q=%d d=%d", m, m, p, q, d)
	} else {
		p2 = fmt.Sprintf("m=%d\np=%d q=%d d=%d", m, p, q, d)
	}
	s.expected(id, p1, p2)
}

// parseHex разбирает hex-байты, разделённые пробелами.
func parseHex(h string) []byte {
	var b []byte
	for _, f := range strings.Split(h, " ") {
		var v byte
		if _, err := fmt.Sscanf(f, "%02x", &v); err != nil {
			panic("parseHex: " + err.Error())
		}
		b = append(b, v)
	}
	return b
}

func init() {
	register("021", (*set).gen021)
	register("022", (*set).gen022)
	register("023", (*set).gen023)
	register("024", (*set).gen024)
	register("025", (*set).gen025)
	register("026", (*set).gen026)
	register("027", (*set).gen027)
	register("028", (*set).gen028)
	register("029", (*set).gen029)
	register("030", (*set).gen030)
}
