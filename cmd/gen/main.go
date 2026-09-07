package main

import (
	"fmt"

	"neon-grid/sol"
)

func main() {
	// 021 part 1: NEONGRID_V1 xor 0x55
	fmt.Println("021-1 cipher:", sol.HexStr(sol.XorByte([]byte("NEONGRID_V1"), 0x55)))
	// 021 part 2: MESSAGE_FOR_THE_GRID xor CROW
	fmt.Println("021-2 cipher:", sol.HexStr(sol.XorRepeat([]byte("MESSAGE_FOR_THE_GRID"), []byte("CROW"))))
	// 021 part 2 decrypt check
	dec, _ := sol.Hex("09 5e 52 1f 5e 13 52 0a 5f 06 5c 4e 5a 06 5b 1c 5f 4e 09 5e 54 1f 4d 1b 5a")
	fmt.Printf("021-2 old cipher decrypts to: %q\n", string(sol.XorRepeat(dec, []byte("CROW"))))

	// 022: encrypt SHUT DOWN THE NEONENTITY +5
	fmt.Println("022 cipher:", sol.Caesar("SHUT DOWN THE NEONENTITY", 5))
	fmt.Println("022 old cipher decrypt:", sol.Caesar("YNJY RGQJ YMJ YSTJXJINQD", -5))

	// 023 part 1: THEBLACKRIVERRISESAGAIN key NEON
	fmt.Println("023-1 cipher:", sol.Vigenere("THEBLACKRIVERRISESAGAIN", "NEON", false, false))
	fmt.Println("023-1 old cipher decrypt:", sol.Vigenere("VFHFVXRQLYIKTMAZIQBVCUIRB", "NEON", true, false))
	// 023 part 2: "CROW, THE SPREADS. TRACKING THE SIGN." key DELTA, skip
	fmt.Println("023-2 cipher:", sol.Vigenere("CROW, THE SPREADS. TRACKING THE SIGN.", "DELTA", false, true))
	fmt.Println("023-2 old cipher decrypt:", sol.Vigenere("XABYG, NRT GSEJZJ. RDLPIIYM EIQ JGWX.", "DELTA", true, true))

	// 029 OTP: fixed 29-byte key
	key, _ := sol.Hex("1a 2b 3c 4d 5e 6f 70 81 92 a3 b4 c5 d6 e7 f8 09 1a 2b 3c 4d 5e 6f 70 81 92 a3 b4 c5 d6")
	p1 := []byte("THE GRID NEVER SLEEPS TONIGHT")
	p2 := []byte("ICE IS ALWAYS WATCHING US ALL")
	c1 := sol.XorBytes(p1, key)
	c2 := sol.XorBytes(p2, key)
	fmt.Println("029 key len:", len(key), "p1 len:", len(p1), "p2 len:", len(p2))
	fmt.Println("029 c1:", sol.HexStr(c1))
	fmt.Println("029 c2:", sol.HexStr(c2))
	fmt.Println("029 c1^c2==p1^p2:", sol.HexStr(sol.XorBytes(c1, c2)) == sol.HexStr(sol.XorBytes(p1, p2)))
	// brute force with known "THE" crib at both starts to sanity check
	crib := []byte("THE ")
	k0 := make([]byte, 4)
	for i := 0; i < 4; i++ {
		k0[i] = c1[i] ^ crib[i]
	}
	fmt.Printf("029 crib key from c1 start: %q\n", string(k0))

	// 030 RSA
	p, q := sol.FactorSmall(143)
	fmt.Println("030 p q:", p, q)
	phi := (p - 1) * (q - 1)
	d := sol.ModInv(65537, phi)
	fmt.Println("030 d:", d)
	c := sol.ModPow(7, 65537, 143)
	fmt.Println("030 c for m=7 e=65537:", c)
	fmt.Println("030 m from c:", sol.ModPow(c, d, 143), "->", string(rune(sol.ModPow(c, d, 143))))

	// 037 CRC32
	fmt.Printf("037 NEON crc: %08X GRID crc: %08X\n", sol.CRC32IEEE([]byte("NEON")), sol.CRC32IEEE([]byte("GRID")))

	// 052 ICMP
	icmp := []byte{8, 0, 0, 0, 0x12, 0x34, 0x00, 0x01}
	icmp = append(icmp, []byte("hello")...)
	cs := sol.ICMPChecksum(icmp)
	fmt.Printf("052 checksum: %04X\n", cs)

	// 070 panel
	fmt.Println("070 decrypted:", sol.XorByte([]byte{0x4A, 0x0C, 0x33, 0x4C, 0x0A, 0x2B}, 0x5A))

	// 031 base64
	fmt.Println("031 MERCURY:", sol.Base64Encode([]byte("MERCURY")))
	fmt.Println("031 NEON:", sol.Base64Encode([]byte("NEON")))

	// 047 websocket
	fmt.Println("047 accept:", sol.WSAccept("dGhlIHNhbXBsZSBub25jZQ=="))

	// 034 LZW
	codes := sol.LZWEncode([]byte("TOBEORNOTTOBEORTOBEORNOT"), 4096)
	fmt.Println("034 codes:", codes)
	fmt.Println("034 decode:", string(sol.LZWDecode(codes)))

	// 033 Huffman
	codesH := sol.HuffmanCodes(map[byte]int{'a': 45, 'b': 13, 'c': 12, 'd': 16, 'e': 9, 'f': 5})
	fmt.Println("033 huffman:", codesH)

	// 036 bitstream
	br := sol.NewBitReader([]byte{0xE1, 0xA5, 0x01, 0x02})
	v, _ := br.Read(3)
	t, _ := br.Read(5)
	f, _ := br.Read(8)
	l, _ := br.Read(16)
	fmt.Printf("036 version=%d type=%d flags=%#X length=%d\n", v, t, f, l)
	bw := &sol.BitWriter{}
	bw.Write(v, 3)
	bw.Write(t, 5)
	bw.Write(f, 8)
	bw.Write(l, 16)
	fmt.Printf("036 written: %X\n", bw.Bytes())

	// 032 RLE
	un, err := sol.RLEUnpack([]byte{0x03, 0x00, 0xAA, 0x02, 0x00, 0x01, 0x01, 0x00, 0x00})
	fmt.Println("032 unpack:", err, sol.HexStr(un))

	// 016 CRC16 sanity
	fmt.Printf("016 crc16 CCITT(empty-ish): %04X\n", sol.CRC16CCITT([]byte("123456789")))
}
