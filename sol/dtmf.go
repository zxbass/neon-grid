package sol

import "math"

// DTMF table (mission 123): digit -> (low tone, high tone) in Hz.
var dtmfTable = map[byte][2]float64{
	'1': {697, 1209}, '2': {697, 1336}, '3': {697, 1477}, 'A': {697, 1633},
	'4': {770, 1209}, '5': {770, 1336}, '6': {770, 1477}, 'B': {770, 1633},
	'7': {852, 1209}, '8': {852, 1336}, '9': {852, 1477}, 'C': {852, 1633},
	'*': {941, 1209}, '0': {941, 1336}, '#': {941, 1477}, 'D': {941, 1633},
}

var dtmfLow = []float64{697, 770, 852, 941}
var dtmfHigh = []float64{1209, 1336, 1477, 1633}
var dtmfGrid = [][]byte{
	{'1', '2', '3', 'A'},
	{'4', '5', '6', 'B'},
	{'7', '8', '9', 'C'},
	{'*', '0', '#', 'D'},
}

// SynthDTMF produces a mono signal: each digit lasts msPerDigit milliseconds
// at sample rate sr, as the sum of its two DTMF tones.
func SynthDTMF(digits string, sr int, msPerDigit int) []float64 {
	per := sr * msPerDigit / 1000
	out := make([]float64, 0, len(digits)*per)
	for i := 0; i < len(digits); i++ {
		t := dtmfTable[digits[i]]
		for n := 0; n < per; n++ {
			x := float64(n) / float64(sr)
			out = append(out, math.Sin(2*math.Pi*t[0]*x)+math.Sin(2*math.Pi*t[1]*x))
		}
	}
	return out
}

// DecodeDTMF recovers digits from a SynthDTMF-style signal, taking a
// 256-sample window from the middle of each digit slot.
func DecodeDTMF(samples []float64, sr int, msPerDigit int) string {
	const win = 256
	per := sr * msPerDigit / 1000
	digits := ""
	for start := 0; start+per <= len(samples); start += per {
		mid := start + per/2 - win/2
		w := make([]float64, win)
		copy(w, samples[mid:mid+win])
		mag := Magnitudes(FFT(w))
		low := 0.0
		hi := 0.0
		loBin, hiBin := 0, 0
		for k := 1; k < win/2; k++ {
			f := float64(k) * float64(sr) / float64(win)
			if f < 1000 {
				if mag[k] > low {
					low = mag[k]
					loBin = k
				}
			} else if f < 2000 {
				if mag[k] > hi {
					hi = mag[k]
					hiBin = k
				}
			}
		}
		lf := float64(loBin) * float64(sr) / float64(win)
		hf := float64(hiBin) * float64(sr) / float64(win)
		digits += dtmfFind(lf, hf)
	}
	return digits
}

func dtmfFind(low, high float64) string {
	li, hi := -1, -1
	lb, hb := 1000.0, 1000.0
	for i, t := range dtmfLow {
		if d := math.Abs(t - low); d < lb {
			lb, li = d, i
		}
	}
	for i, t := range dtmfHigh {
		if d := math.Abs(t - high); d < hb {
			hb, hi = d, i
		}
	}
	if li < 0 || hi < 0 || lb > 40 || hb > 40 {
		return "?"
	}
	return string(dtmfGrid[li][hi])
}
