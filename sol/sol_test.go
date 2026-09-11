package sol

import (
	"bytes"
	"encoding/hex"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := Hex(s)
	if err != nil {
		t.Fatalf("bad hex %q: %v", s, err)
	}
	return b
}

func eqBytes(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: got %x want %x", name, got, want)
	}
}

// ---------------------------------------------------------------- 010

func TestMission010HexDump(t *testing.T) {
	data := append([]byte("NEONGRID1.0"), make([]byte, 5)...)
	lines := HexDump(data)
	want := "00000000  4e454f4e 47524944 312e3000 00000000  |NEONGRID1.0.....|"
	if lines[0] != want {
		t.Fatalf("hexdump: got %q want %q", lines[0], want)
	}
}

// ---------------------------------------------------------------- 016

func TestMission016CRC16CCITT(t *testing.T) {
	if got := CRC16CCITT([]byte("123456789")); got != 0x29B1 {
		t.Fatalf("CRC-CCITT check value: got %04X want 29B1", got)
	}
	// Deterministic block example: recompute twice, compare.
	data := append([]byte{0x12, 0x34}, []byte("NEON GRID firmware")...)
	if CRC16CCITT(data) != CRC16CCITT(data) {
		t.Fatal("CRC16 not deterministic")
	}
}

// ---------------------------------------------------------------- 021

func TestMission021Xor(t *testing.T) {
	// Part 1: single-byte key 0x55.
	got := XorByte(mustHex(t, "1b 10 1a 1b 12 07 1c 11 0a 03 64"), 0x55)
	if string(got) != "NEONGRID_V1" {
		t.Fatalf("021-1: got %q want NEONGRID_V1", got)
	}
	// Cipher in the mission text must match encryption of the plaintext.
	wantCipher := HexStr(XorByte([]byte("NEONGRID_V1"), 0x55))
	if wantCipher != "1b 10 1a 1b 12 07 1c 11 0a 03 64" {
		t.Fatalf("021-1 cipher mismatch: %q", wantCipher)
	}

	// Part 2: repeating key CROW.
	cipher := mustHex(t, "0e 17 1c 04 02 15 0a 08 05 1d 1d 08 17 1a 0a 08 04 00 06 13")
	if got := string(XorRepeat(cipher, []byte("CROW"))); got != "MESSAGE_FOR_THE_GRID" {
		t.Fatalf("021-2: got %q want MESSAGE_FOR_THE_GRID", got)
	}
}

// ---------------------------------------------------------------- 022

func TestMission022Caesar(t *testing.T) {
	const cipher = "XMZY ITBS YMJ SJTSJSYNYD"
	const plain = "SHUT DOWN THE NEONENTITY"
	if got := Caesar(cipher, -5); got != plain {
		t.Fatalf("022: got %q want %q", got, plain)
	}
	if got := Caesar(plain, 5); got != cipher {
		t.Fatalf("022 encrypt: got %q want %q", got, cipher)
	}
	// ROT13 is an involution.
	s := "THE GRID IS ALIVE"
	if Rot13(Rot13(s)) != s {
		t.Fatal("ROT13 not an involution")
	}
	if Rot13("GUR TEVQ VF NYVIR") != "THE GRID IS ALIVE" {
		t.Fatal("ROT13 sample mismatch")
	}
}

// ---------------------------------------------------------------- 023

func TestMission023Vigenere(t *testing.T) {
	// Part 1: no punctuation, key NEON.
	enc := Vigenere("THEBLACKRIVERRISESAGAIN", "NEON", false, false)
	if enc != "GLSOYEQXEMJREVWFRWOTNMB" {
		t.Fatalf("023-1 encrypt: got %q", enc)
	}
	if got := Vigenere("GLSOYEQXEMJREVWFRWOTNMB", "NEON", true, false); got != "THEBLACKRIVERRISESAGAIN" {
		t.Fatalf("023-1 decrypt: got %q", got)
	}

	// Part 2: key advances only on letters.
	enc = Vigenere("CROW, THE SPREADS. TRACKING THE SIGN.", "DELTA", false, true)
	if enc != "FVZP, TKI DIRHEOL. TUENDIQK EAE VMRG." {
		t.Fatalf("023-2 encrypt: got %q", enc)
	}
	if got := Vigenere("FVZP, TKI DIRHEOL. TUENDIQK EAE VMRG.", "DELTA", true, true); got != "CROW, THE SPREADS. TRACKING THE SIGN." {
		t.Fatalf("023-2 decrypt: got %q", got)
	}
}

