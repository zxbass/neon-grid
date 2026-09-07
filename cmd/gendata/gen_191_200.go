package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"neon-grid/sol"
)

func goertzel(samples []float64, target float64, sr int) float64 {
	omega := 2 * math.Pi * target / float64(sr)
	coeff := 2 * math.Cos(omega)
	var s0, s1, s2 float64
	for _, x := range samples {
		s0 = x + coeff*s1 - s2
		s2 = s1
		s1 = s0
	}
	return s1*s1 + s2*s2 - coeff*s1*s2
}

func wavOf(sig []float64) []byte {
	wav := make([]byte, 44+2*len(sig))
	copy(wav[0:4], "RIFF")
	u32le(wav[4:], uint32(len(wav)-8))
	copy(wav[8:12], "WAVE")
	copy(wav[12:16], "fmt ")
	u32le(wav[16:], 16)
	u16le(wav[20:], 1)
	u16le(wav[22:], 1)
	u32le(wav[24:], 8000)
	u32le(wav[28:], 16000)
	u16le(wav[32:], 2)
	u16le(wav[34:], 16)
	copy(wav[36:40], "data")
	u32le(wav[40:], uint32(2*len(sig)))
	for i, v := range sig {
		u16le(wav[44+2*i:], uint16(int16(v*10000)))
	}
	return wav
}

// ---------------------------------------------------------------- 191 blue box

