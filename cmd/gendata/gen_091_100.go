package main

// Generators for missions 091-100 (audio/video/imaging part 2). All
// generators for this group self-register below.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// ---------------------------------------------------------------- 091

var morseCode = map[byte]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.",
	'G': "--.", 'H': "....", 'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..",
	'M': "--", 'N': "-.", 'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.",
	'S': "...", 'T': "-", 'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-",
	'Y': "-.--", 'Z': "--..",
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
}

var morseRev = func() map[string]byte {
	m := map[string]byte{}
	for c, code := range morseCode {
		m[code] = c
	}
	return m
}()

// morseSegments encodes text (words split by '_') into a string of '#'
// (signal) and '.' (silence) 20ms blocks, one char per block.
func morseSegments(text string) string {
	var b strings.Builder
	b.WriteString("..")
	words := strings.Split(text, "_")
	for wi, w := range words {
		for li := 0; li < len(w); li++ {
			code := morseCode[w[li]]
			for ei := 0; ei < len(code); ei++ {
				if code[ei] == '.' {
					b.WriteString("#")
				} else {
					b.WriteString("###")
				}
				if ei < len(code)-1 {
					b.WriteString(".")
				}
			}
			if li < len(w)-1 {
				b.WriteString("...")
			}
		}
		if wi < len(words)-1 {
			b.WriteString(".....")
		}
	}
	b.WriteString("..")
	return b.String()
}

// morseParse decodes a segments string into morse tokens (per word) and
// plain text (words joined by '_').
func morseParse(seg string) (string, string) {
	var words []string
	curMorse := ""
	var curWord []string
	flushLetter := func() {
		if curMorse != "" {
			curWord = append(curWord, curMorse)
			curMorse = ""
		}
	}
	flushWord := func() {
		flushLetter()
		if len(curWord) > 0 {
			words = append(words, strings.Join(curWord, " "))
			curWord = nil
		}
	}
	for i := 0; i < len(seg); {
		j := i
		for j < len(seg) && seg[j] == seg[i] {
			j++
		}
		run := j - i
		if seg[i] == '#' {
			if run <= 2 {
				curMorse += "."
			} else {
				curMorse += "-"
			}
		} else {
			switch {
			case run <= 2:
			case run <= 4:
				flushLetter()
			default:
				flushWord()
			}
		}
		i = j
	}
	flushWord()
	var text []string
	for _, w := range words {
		var sb strings.Builder
		for _, tok := range strings.Fields(w) {
			sb.WriteByte(morseRev[tok])
		}
		text = append(text, sb.String())
	}
	return strings.Join(words, "  "), strings.Join(text, "_")
}

func (s *set) gen091() {
	id := "091"
	const sr = 8000
	const block = 160
	msg := "NEONSECTOR_AR"
	seg := morseSegments(msg)
	mor, text := morseParse(seg)

	var samples []byte
	for _, c := range seg {
		amp := 60.0
		if c == '.' {
			amp = 4.0
		}
		for i := 0; i < block; i++ {
			v := 128 + int(amp*math.Sin(2*math.Pi*800*float64(i)/sr+0.3))
			samples = append(samples, byte(v))
		}
	}
	s.file(id, "morse.wav", wav8(samples, sr))

	part1 := "SEGMENTS: " + seg
	part2 := "MORSE: " + mor + "\nTEXT: " + text
	s.expected(id, part1, part2)
}

// ---------------------------------------------------------------- 092