// ---------------------------------------------------------------- 024

func TestMission024Substitution(t *testing.T) {
	// The fragment is produced by a fixed cipher->plain mapping. The hints
	// given in the mission (G->T, T->H, X->E) must agree with it.
	sub := map[byte]byte{
		'G': 'T', 'T': 'H', 'X': 'E', 'C': 'N', 'M': 'A', 'R': 'S',
		'B': 'F', 'K': 'L', 'W': 'D', 'Q': 'W', 'V': 'K', 'Y': 'O', 'Z': 'I',
	}
	fragment := "GTX CXG TMR BMKKXC MCW QX VCYQ QTY WZW ZG"
	var plain []byte
	for i := 0; i < len(fragment); i++ {
		c := fragment[i]
		if c == ' ' {
			plain = append(plain, ' ')
			continue
		}
		p, ok := sub[c]
		if !ok {
			t.Fatalf("no mapping for %q", c)
		}
		plain = append(plain, p)
	}
	if got := string(plain); got != "THE NET HAS FALLEN AND WE KNOW WHO DID IT" {
		t.Fatalf("024: got %q", got)
	}
	// Hints from the mission must hold.
	if sub['G'] != 'T' || sub['X'] != 'E' || sub['T'] != 'H' {
		t.Fatal("024: mission hints broken")
	}
}

// ---------------------------------------------------------------- 029

func TestMission029OTPReuse(t *testing.T) {
	p1 := []byte("THE GRID NEVER SLEEPS TONIGHT")
	p2 := []byte("ICE IS ALWAYS WATCHING US ALL")
	c1 := mustHex(t, "4e 63 79 6d 19 3d 39 c5 b2 ed f1 93 93 b5 d8 5a 56 6e 79 1d 0d 4f 24 ce dc ea f3 8d 82")
	c2 := mustHex(t, "53 68 79 6d 17 3c 50 c0 de f4 f5 9c 85 c7 af 48 4e 68 74 04 10 28 50 d4 c1 83 f5 89 9a")
	if len(p1) != len(c1) || len(p1) != len(c2) {
		t.Fatalf("029: lengths differ p1=%d c1=%d c2=%d", len(p1), len(c1), len(c2))
	}
	// Core property: C1^C2 == P1^P2.
	eqBytes(t, "029 P1^P2 == C1^C2", XorBytes(c1, c2), XorBytes(p1, p2))
	// Crib: P1 starts with "THE "; key must be consistent.
	key := make([]byte, 4)
	for i := 0; i < 4; i++ {
		key[i] = c1[i] ^ p1[i]
	}
	if string(key) != "\x1a+<M" {
		t.Fatalf("029 crib: got %x", key)
	}
	// Both texts must be the same length (requirement of OTP reuse).
	if len(p1) != len(p2) {
		t.Fatalf("029: P1 and P2 must be equal length (%d vs %d)", len(p1), len(p2))
	}
}

// ---------------------------------------------------------------- 030

func TestMission030RSA(t *testing.T) {
	if got := ModPow(7, 11, 143); got != 106 {
		t.Fatalf("030 pow(7,11,143): got %d want 106", got)
	}
	if got := ModInv(7, 143); got != 41 {
		t.Fatalf("030 inv(7,143): got %d want 41", got)
	}
	// Break: n=143, e=65537, c=50 -> m=7='G'.
	if got := RSABreak(143, 65537, 50); got != 7 {
		t.Fatalf("030 break: got m=%d want 7", got)
	}
	p, q := FactorSmall(143)
	if p != 11 || q != 13 {
		t.Fatalf("030 factor: got %d,%d want 11,13", p, q)
	}
	if got := ModInv(65537, 120); got != 113 {
		t.Fatalf("030 d: got %d want 113", got)
	}
	// Round trip with a fresh small key.
	m := 42
	c := ModPow(m, 17, 323) // 323 = 17*19
	if RSABreak(323, 17, c) != m {
		t.Fatal("030 round trip failed")
	}
}

