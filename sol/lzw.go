package sol

// LZWEncode compresses data to a list of codes. Initial dictionary: 0-255.
// First new entry gets code 256. maxDict bounds the dictionary (4096 =
// 12-bit, mission 034).
func LZWEncode(data []byte, maxDict int) []int {
	dict := make(map[string]int, maxDict)
	for i := 0; i < 256; i++ {
		dict[string([]byte{byte(i)})] = i
	}
	next := 256
	var codes []int
	w := []byte{}
	for _, b := range data {
		ws := append(append([]byte{}, w...), b)
		if _, ok := dict[string(ws)]; ok {
			w = ws
			continue
		}
		codes = append(codes, dict[string(w)])
		if next < maxDict {
			dict[string(ws)] = next
			next++
		}
		w = []byte{b}
	}
	if len(w) > 0 {
		codes = append(codes, dict[string(w)])
	}
	return codes
}

// LZWDecode reverses LZWEncode. Handles the classic "code not yet in
// dictionary" case: entry = prev + prev[0].
func LZWDecode(codes []int) []byte {
	if len(codes) == 0 {
		return nil
	}
	dict := make(map[int][]byte)
	for i := 0; i < 256; i++ {
		dict[i] = []byte{byte(i)}
	}
	next := 256
	prev := dict[codes[0]]
	out := append([]byte{}, prev...)
	for _, c := range codes[1:] {
		var entry []byte
		if e, ok := dict[c]; ok {
			entry = e
		} else if c == next {
			entry = append(append([]byte{}, prev...), prev[0])
		} else {
			return nil
		}
		out = append(out, entry...)
		dict[next] = append(append([]byte{}, prev...), entry[0])
		next++
		prev = entry
	}
	return out
}