func (s *set) gen092() {
	id := "092"
	const sr = 8000
	const block = 320
	msg := "NEONGRID_RELAY_SECTOR7_CHAN_B_OPEN_CROW_NEON_OMEGA_ACK"
	for len(msg) < 56 {
		msg += "_"
	}
	msg = msg[:56]

	var bits []byte
	for _, b := range []byte(msg) {
		for m := 7; m >= 0; m-- {
			bits = append(bits, (b>>uint(m))&1)
		}
	}
	for i := 0; i < 64; i++ {
		bits = append(bits, 2) // noise block marker
	}

	var samples []byte
	for _, b := range bits {
		f := 1000.0
		if b == 1 {
			f = 2000.0
		}
		if b == 2 {
			f = 1500.0
		}
		amp := 40.0
		if b == 2 {
			amp = 20.0
		}
		for i := 0; i < block; i++ {
			v := 128 + int(amp*math.Sin(2*math.Pi*f*float64(i)/sr+0.3))
			samples = append(samples, byte(v))
		}
	}
	s.file(id, "noise.wav", wav8(samples, sr))

	det := func(sl []byte) byte {
		changes := 0
		prev := float64(sl[0]) - 128
		for i := 1; i < len(sl); i++ {
			cur := float64(sl[i]) - 128
			if (prev < 0 && cur >= 0) || (prev >= 0 && cur < 0) {
				changes++
			}
			prev = cur
		}
		freq := float64(changes/2) * 100 / 4
		if math.Abs(freq-1000) <= 150 {
			return 0
		}
		if math.Abs(freq-2000) <= 300 {
			return 1
		}
		return 2
	}
	var out []byte
	for b := 0; b < len(bits); b++ {
		out = append(out, det(samples[b*block:(b+1)*block]))
	}

	part1 := fmt.Sprintf("BITS: %s... (len=%d)", bitsToStr(out[:64]), len(out))

	var valid []byte
	for i := 0; i < len(out); i += 8 {
		b := 0
		ok := true
		for m := 0; m < 8; m++ {
			if out[i+m] == 2 {
				ok = false
				break
			}
			b = b<<1 | int(out[i+m])
		}
		if ok {
			valid = append(valid, byte(b))
		}
	}
	noisePct := 0
	for _, b := range out {
		if b == 2 {
			noisePct++
		}
	}
	var hexp []string
	for i := 0; i < 4 && i < len(valid); i++ {
		hexp = append(hexp, fmt.Sprintf("0x%02X", valid[i]))
	}
	part2 := fmt.Sprintf("BYTES: %s ... (%d bytes)\nTEXT: %s\nNOISE: %.1f%% ambiguous bits",
		strings.Join(hexp, " "), len(valid), string(valid), float64(noisePct)/float64(len(out))*100)
	s.expected(id, part1, part2)
}