func (s *set) gen191() {
	id := "191"
	const sr = 8000
	digits := "519180"
	sig := sol.SynthDTMF(digits, sr, 100)
	// 2600 Hz carrier burst
	burst := make([]float64, sr/2)
	for n := range burst {
		burst[n] = math.Sin(2 * math.Pi * 2600 * float64(n) / sr)
	}
	sig = append(sig, burst...)
	sig = append(sig, sol.SynthDTMF("2600", sr, 100)...)
	s.file(id, "dial.wav", wavOf(sig))

	dec := sol.DecodeDTMF(sig, sr, 100)
	// detect carrier slots: FFT peak near 2600
	per := sr * 100 / 1000
	carrierSlots := 0
	for start := 0; start+per <= len(sig); start += per {
		mid := start + per/2 - 128
		w := make([]float64, 256)
		copy(w, sig[mid:mid+256])
		mag := sol.Magnitudes(sol.FFT(w))
		best, bf := 0.0, 0.0
		for k := 1; k < 128; k++ {
			f := float64(k) * sr / 256
			if mag[k] > best {
				best = mag[k]
				bf = f
			}
		}
		if bf > 2500 && bf < 2700 {
			carrierSlots++
		}
	}
	p1 := fmt.Sprintf("DIGITS: %s", dec)
	p2 := fmt.Sprintf("2600Hz BURST: %.3fs", float64(carrierSlots)*0.1)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 192 SSTV

func (s *set) gen192() {
	id := "192"
	const sr = 8000
	// 16x16 image: value = (x + 2*y) % 10
	var sig []float64
	pixTone := 0.05
	syncTone := 0.1
	for y := 0; y < 16; y++ {
		// sync 1900 Hz
		n := int(syncTone * sr)
		for k := 0; k < n; k++ {
			sig = append(sig, math.Sin(2*math.Pi*1900*float64(k)/sr))
		}
		for x := 0; x < 16; x++ {
			v := (x + 2*y) % 10
			f := 1500.0 + float64(v)*100.0
			n := int(pixTone * sr)
			for k := 0; k < n; k++ {
				sig = append(sig, math.Sin(2*math.Pi*f*float64(k)/sr))
			}
		}
	}
	s.file(id, "sstv.wav", wavOf(sig))

	// decode with Goertzel
	chars := " .:-=+*#%@"
	var p1 []string
	var row0 []string
	for y := 0; y < 16; y++ {
		var row []byte
		for x := 0; x < 16; x++ {
			start := int((syncTone+float64(16)*pixTone)*sr*float64(y)) + int(syncTone*sr) + x*int(pixTone*sr) + 100
			w := sig[start : start+256]
			best, bv := 1500.0, 0
			for v := 0; v < 10; v++ {
				if g := goertzel(w, 1500.0+float64(v)*100, sr); g > best {
					best = g
					bv = v
				}
			}
			row = append(row, chars[bv])
			if y == 0 {
				row0 = append(row0, fmt.Sprint(bv))
			}
		}
		p1 = append(p1, string(row))
	}
	p2 := "ROW 0: " + strings.Join(row0, " ")
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 193 MQTT

func mqttVarint(n int) []byte {
	var out []byte
	for {
		b := byte(n % 128)
		n /= 128
		if n > 0 {
			out = append(out, b|0x80)
		} else {
			out = append(out, b)
			return out
		}
	}
}

func mqttStr(s string) []byte {
	return append([]byte{byte(len(s) >> 8), byte(len(s))}, s...)
}

func (s *set) gen193() {
	id := "193"
	var pkt []byte
	// CONNECT
	connect := []byte{0x00, 0x04, 'M', 'Q', 'T', 'T', 0x04, 0x02, 0x00, 0x3C}
	connect = append(connect, mqttStr("zen-agent")...)
	pkt = append(pkt, 0x10)
	pkt = append(pkt, mqttVarint(len(connect))...)
	pkt = append(pkt, connect...)
	// PUBLISH neon/alerts
	pub := append(mqttStr("neon/alerts"), []byte("ICE_BREACH_LEVEL_3")...)
	pkt = append(pkt, 0x30)
	pkt = append(pkt, mqttVarint(len(pub))...)
	pkt = append(pkt, pub...)
	// SUBSCRIBE neon/#
	sub := append([]byte{0x00, 0x01}, mqttStr("neon/#")...)
	sub = append(sub, 0x00)
	pkt = append(pkt, 0x82)
	pkt = append(pkt, mqttVarint(len(sub))...)
	pkt = append(pkt, sub...)
	s.file(id, "mqtt.bin", pkt)

	p1 := "CONNECT: flags=0x02 client=zen-agent keepalive=60\nPUBLISH: topic=neon/alerts qos=0\nSUBSCRIBE: topic=neon/#"
	p2 := "PAYLOAD: ICE_BREACH_LEVEL_3"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 194 modbus

func (s *set) gen194() {
	id := "194"
	req := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x06, 0x01, 0x03, 0x00, 0x00, 0x00, 0x04}
	resp := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x0B, 0x01, 0x03, 0x08, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	s.text(id, "request.txt", hexBytes(req))
	s.text(id, "response.txt", hexBytes(resp))
	p1 := "TID: 1 UNIT: 1 FUNC: 3 ADDR: 0 COUNT: 4"
	p2 := "REGS: 0x1234 0x5678 0x9ABC 0xDEF0"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 195 polyglot

func (s *set) gen195() {
	id := "195"
	png := makePNG(2, 2)
	zip := buildZip([]zipFile{{"note.txt", []byte("NEON_HIDES_IN_ZIP")}})
	poly := append(append([]byte{}, png...), zip...)
	s.file(id, "poly.bin", poly)
	p1 := "PNG: 2x2 ZIP: note.txt"
	p2 := "EXTRACTED: NEON_HIDES_IN_ZIP"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 196 enigma crack

func (s *set) gen196() {
	id := "196"
	msg := "NEONGRIDSTATUSUNKNOWN"
	crib := "NEONGRID"
	order := [][]byte{rotorII, rotorI, rotorIII}
	pos := []int{3, 17, 9}
	plug := map[byte]byte{}
	cipher := enigmaCrypt(msg, order, pos, plug)
	s.text(id, "cipher.txt", cipher)
	s.text(id, "crib.txt", crib)
	var orders []string
	all := map[string][][]byte{}
	for _, a := range [][]byte{rotorI, rotorII, rotorIII} {
		for _, b := range [][]byte{rotorI, rotorII, rotorIII} {
			if string(b) == string(a) {
				continue
			}
			for _, c := range [][]byte{rotorI, rotorII, rotorIII} {
				if string(c) == string(a) || string(c) == string(b) {
					continue
				}
				name := rotorName(a) + "-" + rotorName(b) + "-" + rotorName(c)
				orders = append(orders, name)
				all[name] = [][]byte{a, b, c}
			}
		}
	}
	// sanity: brute force must recover the key
	found := ""
	for _, name := range orders {
		for p0 := 0; p0 < 26 && found == ""; p0++ {
			for p1 := 0; p1 < 26 && found == ""; p1++ {
				for p2 := 0; p2 < 26; p2++ {
					pt := enigmaCrypt(cipher, all[name], []int{p0, p1, p2}, plug)
					if strings.HasPrefix(pt, crib) {
						found = fmt.Sprintf("%s %d,%d,%d", name, p0, p1, p2)
						break
					}
				}
			}
		}
	}
	if found == "" {
		panic("196: brute force found nothing")
	}
	p1 := "KEY: " + found
	p2 := "PLAINTEXT: " + msg
	s.expected(id, p1, p2)
}

func rotorName(r []byte) string {
	switch string(r) {
	case string(rotorI):
		return "I"
	case string(rotorII):
		return "II"
	default:
		return "III"
	}
}

// ---------------------------------------------------------------- 197 BBS

func (s *set) gen197() {
	id := "197"
	menu := "== NEON-BBS v2.4 ==\n1: General\n2: Files\n3: SYSOP\n4: Exit"
	s.text(id, "bbs.txt", menu)
	p1 := "ITEMS: 4 (General Files SYSOP Exit)"
	p2 := "BOARD: SYSOP -> COMMAND 3"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 198 mavlink

func (s *set) gen198() {
	id := "198"
	mav := func(msgid, sysid byte, payload []byte) []byte {
		m := []byte{0xFE, byte(len(payload)), 0, sysid, 1, msgid}
		return append(m, payload...)
	}
	heartbeat := mav(0, 1, []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	att := make([]byte, 4+8*3)
	binary.LittleEndian.PutUint32(att[0:], 0)
	binary.LittleEndian.PutUint16(att[4:], 0)
	binary.LittleEndian.PutUint16(att[6:], 0)
	binary.LittleEndian.PutUint16(att[8:], 15708)
	attitude := mav(30, 1, att)
	gps := mav(24, 1, []byte{0x2A, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x01})
	var bin []byte
	bin = append(bin, heartbeat...)
	bin = append(bin, attitude...)
	bin = append(bin, gps...)
	s.file(id, "mav.bin", bin)
	p1 := "MSG 0: HEARTBEAT SYS=1\nMSG 1: ATTITUDE SYS=1\nMSG 2: GPS SYS=1"
	p2 := "ROLL: 0.0000 PITCH: 0.0000 YAW: 1.5708"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 199 EM4100

func (s *set) gen199() {
	id := "199"
	id64 := uint64(0x123456789A)
	// header 9x1
	bits := make([]bool, 0, 64)
	for i := 0; i < 9; i++ {
		bits = append(bits, true)
	}
	// 10 groups of 4 data bits
	var dataBits [40]bool
	for i := 39; i >= 0; i-- {
		dataBits[i] = id64&(1<<uint(i)) != 0
	}
	colPar := [4]bool{}
	for g := 0; g < 10; g++ {
		var par bool
		for i := 0; i < 4; i++ {
			b := dataBits[g*4+i]
			bits = append(bits, b)
			par = par != b
			if b {
				colPar[i] = !colPar[i]
			}
		}
		bits = append(bits, par)
	}
	for i := 0; i < 4; i++ {
		bits = append(bits, colPar[i])
	}
	bits = append(bits, false) // stop
	var pack []byte
	for i := 0; i < len(bits); i += 8 {
		var b byte
		for j := 0; j < 8 && i+j < len(bits); j++ {
			if bits[i+j] {
				b |= 1 << (7 - uint(j))
			}
		}
		pack = append(pack, b)
	}
	s.file(id, "tag.bin", pack)
	p1 := "ID: 0x123456789A"
	p2 := fmt.Sprintf("PARITY: OK\nHID: %d", id64)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 200 finale

func rot13(s string) string {
	out := make([]byte, len(s))
	for i, c := range []byte(s) {
		switch {
		case c >= 'A' && c <= 'Z':
			out[i] = (c-'A'+13)%26 + 'A'
		case c >= 'a' && c <= 'z':
			out[i] = (c-'a'+13)%26 + 'a'
		default:
			out[i] = c
		}
	}
	return string(out)
}

func (s *set) gen200() {
	id := "200"
	flag := "GRID_IS_ALIVE_2049"
	key := "GRID"
	step1 := rot13(flag)
	var step2 []byte
	for i := 0; i < len(step1); i++ {
		step2 = append(step2, step1[i]^key[i%len(key)])
	}
	chip := base64.StdEncoding.EncodeToString(step2)
	s.text(id, "chip.txt", chip)
	s.text(id, "key.txt", key)
	p1 := "FLAG: " + flag
	p2 := "LAYERS: BASE64 XOR ROT13 ALL OK"
	s.expected(id, p1, p2)
}

func init() {
	register("191", (*set).gen191)
	register("192", (*set).gen192)
	register("193", (*set).gen193)
	register("194", (*set).gen194)
	register("195", (*set).gen195)
	register("196", (*set).gen196)
	register("197", (*set).gen197)
	register("198", (*set).gen198)
	register("199", (*set).gen199)
	register("200", (*set).gen200)
}