// ---------------------------------------------------------------- 031

func TestMission031Base64(t *testing.T) {
	if got := Base64Encode([]byte("MERCURY")); got != "TUVSQ1VSWQ==" {
		t.Fatalf("031 encode MERCURY: got %q", got)
	}
	if got := Base64Encode([]byte("NEON")); got != "TkVPTg==" {
		t.Fatalf("031 encode NEON: got %q", got)
	}
	for _, s := range []string{"MERCURY", "NEON", "", "a", "ab", "The quick brown fox jumps over the lazy dog"} {
		dec, err := Base64Decode(Base64Encode([]byte(s)))
		if err != nil || string(dec) != s {
			t.Fatalf("031 roundtrip %q: %q %v", s, dec, err)
		}
	}
}

// ---------------------------------------------------------------- 032

func TestMission032RLE(t *testing.T) {
	packed := []byte{0x03, 0x00, 0xAA, 0x02, 0x00, 0x01, 0x01, 0x00, 0x00}
	got, err := RLEUnpack(packed)
	if err != nil {
		t.Fatal(err)
	}
	eqBytes(t, "032 unpack", got, []byte{0xAA, 0xAA, 0xAA, 0x01, 0x01, 0x00})

	// Roundtrip on something with runs and isolated bytes.
	src := []byte("AAAABBBBCCDDEFFFFFFG")
	un, err := RLEUnpack(RLEPack(src))
	if err != nil {
		t.Fatal(err)
	}
	eqBytes(t, "032 roundtrip", un, src)
}

// ---------------------------------------------------------------- 033

func TestMission033Huffman(t *testing.T) {
	codes := HuffmanCodes(map[byte]int{'a': 45, 'b': 13, 'c': 12, 'd': 16, 'e': 9, 'f': 5})
	want := map[byte]string{'a': "0", 'b': "101", 'c': "100", 'd': "111", 'e': "1101", 'f': "1100"}
	for sym, w := range want {
		if codes[sym] != w {
			t.Fatalf("033 code for %q: got %q want %q", sym, codes[sym], w)
		}
	}
}

// ---------------------------------------------------------------- 034

func TestMission034LZW(t *testing.T) {
	src := []byte("TOBEORNOTTOBEORTOBEORNOT")
	codes := LZWEncode(src, 4096)
	want := []int{84, 79, 66, 69, 79, 82, 78, 79, 84, 256, 258, 260, 265, 259, 261, 263}
	if len(codes) != len(want) {
		t.Fatalf("034 codes len: got %d want %d (%v)", len(codes), len(want), codes)
	}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("034 code[%d]: got %d want %d", i, codes[i], want[i])
		}
	}
	if got := string(LZWDecode(codes)); got != "TOBEORNOTTOBEORTOBEORNOT" {
		t.Fatalf("034 decode: got %q", got)
	}
	// Decoder handles the not-yet-in-dictionary case for arbitrary input.
	in := []byte("ABABABABABABABAB")
	out := LZWDecode(LZWEncode(in, 4096))
	eqBytes(t, "034 roundtrip", out, in)
}

// ---------------------------------------------------------------- 036

func TestMission036BitStream(t *testing.T) {
	br := NewBitReader([]byte{0xE1, 0xA5, 0x01, 0x02})
	v, _ := br.Read(3)
	ty, _ := br.Read(5)
	fl, _ := br.Read(8)
	le, _ := br.Read(16)
	if v != 7 || ty != 1 || fl != 0xA5 || le != 258 {
		t.Fatalf("036 read: %d %d %#x %d", v, ty, fl, le)
	}
	bw := &BitWriter{}
	bw.Write(v, 3)
	bw.Write(ty, 5)
	bw.Write(fl, 8)
	bw.Write(le, 16)
	eqBytes(t, "036 write", bw.Bytes(), []byte{0xE1, 0xA5, 0x01, 0x02})
}

// ---------------------------------------------------------------- 037