func bitsToStr(b []byte) string {
	var sb strings.Builder
	for _, x := range b {
		if x == 2 {
			sb.WriteByte('?')
		} else {
			sb.WriteByte('0' + x)
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------- 093

func gifInterlaceRows(ind []byte, w, h int) []byte {
	out := make([]byte, 0, w*h)
	for pass := 0; pass < 4; pass++ {
		start, step := 0, 8
		switch pass {
		case 0:
			start, step = 0, 8
		case 1:
			start, step = 4, 8
		case 2:
			start, step = 2, 4
		case 3:
			start, step = 1, 2
		}
		for y := start; y < h; y += step {
			out = append(out, ind[y*w:(y+1)*w]...)
		}
	}
	return out
}

func gifUnInterlaceRows(ind []byte, w, h int) []byte {
	out := make([]byte, w*h)
	pos := 0
	for pass := 0; pass < 4; pass++ {
		start, step := 0, 8
		switch pass {
		case 0:
			start, step = 0, 8
		case 1:
			start, step = 4, 8
		case 2:
			start, step = 2, 4
		case 3:
			start, step = 1, 2
		}
		for y := start; y < h; y += step {
			copy(out[y*w:(y+1)*w], ind[pos:pos+w])
			pos += w
		}
	}
	return out
}

func gif093Build(w, h int, delays []int, frames [][]byte) []byte {
	var b []byte
	b = append(b, "GIF89a"...)
	b = append(b, byte(w), byte(w>>8), byte(h), byte(h>>8))
	b = append(b, 0x80|7, 0, 0) // global table: 256 entries
	for i := 0; i < 256; i++ {
		b = append(b, byte(i), byte(i*3), byte(i*5))
	}
	for fi, ind := range frames {
		b = append(b, 0x21, 0xF9, 0x04, 0x04, byte(delays[fi]), byte(delays[fi]>>8), 0, 0)
		b = append(b, 0x2C, 0, 0, 0, 0, byte(w), byte(w>>8), byte(h), byte(h>>8), 0x40)
		packed := gifLZW(gifInterlaceRows(ind, w, h), 8)
		b = append(b, 8)
		for len(packed) > 0 {
			l := len(packed)
			if l > 255 {
				l = 255
			}
			b = append(b, byte(l))
			b = append(b, packed[:l]...)
			packed = packed[l:]
		}
		b = append(b, 0)
	}
	b = append(b, 0x3B)
	return b
}

type gifFrame struct{ w, h, delay int; pix []byte }

func gif093Decode(g []byte) (w, h, colors, bg int, frames []gifFrame) {
	w = int(g[6]) | int(g[7])<<8
	h = int(g[8]) | int(g[9])<<8
	flags := g[10]
	bg = int(g[11])
	colors = 1 << int(flags&7+1)
	i := 13 + colors*3
	pendingDelay := 0
	for i < len(g) {
		switch g[i] {
		case 0x2C:
			iw := int(g[i+5]) | int(g[i+6])<<8
			ih := int(g[i+7]) | int(g[i+8])<<8
			iflags := g[i+9]
			i += 10
			lit := int(g[i])
			i++
			var packed []byte
			for {
				l := int(g[i])
				i++
				if l == 0 {
					break
				}
				packed = append(packed, g[i:i+l]...)
				i += l
			}
			ind, ok := gifLZWDecode(packed, lit)
			if !ok {
				panic("gif lzw")
			}
			if iflags&0x40 != 0 {
				ind = gifUnInterlaceRows(ind, iw, ih)
			}
			frames = append(frames, gifFrame{iw, ih, pendingDelay, ind})
			pendingDelay = 0
		case 0x21:
			i++
			label := g[i]
			i++
			size := int(g[i])
			i++
			if label == 0xF9 && size == 4 {
				pendingDelay = int(g[i+1]) | int(g[i+2])<<8
			}
			i += size
			i++
		case 0x3B:
			return
		default:
			i++
		}
	}
	return
}

func (s *set) gen093() {
	id := "093"
	const w, h = 16, 16
	bg := func(x, y int) int { return (x*3 + y*5) % 256 }
	key := []byte("NEON")
	var frames [][]byte
	for _, k := range key {
		fr := make([]byte, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				fr[y*w+x] = byte(bg(x, y))
			}
		}
		fr[5*w+5] = k
		frames = append(frames, fr)
	}
	delays := []int{10, 10, 10, 10}
	gif := gif093Build(w, h, delays, frames)
	s.file(id, "poltergeist.gif", gif)

	gw, gh, colors, bgc, gframes := gif093Decode(gif)
	part1 := fmt.Sprintf("GIF89a %dx%d global_colors=%d bg=%d", gw, gh, colors, bgc)
	var p2 strings.Builder
	var keyBytes []byte
	for fi, fr := range gframes {
		fmt.Fprintf(&p2, "FRAME %d: %dx%d delay=%d local_pal=no\n", fi, fr.w, fr.h, fr.delay)
		for y := 0; y < fr.h; y++ {
			for x := 0; x < fr.w; x++ {
				if int(fr.pix[y*fr.w+x]) != bg(x, y) {
					keyBytes = append(keyBytes, fr.pix[y*fr.w+x])
				}
			}
		}
	}
	var hexp []string
	for _, k := range keyBytes {
		hexp = append(hexp, fmt.Sprintf("%02x", k))
	}
	fmt.Fprintf(&p2, "KEY LAYER: %s -> %s", strings.Join(hexp, " "), string(keyBytes))
	s.expected(id, part1, p2.String())
}

// ---------------------------------------------------------------- 094

func bmp24(w, h int) []byte {
	rowBytes := (w*3 + 3) &^ 3
	imgSize := rowBytes * h
	bmp := make([]byte, 54+imgSize)
	copy(bmp[0:2], "BM")
	u32le(bmp[2:], uint32(len(bmp)))
	u32le(bmp[10:], 54)
	u32le(bmp[14:], 40)
	u32le(bmp[18:], uint32(w))
	u32le(bmp[22:], uint32(h))
	u16le(bmp[26:], 1)
	u16le(bmp[28:], 24)
	u32le(bmp[34:], uint32(imgSize))
	return bmp
}

func asciiChar(b int) byte {
	levels := " .:-=+*#%@"
	idx := b * 10 / 256
	if idx > 9 {
		idx = 9
	}
	return levels[idx]
}

func (s *set) gen094() {
	id := "094"
	const w, h = 160, 64
	bright := func(x, y int) int {
		return (x + y) * 255 / (w + h - 2)
	}
	rowBytes := (w*3 + 3) &^ 3
	bmp := bmp24(w, h)
	for sy := 0; sy < h; sy++ {
		row := bmp[54+(h-1-sy)*rowBytes:]
		for x := 0; x < w; x++ {
			b := bright(x, sy)
			row[x*3], row[x*3+1], row[x*3+2] = byte(b), byte(b), byte(b)
		}
	}
	s.file(id, "face.bmp", bmp)

	var p1 strings.Builder
	for sy := 0; sy < h; sy++ {
		for x := 0; x < w; x++ {
			p1.WriteByte(asciiChar(bright(x, sy)))
		}
		p1.WriteByte('\n')
	}

	const outW = 80
	outH := h * outW / w / 2
	blockW := w / outW
	blockH := h / outH
	var p2 strings.Builder
	fmt.Fprintf(&p2, "SRC: %dx%d  OUT: %dx%d  (%d cols)\n", w, h, outW, outH, outW)
	for oy := 0; oy < outH; oy++ {
		for ox := 0; ox < outW; ox++ {
			sum := 0
			for dy := 0; dy < blockH; dy++ {
				for dx := 0; dx < blockW; dx++ {
					sum += bright(ox*blockW+dx, oy*blockH+dy)
				}
			}
			p2.WriteByte(asciiChar(sum / (blockW * blockH)))
		}
		p2.WriteByte('\n')
	}
	s.expected(id, strings.TrimRight(p1.String(), "\n"), strings.TrimRight(p2.String(), "\n"))
}

// ---------------------------------------------------------------- 095

var font57 = map[byte][]string{
	'S': {".XXX.", "X...X", "X....", ".XXX.", "....X", "X...X", ".XXX."},
	'T': {"XXXXX", "..X..", "..X..", "..X..", "..X..", "..X..", "..X.."},
	'R': {"XXXX.", "X...X", "X...X", "XXXX.", "X.X..", "X..X.", "X...X"},
	'E': {"XXXXX", "X....", "X....", "XXXX.", "X....", "X....", "XXXXX"},
	'L': {"X....", "X....", "X....", "X....", "X....", "X....", "XXXXX"},
	'I': {".XXX.", "..X..", "..X..", "..X..", "..X..", "..X..", ".XXX."},
	'G': {".XXX.", "X...X", "X....", "X.XXX", "X...X", "X...X", ".XXX."},
	'H': {"X...X", "X...X", "X...X", "XXXXX", "X...X", "X...X", "X...X"},
}

func (s *set) gen095() {
	id := "095"
	const N = 64
	pal := make([][3]byte, 256)
	cga := [][3]byte{
		{0, 0, 0}, {255, 0, 0}, {0, 255, 0}, {0, 0, 255},
		{255, 255, 0}, {0, 255, 255}, {255, 0, 255}, {192, 192, 192},
	}
	copy(pal[:8], cga)
	pal[8] = [3]byte{5, 8, 12}     // background
	pal[9] = [3]byte{240, 220, 180} // text
	for i := 10; i < 256; i++ {
		v := byte(12 + (i-10)*200/245)
		pal[i] = [3]byte{v, v, v}
	}
	pixels := make([]byte, N*N)
	for i := range pixels {
		pixels[i] = 8
	}
	words := []string{"STREET", "LIGHTS"}
	for wi, w := range words {
		x0 := (N - len(w)*6) / 2
		y0 := 6 + wi*9
		for li, ch := range []byte(w) {
			for gy, row := range font57[ch] {
				for gx := 0; gx < 5; gx++ {
					if row[gx] == 'X' {
						pixels[(y0+gy)*N+x0+li*6+gx] = 9
					}
				}
			}
		}
	}

	var file bytes.Buffer
	file.WriteString("INDX")
	u16le16(&file, uint16(N))
	u16le16(&file, uint16(N))
	u16le16(&file, 256)
	for _, c := range pal {
		file.Write([]byte{c[0], c[1], c[2]})
	}
	file.Write(pixels)
	s.file(id, "screen.ind", file.Bytes())

	var p1 strings.Builder
	fmt.Fprintf(&p1, "256 colors  %dx%d\n", N, N)
	var ps []string
	for i := 0; i < 8; i++ {
		ps = append(ps, fmt.Sprintf("PAL[%d]=%02X%02X%02X", i, pal[i][0], pal[i][1], pal[i][2]))
	}
	p1.WriteString(strings.Join(ps, " "))

	var p2 strings.Builder
	p2.WriteString("REORDERED\n")
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			c := pal[pixels[y*N+x]]
			sum := int(c[0]) + int(c[1]) + int(c[2])
			level := sum * 3 / 766
			switch {
			case level >= 2:
				p2.WriteByte('@')
			case level == 1:
				p2.WriteByte('#')
			default:
				p2.WriteByte('.')
			}
		}
		p2.WriteByte('\n')
	}
	p2.WriteString("KEY: STREET_LIGHTS")
	s.expected(id, p1.String(), p2.String())
}

