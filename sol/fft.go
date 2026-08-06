package sol

import "math"

// FFT computes the radix-2 Cooley-Tukey FFT of x. The input length must be
// a power of two (missions 123, and the signal analysis tooling around it).
func FFT(x []float64) []complex128 {
	n := len(x)
	if n == 0 || n&(n-1) != 0 {
		panic("FFT: length must be a power of two")
	}
	y := make([]complex128, n)
	for i, v := range x {
		y[i] = complex(v, 0)
	}
	for i, j := 1, 0; i < n; i++ {
		bit := n >> 1
		for ; j&bit != 0; bit >>= 1 {
			j ^= bit
		}
		j ^= bit
		if i < j {
			y[i], y[j] = y[j], y[i]
		}
	}
	for size := 2; size <= n; size <<= 1 {
		ang := -2 * math.Pi / float64(size)
		w := complex(math.Cos(ang), math.Sin(ang))
		for i := 0; i < n; i += size {
			wk := complex(1, 0)
			for k := 0; k < size/2; k++ {
				a := y[i+k]
				b := y[i+k+size/2] * wk
				y[i+k] = a + b
				y[i+k+size/2] = a - b
				wk *= w
			}
		}
	}
	return y
}

// Magnitudes returns |y[k]| for each bin.
func Magnitudes(y []complex128) []float64 {
	out := make([]float64, len(y))
	for i, v := range y {
		out[i] = math.Hypot(real(v), imag(v))
	}
	return out
}

// PeakBin returns the index k (>=1) with the largest magnitude, ignoring
// the DC bin 0. Used to locate a tone's spectral peak.
func PeakBin(mag []float64) int {
	best := 1
	for k := 2; k < len(mag); k++ {
		if mag[k] > mag[best] {
			best = k
		}
	}
	return best
}