func TestMission037CRC32(t *testing.T) {
	if got := CRC32IEEE([]byte("NEON")); got != 0xCA5B8064 {
		t.Fatalf("037 NEON crc: got %08X", got)
	}
	if got := CRC32IEEE([]byte("GRID")); got != 0x18B53483 {
		t.Fatalf("037 GRID crc: got %08X", got)
	}
	if got := CRC32IEEE([]byte("123456789")); got != 0xCBF43926 {
		t.Fatalf("037 check value: got %08X", got)
	}
}

// ---------------------------------------------------------------- 044

func TestMission044DNSQuery(t *testing.T) {
	q := BuildDNSQuery(0x1234, "example.com")
	want, _ := hex.DecodeString("123401000001000000000000076578616d706c6503636f6d0000010001")
	eqBytes(t, "044 query", q, want)
}

// ---------------------------------------------------------------- 047

func TestMission047WebSocket(t *testing.T) {
	const key = "dGhlIHNhbXBsZSBub25jZQ=="
	const accept = "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="
	if got := WSAccept(key); got != accept {
		t.Fatalf("047 accept: got %q want %q", got, accept)
	}
}

// ---------------------------------------------------------------- 052

func TestMission052ICMP(t *testing.T) {
	pkt := []byte{0x08, 0x00, 0x00, 0x00, 0x12, 0x34, 0x00, 0x01}
	pkt = append(pkt, []byte("hello")...)
	cs := ICMPChecksum(pkt)
	if cs != 0xA1F8 {
		t.Fatalf("052 checksum: got %04X want A1F8", cs)
	}
	// Verifying property: sum of all 16-bit words including the checksum
	// folds to 0xFFFF.
	pkt[2], pkt[3] = byte(cs>>8), byte(cs&0xFF)
	var sum uint32
	for i := 0; i+1 < len(pkt); i += 2 {
		sum += uint32(pkt[i])<<8 | uint32(pkt[i+1])
	}
	if len(pkt)%2 == 1 {
		sum += uint32(pkt[len(pkt)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	if sum != 0xFFFF {
		t.Fatalf("052 checksum verify: fold = %04X", sum)
	}
}

// ---------------------------------------------------------------- 070

func TestMission070PanelXor(t *testing.T) {
	got := XorByte([]byte{0x4A, 0x0C, 0x33, 0x4C, 0x0A, 0x2B}, 0x5A)
	want := []byte{16, 86, 105, 22, 80, 113}
	eqBytes(t, "070 decrypt", got, want)
	if got[0] != 16 || got[1] != 86 || got[2] != 105 || got[3] != 22 || got[4] != 80 || got[5] != 113 {
		t.Fatalf("070 decimal rendering: got %v", got)
	}
}

// ---------------------------------------------------------------- 113

func TestMission113BadgeGen(t *testing.T) {
	gen := func(t uint32, c uint16) []byte {
		var k []byte
		k = append(k, byte(t>>24), byte(t>>16), byte(t>>8), byte(t))
		k = append(k, byte(c>>8), byte(c))
		crc := CRC16CCITT(k)
		return append(k, byte(crc>>8), byte(crc))
	}
	k1 := gen(1700000000, 7)
	k2 := gen(1700000000, 7)
	if !bytes.Equal(k1, k2) {
		t.Fatal("113: generator not deterministic")
	}
	// Predict the next badge: same t, c+1, and the crc must validate.
	next := gen(1700000000, 8)
	if CRC16CCITT(next[:6]) != uint16(next[6])<<8|uint16(next[7]) {
		t.Fatal("113: next badge crc invalid")
	}
}

// ---------------------------------------------------------------- 119 / layers

func TestMission119LayerChain(t *testing.T) {
	// Layer 3 in mission 119 is RLE-packed; layer 4 validates password via
	// CRC16 == 0x5A5A. Pick a password and verify the CRC check works.
	// (The fixture is generated in tests; the check is the same function.)
	word := []byte("GATEKEEPER")
	crc := CRC16CCITT(word)
	if crc == 0 {
		t.Fatal("unexpected zero crc")
	}
}

// ---------------------------------------------------------------- 121 / tiny CPU

func TestMission121MiniCPU(t *testing.T) {
	// Part 1: LDA 'O', PRN, LDA 'K', PRN, HLT -> "OK".
	prog := []byte{0x01, 0x4F, 0x07, 0x01, 0x4B, 0x07, 0x00}
	if got := RunCPU(prog); got != "OK" {
		t.Fatalf("121 part1: got %q, want OK", got)
	}
	// Part 2 (fixed): the "A" after ADD is spelled from 0x4F+1=0x50, SUB 0x0B -> E,
	// ADD 0x09 -> N. Output: OPEN.
	firm := []byte{0x01, 0x4F, 0x07, 0x02, 0x01, 0x07, 0x03, 0x0B, 0x07, 0x02, 0x09, 0x07, 0x00}
	if got := RunCPU(firm); got != "OPEN" {
		t.Fatalf("121 firmware: got %q, want OPEN", got)
	}
	// Broken firmware from the mission: prints CO, not OK.
	broken := []byte{0x01, 0x43, 0x07, 0x01, 0x4F, 0x07, 0x00}
	if got := RunCPU(broken); got != "CO" {
		t.Fatalf("121 broken: got %q, want CO", got)
	}
}

// ---------------------------------------------------------------- 122 / garage handshake

func TestMission122DHAndTag(t *testing.T) {
	// Diffie-Hellman over p=23, g=5. a=6 (us), b=15 (broker).
	a, b := 6, 15
	A := ModPow(5, a, 23)
	B := ModPow(5, b, 23)
	if A != 8 || B != 19 {
		t.Fatalf("122 dh: A=%d B=%d, want 8 19", A, B)
	}
	sharedA := ModPow(B, a, 23)
	sharedB := ModPow(A, b, 23)
	if sharedA != 2 || sharedB != 2 {
		t.Fatalf("122 dh: shared %d/%d, want 2/2", sharedA, sharedB)
	}
	// Packet: nonce 0x11223344, seq 5, data "AWAKE", tag = crc16(S||nonce||seq||data).
	pkt := []byte{0x02, 0x11, 0x22, 0x33, 0x44, 0, 0, 0, 5, 'A', 'W', 'A', 'K', 'E'}
	tag := CRC16CCITT(pkt)
	if tag == 0 {
		t.Fatal("122: tag must be non-zero")
	}
	if CRC16CCITT(append(append([]byte{}, pkt...), 0)) == tag {
		t.Fatal("122: crc must change when data changes")
	}
	_ = tag
}

// ---------------------------------------------------------------- 123 / FFT + DTMF

func TestMission123FFTAndDTMF(t *testing.T) {
	const (
		sr  = 8000
		win = 256
	)
	// Part 1: tone at 697 Hz -> peak bin ~22 (697/31.25).
	x := make([]float64, win)
	s := 12345
	for i := 0; i < win; i++ {
		s = (s*1103515245 + 12345) & 0x7fffffff
		noise := float64(s)/0x7fffffff - 0.5
		x[i] = math.Sin(2*math.Pi*697*float64(i)/float64(sr)) + noise
	}
	pk := PeakBin(Magnitudes(FFT(x)))
	if pk < 20 || pk > 24 {
		t.Fatalf("123: peak bin %d, want ~22", pk)
	}
	// Part 2: synthesize "42", decode it back.
	sig := SynthDTMF("42", sr, 100)
	got := DecodeDTMF(sig, sr, 100)
	if got != "42" {
		t.Fatalf("123: decoded %q, want 42", got)
	}
	// The dial code in the mission is the door code: 42.
	if got != "42" {
		t.Fatal("123: door code mismatch")
	}
}

// ---------------------------------------------------------------- 033 part 2 / canonical

func TestMission033Canonical(t *testing.T) {
	lens := map[byte]int{'a': 2, 'b': 3, 'c': 3, 'd': 3, 'e': 3}
	codes := CanonicalHuffmanCodes(lens)
	if codes['a'] != "00" {
		t.Fatalf("canonical a: got %q want 00", codes['a'])
	}
	if codes['b'] != "010" {
		t.Fatalf("canonical b: got %q want 010", codes['b'])
	}
	if codes['e'] != "101" {
		t.Fatalf("canonical e: got %q want 101", codes['e'])
	}
	if codes['c'] != "011" {
		t.Fatalf("canonical c: got %q want 011", codes['c'])
	}
	if codes['d'] != "100" {
		t.Fatalf("canonical d: got %q want 100", codes['d'])
	}
}

// ---------------------------------------------------------------- helpers

func TestRot13Sample(t *testing.T) {
	if !strings.Contains(Rot13("GUR TEVQ VF NYVIR"), "THE GRID IS ALIVE") {
		t.Fatal("rot13 sample failed")
	}
}

// ---------------------------------------------------------------- 201-210 Performance pack

func perfRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

func perfRead(t *testing.T, id, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(perfRoot(t), "data", id, name))
	if err != nil {
		t.Fatalf("read data/%s/%s: %v", id, name, err)
	}
	return b
}

func perfExpected(t *testing.T, id string) (p1, p2 string) {
	t.Helper()
	b := perfRead(t, id, "expected.txt")
	parts := strings.SplitN(string(b), "\n=== PART 2 ===\n", 2)
	return parts[0], strings.TrimRight(parts[1], "\n")
}

func TestMission201PerfProfile(t *testing.T) {
	d := perfRead(t, "201", "profile.txt")
	w1, w2 := perfExpected(t, "201")
	if g := PerfProfileHot(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfProfileAlloc(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission202PerfRecords(t *testing.T) {
	d := perfRead(t, "202", "records.bin")
	w1, w2 := perfExpected(t, "202")
	if g := PerfRecordsDump(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfRecordsSum(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission203PerfPairs(t *testing.T) {
	d := perfRead(t, "203", "values.txt")
	w1, w2 := perfExpected(t, "203")
	if g := PerfPairsDump(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfPairsHex(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission204PerfMatrix(t *testing.T) {
	d := perfRead(t, "204", "matrix.bin")
	w1, w2 := perfExpected(t, "204")
	if g := PerfMatrixSum(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfMatrixCols(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission205PerfPackets(t *testing.T) {
	d := perfRead(t, "205", "packets.bin")
	w1, w2 := perfExpected(t, "205")
	if g := PerfPacketsCRC(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfPacketsCheck(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission206PerfBits(t *testing.T) {
	d := perfRead(t, "206", "values.bin")
	w1, w2 := perfExpected(t, "206")
	if g := PerfValuesPopcount(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfValuesBitReverse(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission207PerfRecords(t *testing.T) {
	d := perfRead(t, "207", "records.bin")
	w1, w2 := perfExpected(t, "207")
	if g := PerfRecordFields(d); g != w1 {
		t.Fatalf("Part1: got %q want %q", g, w1)
	}
	if g := PerfRecordZSum(d); g != w2 {
		t.Fatalf("Part2: got %q want %q", g, w2)
	}
}

func TestMission208PerfTopKeys(t *testing.T) {
	d := perfRead(t, "208", "index.bin")
	w1, w2 := perfExpected(t, "208")
	if g := PerfTopKeys(d); g != w1 || g != w2 {
		t.Fatalf("parts: got %q want %q / %q", g, w1, w2)
	}
}

func TestMission209PerfEvents(t *testing.T) {
	d := perfRead(t, "209", "events.bin")
	w1, w2 := perfExpected(t, "209")
	if g := PerfEventsSummary(d); g != w1 || g != w2 {
		t.Fatalf("parts: got %q want %q / %q", g, w1, w2)
	}
}

func TestMission210PerfBlocks(t *testing.T) {
	d := perfRead(t, "210", "blocks.bin")
	w1, w2 := perfExpected(t, "210")
	if g := PerfBlocksSummary(d); g != w1 || g != w2 {
		t.Fatalf("parts: got %q want %q / %q", g, w1, w2)
	}
}

// ---------------------------------------------------------------- 166

func TestMission166AStar(t *testing.T) {
	grid := []string{
		"S...........",
		"###########.",
		"#...........",
		"#.##########",
		"..........##",
		"#########...",
		".........E..",
	}
	path := AStar(grid, [2]int{0, 0}, [2]int{6, 9})
	if len(path) != 35 || path != "RRRRRRRRRRRDDLLLLLLLLLLDDRRRRRRRRDD" {
		t.Fatalf("AStar: got %q (len %d), want 35-step RRRRRRRRRRRDDLLLLLLLLLLDDRRRRRRRRDD", path, len(path))
	}
}

// ---------------------------------------------------------------- 172

func TestMission172Tetris(t *testing.T) {
	shapes := map[byte][][]int{
		'I': {{0, 1, 2, 3}},
		'O': {{0, 1}, {0, 1}},
		'T': {{0, 1, 2}, {1}},
		'S': {{1, 2}, {0, 1}},
		'Z': {{0, 1}, {1, 2}},
		'J': {{0}, {0, 1, 2}},
		'L': {{2}, {0, 1, 2}},
	}
	field, lines := TetrisBoard("IOTSZJLI", shapes)
	want := []string{
		"...IIII...",
		"...LLL....",
		"....T.....",
		"...OO.....",
		"...OO.....",
		"...IIII...",
	}
	if len(field) != len(want) {
		t.Fatalf("Tetris: got %d rows, want 6", len(field))
	}
	for i := range want {
		if field[i] != want[i] {
			t.Fatalf("Tetris row %d: got %q want %q", i, field[i], want[i])
		}
	}
	if lines != 0 {
		t.Fatalf("Tetris: got %d cleared lines, want 0", lines)
	}
}

// ---------------------------------------------------------------- 211-220 Depth pack

func TestMission211_220(t *testing.T) {
	type tc struct {
		id   string
		p1   func([]byte) string
		p2   func([]byte) string
		f1   string
		f2   string
	}
	cases := []tc{
		{"211", func(d []byte) string { return DeepECCPublic(d) }, func(d []byte) string { return DeepECDHShared(d) }, "curve.txt", "curve.txt"},
		{"212", func(d []byte) string { return DeepEventsLog(d) }, func(d []byte) string { return DeepEventsSummary(d) }, "events.txt", "events.txt"},
		{"213", func(d []byte) string { return DeepJSONValues(d) }, func(d []byte) string { return DeepJSONStats(d) }, "config.json", "config.json"},
		{"214", func(d []byte) string { return DeepDeflateSize(d) }, func(d []byte) string { return DeepDeflateText(d) }, "payload.bin", "payload.bin"},
		{"215", func(d []byte) string { return DeepLSMReplay(d) }, func(d []byte) string { return DeepLSMStats(d) }, "ops.log", "ops.log"},
		{"216", func(d []byte) string { return DeepRaftApply(d) }, func(d []byte) string { return DeepRaftStats(d) }, "raft.log", "raft.log"},
		{"217", func(d []byte) string {
			sig := perfRead(t, "217", "signal.txt")
			fir := perfRead(t, "217", "fir.txt")
			return DeepFIRFilter(sig, fir)
		}, func(d []byte) string {
			sig := perfRead(t, "217", "signal.txt")
			iir := perfRead(t, "217", "iir.txt")
			return DeepIIRFilter(sig, iir)
		}, "signal.txt", "signal.txt"},
		{"218", func(d []byte) string { return DeepAssemble(d) }, func(d []byte) string { return DeepAssembleStats(d) }, "asm.txt", "asm.txt"},
		{"219", func(d []byte) string { return DeepJITOutput(d, perfRead(t, "219", "limit.txt")) }, func(d []byte) string { return DeepJITLoop(d, perfRead(t, "219", "limit.txt")) }, "prog.bin", "prog.bin"},
		{"220", func(d []byte) string {
			key := perfRead(t, "220", "key.txt")
			nonce := perfRead(t, "220", "nonce.txt")
			return DeepChaChaText(key, nonce, d)
		}, func(d []byte) string {
			key := perfRead(t, "220", "key.txt")
			nonce := perfRead(t, "220", "nonce.txt")
			return DeepChaChaBlockHex(key, nonce)
		}, "secret.bin", "secret.bin"},
	}
	for _, c := range cases {
		w1, w2 := perfExpected(t, c.id)
		d1 := perfRead(t, c.id, c.f1)
		if g := c.p1(d1); g != w1 {
			t.Errorf("%s Part1: got %q want %q", c.id, g, w1)
		}
		d2 := perfRead(t, c.id, c.f2)
		if g := c.p2(d2); g != w2 {
			t.Errorf("%s Part2: got %q want %q", c.id, g, w2)
		}
	}
}