func u16le16(b *bytes.Buffer, v uint16) {
	var tmp [2]byte
	binary.LittleEndian.PutUint16(tmp[:], v)
	b.Write(tmp[:])
}

// ---------------------------------------------------------------- 096

func wav16(samples []int16, sr int) []byte {
	wav := make([]byte, 44+2*len(samples))
	copy(wav[0:4], "RIFF")
	u32le(wav[4:], uint32(len(wav)-8))
	copy(wav[8:12], "WAVE")
	copy(wav[12:16], "fmt ")
	u32le(wav[16:], 16)
	u16le(wav[20:], 1)
	u16le(wav[22:], 1)
	u32le(wav[24:], uint32(sr))
	u32le(wav[28:], uint32(sr*2))
	u16le(wav[32:], 2)
	u16le(wav[34:], 16)
	copy(wav[36:40], "data")
	u32le(wav[40:], uint32(2*len(samples)))
	for i, v := range samples {
		u16le(wav[44+2*i:], uint16(v))
	}
	return wav
}

func wav8(samples []byte, sr int) []byte {
	wav := make([]byte, 44+len(samples))
	copy(wav[0:4], "RIFF")
	u32le(wav[4:], uint32(len(wav)-8))
	copy(wav[8:12], "WAVE")
	copy(wav[12:16], "fmt ")
	u32le(wav[16:], 16)
	u16le(wav[20:], 1)
	u16le(wav[22:], 1)
	u32le(wav[24:], uint32(sr))
	u32le(wav[28:], uint32(sr))
	u16le(wav[32:], 1)
	u16le(wav[34:], 8)
	copy(wav[36:40], "data")
	u32le(wav[40:], uint32(len(samples)))
	copy(wav[44:], samples)
	return wav
}

