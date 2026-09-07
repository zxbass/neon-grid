package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"neon-grid/sol"
)

// ---------------------------------------------------------------- 131 Enigma

var (
	rotorI   = []byte("EKMFLGDQVZNTOWYHXUSPAIBRCJ")
	rotorII  = []byte("AJDKSIRUXBLHWTMCQGZNPYFVOE")
	rotorIII = []byte("BDFHJLCPRTXVZNYEIWGAKMUSQO")
	reflectB = []byte("YRUHQSLDPXNGOKMIEBFZCWVJAT")
)

func enigmaCrypt(msg string, order [][]byte, pos []int, plug map[byte]byte) string {
	inv := make([][]byte, 3)
	for i, r := range order {
		inv[i] = make([]byte, 26)
		for j, c := range r {
			inv[i][c-'A'] = byte(j) + 'A'
		}
	}
	p := append([]int{}, pos...)
	fwd := func(c byte, r, i []byte, off int) byte {
		x := int(c-'A'+byte(off)) % 26
		x = (int(r[x]-'A') - off) % 26
		if x < 0 {
			x += 26
		}
		return byte(x) + 'A'
	}
	out := make([]byte, 0, len(msg))
	for _, ch := range []byte(msg) {
		p[0] = (p[0] + 1) % 26
		if p[0] == 0 {
			p[1] = (p[1] + 1) % 26
			if p[1] == 0 {
				p[2] = (p[2] + 1) % 26
			}
		}
		c := ch
		if q, ok := plug[c]; ok {
			c = q
		}
		c = fwd(c, order[0], inv[0], p[0])
		c = fwd(c, order[1], inv[1], p[1])
		c = fwd(c, order[2], inv[2], p[2])
		c = reflectB[c-'A']
		c = fwd(c, inv[2], order[2], p[2])
		c = fwd(c, inv[1], order[1], p[1])
		c = fwd(c, inv[0], order[0], p[0])
		if q, ok := plug[c]; ok {
			c = q
		}
		out = append(out, c)
	}
	return string(out)
}

func (s *set) gen131() {
	id := "131"
	order := [][]byte{rotorI, rotorII, rotorIII}
	plug := map[byte]byte{}
	for _, pair := range []string{"AB", "CD"} {
		plug[pair[0]] = pair[1]
		plug[pair[1]] = pair[0]
	}
	pos := []int{0, 0, 0}
	msg := "HELLOGRIDFROMZEN"
	cipher := enigmaCrypt(msg, order, pos, plug)
	back := enigmaCrypt(cipher, order, pos, plug)
	s.text(id, "config.txt", "ROTORS: I II III\nPOS: AAA\nPLUGBOARD: AB CD\n")
	s.text(id, "message.txt", msg)
	s.expected(id, cipher, back)
}

// ---------------------------------------------------------------- 132 SHA-256

func (s *set) gen132() {
	id := "132"
	msg := "GRID NEVER DIES 2049"
	s.text(id, "msg.txt", msg)
	h1 := fmt.Sprintf("%x", sha256.Sum256([]byte(msg)))
	h2 := fmt.Sprintf("%x", sha256.Sum256([]byte(h1)))
	s.expected(id, h1, h2)
}

// ---------------------------------------------------------------- 133 padding oracle

func pkcs7Pad(b []byte, bs int) []byte {
	n := bs - len(b)%bs
	return append(append([]byte{}, b...), bytesRepeat(byte(n), n)...)
}

func bytesRepeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func aesCBCEncrypt(key, iv, pt []byte) []byte {
	blk, _ := aes.NewCipher(key)
	padded := pkcs7Pad(pt, blk.BlockSize())
	ct := make([]byte, len(padded))
	cipher.NewCBCEncrypter(blk, iv).CryptBlocks(ct, padded)
	return ct
}