func msgToBits(data []byte) []byte {
	var bits []byte
	for _, b := range data {
		for m := 7; m >= 0; m-- {
			bits = append(bits, (b>>uint(m))&1)
		}
	}
	return bits
}

func bitsToMsg(bits []byte) []byte {
	out := make([]byte, 0, len(bits)/8)
	for i := 0; i+8 <= len(bits); i += 8 {
		b := 0
		for m := 0; m < 8; m++ {
			b = b<<1 | int(bits[i+m]&1)
		}
		out = append(out, byte(b))
	}
	return out
}

func (s *set) gen096() {
	id := "096"
	const sr = 8000
	tone := func(i int) int16 {
		return int16(6000 * math.Sin(2*math.Pi*440*float64(i)/sr+0.3))
	}

	// part 1: LSB stego, 1 bit per sample.
	payload1 := "DROP_AT_NEON_PIER"
	msg1 := append(make([]byte, 4), []byte(payload1)...)
	u32le(msg1, uint32(len(payload1)))
	bits1 := msgToBits(msg1)
	samples1 := make([]int16, 600)
	for i := range samples1 {
		samples1[i] = tone(i)
	}
	for i, b := range bits1 {
		samples1[i] = int16(uint16(samples1[i]&^1) | uint16(b))
	}
	s.file(id, "rain.wav", wav16(samples1, sr))

	var extract1 []byte
	{
		var bits []byte
		for _, smp := range samples1 {
			bits = append(bits, byte(smp&1))
		}
		msg := bitsToMsg(bits)
		n := int(binary.LittleEndian.Uint32(msg[:4]))
		extract1 = msg[4 : 4+n]
	}
	part1 := "PAYLOAD: " + string(extract1)

	// part 2: 3x redundancy + noise + majority vote.
	payload2 := "NEON_SECTOR_7"
	msg2 := append(make([]byte, 4), []byte(payload2)...)
	u32le(msg2, uint32(len(payload2)))
	bits2 := msgToBits(msg2)
	stream := make([]byte, 0, len(bits2)*3)
	for _, b := range bits2 {
		stream = append(stream, b, b, b)
	}
	samples2 := make([]int16, len(stream))
	for i := range samples2 {
		samples2[i] = tone(i)
	}
	for i, b := range stream {
		samples2[i] = int16(uint16(samples2[i]&^1) | uint16(b))
	}
	stegoBytes := wav16(samples2, sr)
	s.file(id, "stego.wav", stegoBytes)

	flipped := 0
	for i := range samples2 {
		if i%37 == 7 {
			samples2[i] ^= 1
			flipped++
		}
	}
	var recv []byte
	for k := 0; k < len(bits2); k++ {
		votes := 0
		for m := 0; m < 3; m++ {
			votes += int(samples2[k*3+m] & 1)
		}
		recv = append(recv, byte(votes/2))
	}
	recovered := bitsToMsg(recv)
	decodable := "NO"
	if bytes.Equal(recovered, msg2) {
		decodable = "YES"
	}
	part2 := fmt.Sprintf("STEGO WROTE: %d bytes\nNOISE ERROR: %.1f%% bits flipped -> still decodable: %s",
		len(stegoBytes), float64(flipped)/float64(len(stream))*100, decodable)
	s.expected(id, part1, part2)
}

// ---------------------------------------------------------------- 097

func (s *set) gen097() {
	id := "097"
	const sr = 8000
	const block = 80
	msg := "TRANSMISSION_RECEIVED"
	pad := msg
	for len(pad) < 30 {
		pad += "_"
	}
	enc := func(b byte) byte {
		b ^= 0xA5
		return b>>4 | b<<4
	}
	frameData := make([]byte, 32)
	frameData[0], frameData[31] = 0x7E, 0x7E
	for i := 0; i < 30; i++ {
		frameData[i+1] = enc(pad[i])
	}
	var bits []byte
	for f := 0; f < 3; f++ {
		bits = append(bits, msgToBits(frameData)...)
	}
	var samples []int16
	for _, b := range bits {
		f := 1200.0
		if b == 1 {
			f = 2400.0
		}
		for i := 0; i < block; i++ {
			samples = append(samples, int16(6000*math.Sin(2*math.Pi*f*float64(i)/sr+0.3)))
		}
	}
	s.file(id, "fsk.wav", wav16(samples, sr))

	var out []byte
	for b := 0; b < len(bits); b++ {
		sl := samples[b*block : (b+1)*block]
		changes := 0
		prev := sl[0]
		for i := 1; i < len(sl); i++ {
			if (prev < 0 && sl[i] >= 0) || (prev >= 0 && sl[i] < 0) {
				changes++
			}
			prev = sl[i]
		}
		freq := changes / 2 * 100
		if freq >= 1800 {
			out = append(out, 1)
		} else {
			out = append(out, 0)
		}
	}
	part1 := fmt.Sprintf("BITS: %s...", bitsToStr(out[:64]))

	bytesAll := bitsToMsg(out)
	var frames []string
	i := 0
	for i+32 <= len(bytesAll) {
		if bytesAll[i] == 0x7E {
			data := bytesAll[i+1 : i+31]
			var dec []byte
			for _, b := range data {
				d := b>>4 | b<<4
				d ^= 0xA5
				dec = append(dec, d)
			}
			frames = append(frames, strings.TrimRight(string(dec), "_"))
			i += 32
		} else {
			i++
		}
	}
	var p2 strings.Builder
	for _, fr := range frames {
		fmt.Fprintf(&p2, "FRAME: %s\n", fr)
	}
	fmt.Fprintf(&p2, "FRAMES: %d", len(frames))
	s.expected(id, part1, p2.String())
}

// ---------------------------------------------------------------- 098