func (s *set) gen133() {
	id := "133"
	key := []byte("PADORACLE1337KEY")
	iv := []byte("0123456789ABCDEF")
	msgs := []string{"THE GRID PAYS FOR EVERYTHING", "MEET AT THE NEON DRAGON"}
	var lines []string
	for _, m := range msgs {
		ct := aesCBCEncrypt(key, iv, []byte(m))
		lines = append(lines, hex.EncodeToString(iv)+hex.EncodeToString(ct))
	}
	s.text(id, "key.txt", string(key))
	s.text(id, "cipher.txt", strings.Join(lines, "\n"))
	p1 := "PAYLOAD: THE GRID PAYS FOR EVERYTHING"
	p2 := "PAYLOAD: MEET AT THE NEON DRAGON"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 134 CBC bit flip

func (s *set) gen134() {
	id := "134"
	key := []byte("BITFLIPKEY1337!!")
	iv := []byte("AAAAAAAAAAAAAAAA")
	pt := []byte("SEND 111 TO CROW!!")
	ct := aesCBCEncrypt(key, iv, pt)
	token := append(append([]byte{}, iv...), ct...)
	s.text(id, "token.txt", hex.EncodeToString(token))

	mod := append([]byte{}, iv...)
	// '1' (0x31) -> '9' (0x39): flip bit 0x08 at positions 5,6,7 of IV
	for i := 5; i <= 7; i++ {
		mod[i] ^= 0x08
	}
	full := append(append([]byte{}, mod...), ct...)
	s.expected(id, hex.EncodeToString(full), "SEND 999 TO CROW!!")
}

// ---------------------------------------------------------------- 135 hash length extension

var shaK = [64]uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

func rotr(x uint32, n uint) uint32 { return x>>n | x<<(32-n) }

func sha256Compress(h []uint32, chunk []byte) {
	var w [64]uint32
	for i := 0; i < 16; i++ {
		w[i] = uint32(chunk[i*4])<<24 | uint32(chunk[i*4+1])<<16 | uint32(chunk[i*4+2])<<8 | uint32(chunk[i*4+3])
	}
	for i := 16; i < 64; i++ {
		s0 := rotr(w[i-15], 7) ^ rotr(w[i-15], 18) ^ (w[i-15] >> 3)
		s1 := rotr(w[i-2], 17) ^ rotr(w[i-2], 19) ^ (w[i-2] >> 10)
		w[i] = w[i-16] + s0 + w[i-7] + s1
	}
	a, b, c, d, e, f, g, hh := h[0], h[1], h[2], h[3], h[4], h[5], h[6], h[7]
	for i := 0; i < 64; i++ {
		S1 := rotr(e, 6) ^ rotr(e, 11) ^ rotr(e, 25)
		ch := (e & f) ^ (^e & g)
		t1 := hh + S1 + ch + shaK[i] + w[i]
		S0 := rotr(a, 2) ^ rotr(a, 13) ^ rotr(a, 22)
		maj := (a & b) ^ (a & c) ^ (b & c)
		t2 := S0 + maj
		hh, g, f, e = g, f, e, d+t1
		d, c, b, a = c, b, a, t1+t2
	}
	h[0] += a
	h[1] += b
	h[2] += c
	h[3] += d
	h[4] += e
	h[5] += f
	h[6] += g
	h[7] += hh
}

// sha256WithState continues hashing from an internal state (for length extension).
func sha256WithState(h []uint32, msg []byte, totalLen uint64) []uint32 {
	var buf []byte
	buf = append(buf, msg...)
	buf = append(buf, 0x80)
	for len(buf)%64 != 56 {
		buf = append(buf, 0)
	}
	bits := (totalLen) * 8
	buf = append(buf,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
	st := append([]uint32{}, h...)
	for i := 0; i < len(buf); i += 64 {
		sha256Compress(st, buf[i:i+64])
	}
	return st
}

func sha256State(msg []byte) []uint32 {
	h := []uint32{
		0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
		0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
	}
	var buf []byte
	buf = append(buf, msg...)
	buf = append(buf, 0x80)
	for len(buf)%64 != 56 {
		buf = append(buf, 0)
	}
	bits := uint64(len(msg)) * 8
	buf = append(buf,
		byte(bits>>56), byte(bits>>48), byte(bits>>40), byte(bits>>32),
		byte(bits>>24), byte(bits>>16), byte(bits>>8), byte(bits))
	for i := 0; i < len(buf); i += 64 {
		sha256Compress(h, buf[i:i+64])
	}
	return h
}

func sha256HexFromState(st []uint32) string {
	var out strings.Builder
	for _, w := range st {
		fmt.Fprintf(&out, "%08x", w)
	}
	return out.String()
}

// sha256PadFor returns the padding bytes for a message whose length is `total`.
func sha256PadFor(total uint64) []byte {
	pad := []byte{0x80}
	for (total+uint64(len(pad)))%64 != 56 {
		pad = append(pad, 0)
	}
	bits := total * 8
	var lenField [8]byte
	for i := 0; i < 8; i++ {
		lenField[i] = byte(bits >> (56 - 8*i))
	}
	return append(pad, lenField[:]...)
}

func (s *set) gen135() {
	id := "135"
	key := []byte("SUPERSECRETKEY12")
	msg := []byte("admin=0")
	keyLen := len(key) + len(msg)
	mac := sha256.Sum256(append(append([]byte{}, key...), msg...))
	s.text(id, "input.txt", fmt.Sprintf("KEYLEN: %d\nMESSAGE: %s\nMAC: %x", len(key), msg, mac))

	st := sha256State(append(append([]byte{}, key...), msg...))

	forge := func(suffix []byte) (macStr string, payload []byte) {
		pad := sha256PadFor(uint64(keyLen))
		payload = append(append(append([]byte{}, msg...), pad...), suffix...)
		st2 := sha256WithState(st, suffix, uint64(keyLen+len(pad)+len(suffix)))
		macStr = sha256HexFromState(st2)
		full := sha256.Sum256(append(append([]byte{}, key...), payload...))
		if fmt.Sprintf("%x", full) != macStr {
			panic("135: hash length extension sanity check failed")
		}
		return macStr, payload
	}

	mac2, _ := forge([]byte("&admin=1"))
	mac3, payload3 := forge([]byte("&role=zen"))
	s.expected(id,
		fmt.Sprintf("FORGED_MAC: %s", mac2),
		fmt.Sprintf("FORGED_MAC2: %s\nPAYLOAD: %s", mac3, hex.EncodeToString(payload3)))
}

// ---------------------------------------------------------------- 136 LCG

const (
	lcgA = uint32(1664525)
	lcgC = uint32(1013904223)
)

func lcgNext(x *uint32) uint32 {
	*x = lcgA**x + lcgC
	return *x
}

func invMod2_32(a uint32) uint32 {
	oldR, r := int64(a), int64(1)<<32
	oldS, s := int64(1), int64(0)
	for r != 0 {
		q := oldR / r
		oldR, r = r, oldR-q*r
		oldS, s = s, oldS-q*s
	}
	m := int64(1) << 32
	return uint32((oldS%m + m) % m)
}

func (s *set) gen136() {
	id := "136"
	var x uint32 = 123456789
	outs := []uint32{}
	for i := 0; i < 5; i++ {
		outs = append(outs, lcgNext(&x))
	}
	s.text(id, "lcg.txt", fmt.Sprintf("%08X %08X %08X\n", outs[0], outs[1], outs[2]))
	// solve: a = (X2-X1)/(X1-X0) mod 2^32
	d1 := outs[1] - outs[0]
	d2 := outs[2] - outs[1]
	a := uint32(uint64(d2) * uint64(invMod2_32(d1)) & 0xFFFFFFFF)
	c := outs[1] - a*outs[0]
	x3 := a*outs[2] + c
	x4 := a*x3 + c
	x5 := a*x4 + c
	if x3 != outs[3] || x4 != outs[4] {
		panic("136: LCG recovery mismatch")
	}
	p1 := fmt.Sprintf("%08X\n%08X\n%08X", x3, x4, x5)
	p2 := fmt.Sprintf("A=0x%08X\nC=0x%08X", a, c)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 137 MT19937

type mt19937 struct {
	mt  [624]uint32
	idx int
}

func (m *mt19937) seed(s uint32) {
	m.mt[0] = s
	for i := 1; i < 624; i++ {
		m.mt[i] = 1812433253*(m.mt[i-1]^(m.mt[i-1]>>30)) + uint32(i)
	}
	m.idx = 624
}

func (m *mt19937) next() uint32 {
	if m.idx >= 624 {
		for i := 0; i < 624; i++ {
			y := (m.mt[i] & 0x80000000) + (m.mt[(i+1)%624] & 0x7fffffff)
			m.mt[i] = m.mt[(i+397)%624] ^ (y >> 1)
			if y&1 != 0 {
				m.mt[i] ^= 0x9908b0df
			}
		}
		m.idx = 0
	}
	y := m.mt[m.idx]
	m.idx++
	y ^= y >> 11
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= y >> 18
	return y
}

func untemper(y uint32) uint32 {
	// undo the temper transforms in reverse order.
	// undo: y ^= y >> 18 (one pass: 18 >= 16)
	y ^= y >> 18
	// undo: y ^= (y << shift) & mask, iterating 5 times (covers shift 7 and 15)
	undoLeft := func(x uint32, shift uint, mask uint32) uint32 {
		x0 := x
		for i := 0; i < 5; i++ {
			x = x0 ^ ((x << shift) & mask)
		}
		return x
	}
	y = undoLeft(y, 15, 0xefc60000)
	y = undoLeft(y, 7, 0x9d2c5680)
	// undo: y ^= y >> 11 (two passes: 11, 22)
	y ^= y >> 11 ^ y >> 22
	return y
}

func (s *set) gen137() {
	id := "137"
	var m mt19937
	m.seed(1337)
	var out []uint32
	for i := 0; i < 5; i++ {
		out = append(out, m.next())
	}
	s.text(id, "mt.txt", "SEED: 1337\n")
	ut := untemper(out[4])
	line := make([]string, len(out))
	for i, v := range out {
		line[i] = fmt.Sprint(v)
	}
	p1 := strings.Join(line, " ")
	p2 := fmt.Sprintf("UNTEMPER(%d) = %d\nSTATE0: %d", out[4], ut, ut)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 138 RSA small e

func base26(s string) int64 {
	n := int64(0)
	for _, c := range s {
		n = n*26 + int64(c-'A')
	}
	return n
}

func decBase26(n int64) string {
	if n == 0 {
		return "A"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte(n%26) + 'A'}, b...)
		n /= 26
	}
	return string(b)
}

func (s *set) gen138() {
	id := "138"
	p, q := int64(1000000007), int64(1000000009)
	N := p * q
	words := []string{"ZEN", "GRID"}
	var cs []int64
	for i, w := range words {
		m := base26(w)
		c := new(big.Int).Exp(big.NewInt(m), big.NewInt(3), nil)
		cr := new(big.Int).Mod(c, big.NewInt(N)).Int64()
		cs = append(cs, cr)
		s.text(id, fmt.Sprintf("msg%d.txt", i+1), w)
	}
	s.text(id, "rsa.txt", fmt.Sprintf("N=%d\nE=3\nC1=%d\nC2=%d", N, cs[0], cs[1]))
	// integer cube root (message is small: m^3 < N)
	cb := func(cc int64) int64 {
		lo, hi := int64(0), int64(2000000)
		for lo < hi {
			mid := (lo + hi + 1) / 2
			if mid*mid*mid <= cc {
				lo = mid
			} else {
				hi = mid - 1
			}
		}
		return lo
	}
	r1, r2 := cb(cs[0]), cb(cs[1])
	if r1*r1*r1 != cs[0] || r2*r2*r2 != cs[1] {
		panic("138: cube root mismatch")
	}
	s.expected(id,
		fmt.Sprintf("%d -> %s", r1, decBase26(r1)),
		fmt.Sprintf("%d -> %s", r2, decBase26(r2)))
}

// ---------------------------------------------------------------- 139 CRC collision

func crcAppendForce(msg []byte, target uint32) []byte {
	// f(x) = crc32(msg || x), x = 32 appended bits.
	// f is affine over GF(2): f(x) = base ^ XOR_i(x_i * col[i]),
	// where base = f(0) and col[i] = f(e_i) ^ base.
	base := sol.CRC32IEEE(append(append([]byte{}, msg...), 0, 0, 0, 0))
	var col [32]uint32
	for bit := 0; bit < 32; bit++ {
		v := append([]byte{}, msg...)
		v = append(v, 0, 0, 0, 0)
		v[len(msg)+bit/8] |= 1 << (7 - uint(bit)%8)
		col[bit] = sol.CRC32IEEE(v) ^ base
	}
	rhs := target ^ base
	// rows: A[i] = i-th equation; unknown j occupies bit (31-j).
	var A [32]uint32
	var rhsBit [32]bool
	for i := 0; i < 32; i++ {
		for j := 0; j < 32; j++ {
			if col[j]&(1<<(31-uint(i))) != 0 {
				A[i] |= 1 << (31 - uint(j))
			}
		}
		rhsBit[i] = rhs&(1<<(31-uint(i))) != 0
	}
	var sel [32]int
	for i := range sel {
		sel[i] = -1
	}
	pivot := 0
	for j := 0; j < 32 && pivot < 32; j++ {
		mask := uint32(1) << (31 - uint(j))
		p := -1
		for r := pivot; r < 32; r++ {
			if A[r]&mask != 0 {
				p = r
				break
			}
		}
		if p < 0 {
			continue
		}
		A[pivot], A[p] = A[p], A[pivot]
		rhsBit[pivot], rhsBit[p] = rhsBit[p], rhsBit[pivot]
		sel[j] = pivot
		for r := 0; r < 32; r++ {
			if r != pivot && A[r]&mask != 0 {
				A[r] ^= A[pivot]
				rhsBit[r] = rhsBit[r] != rhsBit[pivot]
			}
		}
		pivot++
	}
	var x uint32
	for j := 0; j < 32; j++ {
		if sel[j] >= 0 && rhsBit[sel[j]] {
			x |= 1 << (31 - uint(j))
		}
	}
	out := append(append([]byte{}, msg...), byte(x>>24), byte(x>>16), byte(x>>8), byte(x))
	if sol.CRC32IEEE(out) != target {
		panic("139: CRC force sanity check failed")
	}
	return out
}

func (s *set) gen139() {
	id := "139"
	honest := []byte("PAYLOAD SECURED")
	mine := []byte("TAMPERED DATA!")
	target := sol.CRC32IEEE(honest)
	crafted := crcAppendForce(mine, target)
	s.text(id, "honest.txt", hex.EncodeToString(honest))
	s.text(id, "mine.txt", hex.EncodeToString(mine))
	got := sol.CRC32IEEE(crafted)
	s.expected(id, hex.EncodeToString(crafted), fmt.Sprintf("CRC: 0x%08X", got))
}

// ---------------------------------------------------------------- 140 RC4 key reuse

func rc4(key, data []byte) []byte {
	var sbox [256]byte
	for i := range sbox {
		sbox[i] = byte(i)
	}
	j := 0
	for i := range sbox {
		j = (j + int(sbox[i]) + int(key[i%len(key)])) % 256
		sbox[i], sbox[j] = sbox[j], sbox[i]
	}
	out := make([]byte, len(data))
	i, j := 0, 0
	for n := range data {
		i = (i + 1) % 256
		j = (j + int(sbox[i])) % 256
		sbox[i], sbox[j] = sbox[j], sbox[i]
		out[n] = data[n] ^ sbox[(int(sbox[i])+int(sbox[j]))%256]
	}
	return out
}

func (s *set) gen140() {
	id := "140"
	key := []byte("WEPKEY12")
	p1 := []byte("FLAG=GRID_WEP_IS_DEAD=")
	p2 := []byte("TRAFFIC ROUTED VIA ZEN")
	// reuse attack: один и тот же поток ключа шифрует оба пакета.
	// rc4(key, zeros) даёт чистый keystream (XOR с нулями).
	ks := rc4(key, make([]byte, len(p2)))
	c1 := make([]byte, len(p1))
	c2 := make([]byte, len(p2))
	for i := range p1 {
		c1[i] = p1[i] ^ ks[i]
	}
	for i := range p2 {
		c2[i] = p2[i] ^ ks[i]
	}
	s.text(id, "wep.txt", fmt.Sprintf("C1: %x\nC2: %x\nP1: %s", c1, c2, p1))
	p1exp := fmt.Sprintf("P2: %s", p2)
	p2exp := fmt.Sprintf("KEYSTREAM: %x", ks[:8])
	s.expected(id, p1exp, p2exp)
}

func init() {
	register("131", (*set).gen131)
	register("132", (*set).gen132)
	register("133", (*set).gen133)
	register("134", (*set).gen134)
	register("135", (*set).gen135)
	register("136", (*set).gen136)
	register("137", (*set).gen137)
	register("138", (*set).gen138)
	register("139", (*set).gen139)
	register("140", (*set).gen140)
}