func (s *set) gen098() {
	id := "098"
	const N = 21
	msg := "WELCOME_TO_THE_GRID"
	body := append([]byte{byte(len(msg))}, []byte(msg)...)
	crc := byte(0)
	for _, b := range []byte(msg) {
		crc ^= b
	}
	body = append(body, crc)
	dataBits := msgToBits(body)

	skip := func(x, y int) bool {
		if (x < 7 && y < 7) || (x < 7 && y >= N-7) || (x >= N-7 && y < 7) {
			return true
		}
		if x == 6 || y == 6 {
			return true
		}
		return false
	}

	modules := make([][]byte, N)
	for i := range modules {
		modules[i] = make([]byte, N)
	}
	drawFinder := func(x0, y0 int) {
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 7; dx++ {
				if dx == 0 || dx == 6 || dy == 0 || dy == 6 || (dx >= 2 && dx <= 4 && dy >= 2 && dy <= 4) {
					modules[y0+dy][x0+dx] = 1
				}
			}
		}
	}
	drawFinder(0, 0)
	drawFinder(N-7, 0)
	drawFinder(0, N-7)
	for x := 8; x <= N-9; x++ {
		modules[6][x] = byte(x % 2)
	}
	for y := 8; y <= N-9; y++ {
		modules[y][6] = byte(y % 2)
	}

	var walk [][2]int
	dir := -1
	for colPair := 0; ; colPair += 2 {
		right := N - 1 - colPair
		left := right - 1
		if right < 0 {
			break
		}
		for _, col := range []int{right, left} {
			if col < 0 {
				continue
			}
			if dir == -1 {
				for y := N - 1; y >= 0; y-- {
					walk = append(walk, [2]int{col, y})
				}
			} else {
				for y := 0; y < N; y++ {
					walk = append(walk, [2]int{col, y})
				}
			}
			dir = -dir
		}
	}
	pos := 0
	for _, p := range walk {
		if skip(p[0], p[1]) {
			continue
		}
		if pos < len(dataBits) {
			modules[p[1]][p[0]] = dataBits[pos]
		} else {
			padByte := byte(0xEC)
			if (pos/8)%2 == 1 {
				padByte = 0x11
			}
			modules[p[1]][p[0]] = (padByte >> uint(7-(pos%8))) & 1
		}
		pos++
	}

	const scale = 8
	bmp := bmp24(N*scale, N*scale)
	rowBytes := (N*scale*3 + 3) &^ 3
	for my := 0; my < N; my++ {
		for mx := 0; mx < N; mx++ {
			bit := modules[my][mx]
			v := byte(255)
			if bit == 1 {
				v = 0
			}
			for dy := 0; dy < scale; dy++ {
				row := bmp[54+(N*scale-1-(my*scale+dy))*rowBytes:]
				for dx := 0; dx < scale; dx++ {
					o := (mx*scale+dx)*3 + 0
					row[o], row[o+1], row[o+2] = v, v, v
				}
			}
		}
	}
	s.file(id, "grid.bmp", bmp)

	// reference decode
	var grid []byte
	for my := 0; my < N; my++ {
		for mx := 0; mx < N; mx++ {
			fileRow := N*scale - 1 - (my*scale + scale/2)
			px := mx*scale + scale/2
			o := fileRow*rowBytes + px*3
			b := int(bmp[54+o])
			if b > 128 {
				grid = append(grid, 0)
			} else {
				grid = append(grid, 1)
			}
		}
	}
	var got []byte
	for _, p := range walk {
		if skip(p[0], p[1]) {
			continue
		}
		b := grid[p[1]*N+p[0]]
		got = append(got, b)
	}
	bodyGot := bitsToMsg(got[:len(dataBits)])

	part1 := fmt.Sprintf("QR %dx%d finders at (0,0),(0,%d),(%d,0)", N, N, N-7, N-7)
	part2 := "DATA: " + string(bodyGot[1:1+int(bodyGot[0])])
	s.expected(id, part1, part2)
}

// ---------------------------------------------------------------- 099

func (s *set) gen099() {
	id := "099"
	const w, h = 32, 32
	bg := func(x, y int) byte { return byte(((x / 4) + (y / 4)) % 4) }
	frame := func(rects ...rect) []byte {
		f := make([]byte, w*h)
		for i := range f {
			f[i] = bg(i%w, i/w)
		}
		for _, r := range rects {
			for yy := r.y; yy < r.y+r.h; yy++ {
				for xx := r.x; xx < r.x+r.w; xx++ {
					f[yy*w+xx] = r.data[(yy-r.y)*r.w+(xx-r.x)]
				}
			}
		}
		return f
	}
	solid := func(x, y, rw, rh int, c byte) rect {
		return rect{x, y, rw, rh, bytes.Repeat([]byte{c}, rw*rh)}
	}
	checker := func(x, y, rw, rh int) rect {
		d := make([]byte, rw*rh)
		for i := range d {
			if i%2 == 0 {
				d[i] = 12
			} else {
				d[i] = 13
			}
		}
		return rect{x, y, rw, rh, d}
	}

	f0 := frame()
	frames := []struct {
		first byte
		rects []rect
	}{{0xFF, nil}, {0x00, []rect{solid(4, 4, 8, 6, 12), solid(20, 10, 6, 8, 13)}}, {0x00, []rect{checker(4, 20, 8, 6)}}, {0x00, []rect{solid(12, 8, 4, 4, 7)}}}

	var file bytes.Buffer
	u16le16(&file, uint16(w))
	u16le16(&file, uint16(h))
	for _, fr := range frames {
		file.WriteByte(fr.first)
		if fr.first == 0x00 {
			u16le16(&file, uint16(len(fr.rects)))
			for _, r := range fr.rects {
				u16le16(&file, uint16(r.x))
				u16le16(&file, uint16(r.y))
				u16le16(&file, uint16(r.w))
				u16le16(&file, uint16(r.h))
				file.Write(r.data)
			}
		} else {
			file.Write(f0)
		}
	}
	s.file(id, "projector.bin", file.Bytes())

	var p1 strings.Builder
	var p2 strings.Builder
	fmt.Fprintf(&p1, "FRAME 0: full (%d px)\n", w*h)

	state := make([]byte, w*h)
	for i := range state {
		state[i] = byte(((i%w)/4 + (i/w)/4) % 4)
	}
	for fi := 1; fi <= 3; fi++ {
		prev := make([]byte, w*h)
		copy(prev, state)
		fr := frames[fi]
		for _, r := range fr.rects {
			for yy := r.y; yy < r.y+r.h; yy++ {
				for xx := r.x; xx < r.x+r.w; xx++ {
					state[yy*w+xx] = r.data[(yy-r.y)*r.w+(xx-r.x)]
				}
			}
		}
		changed := 0
		for i := range state {
			if state[i] != prev[i] {
				changed++
			}
		}
		plural := "rect"
		if len(fr.rects) != 1 {
			plural = "rects"
		}
		line := fmt.Sprintf("FRAME %d: %d %s, %d px changed\n", fi, len(fr.rects), plural, changed)
		if fi == 3 {
			p2.WriteString(line)
			r := fr.rects[0]
			fmt.Fprintf(&p2, "FRAME %d: pointer at (%d,%d) size %dx%d color=%d", fi, r.x, r.y, r.w, r.h, r.data[0])
		} else {
			p1.WriteString(line)
		}
	}
	s.expected(id, p1.String(), p2.String())
}

type rect struct{ x, y, w, h int; data []byte }

// ---------------------------------------------------------------- 100

func dpcmErr(wave []int16, adaptive bool) float64 {
	step := 32.0
	pred := 0.0
	var sum float64
	for _, s := range wave {
		e := float64(s) - pred
		q := math.Round(e / step)
		if q < -8 {
			q = -8
		}
		if q > 7 {
			q = 7
		}
		rec := pred + q*step
		sum += math.Abs(float64(s) - rec)
		pred = rec
		if adaptive {
			aq := math.Abs(q)
			if aq >= 5 {
				step = math.Min(256, step*1.5)
			}
			if aq <= 2 {
				step = math.Max(4, step/1.2)
			}
		}
	}
	return sum / float64(len(wave))
}

func (s *set) gen100() {
	id := "100"
	const N = 4096
	wave := make([]int16, N)
	for i := range wave {
		v := 1500*math.Sin(2*math.Pi*2*float64(i)/N) +
			700*math.Sin(2*math.Pi*17*float64(i)/N) +
			300*math.Sin(2*math.Pi*43*float64(i)/N)
		wave[i] = int16(v)
	}
	var wb bytes.Buffer
	for _, v := range wave {
		var tmp [2]byte
		binary.LittleEndian.PutUint16(tmp[:], uint16(v))
		wb.Write(tmp[:])
	}
	s.file(id, "wave.bin", wb.Bytes())

	e1 := dpcmErr(wave, false)
	e2 := dpcmErr(wave, true)
	better := (e1 - e2) / e1 * 100
	part1 := fmt.Sprintf("CODEC: 4-bit DPCM  avg_err=%.1f  ratio=4.0x", e1)
	part2 := fmt.Sprintf("DPCM:  avg_err=%.1f\nADAPT: avg_err=%.1f  (better by %.1f%%)", e1, e2, better)
	s.expected(id, part1, part2)
}

func init() {
	register("091", (*set).gen091)
	register("092", (*set).gen092)
	register("093", (*set).gen093)
	register("094", (*set).gen094)
	register("095", (*set).gen095)
	register("096", (*set).gen096)
	register("097", (*set).gen097)
	register("098", (*set).gen098)
	register("099", (*set).gen099)
	register("100", (*set).gen100)
}
