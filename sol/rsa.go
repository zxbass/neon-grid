package sol

// ModPow is binary exponentiation: a^b mod m (mission 030).
func ModPow(a, b, m int) int {
	a %= m
	if a < 0 {
		a += m
	}
	r := 1 % m
	for b > 0 {
		if b&1 == 1 {
			r = (r * a) % m
		}
		a = (a * a) % m
		b >>= 1
	}
	return r
}

// ModInv returns the multiplicative inverse of a modulo m via the extended
// Euclid algorithm (mission 030).
func ModInv(a, m int) int {
	a %= m
	if a < 0 {
		a += m
	}
	t, newT := 0, 1
	r, newR := m, a
	for newR != 0 {
		q := r / newR
		t, newT = newT, t-q*newT
		r, newR = newR, r-q*newR
	}
	if r != 1 {
		return 0
	}
	if t < 0 {
		t += m
	}
	return t
}

// FactorSmall returns (p, q) with p <= q, p*q = n, found by trial division.
func FactorSmall(n int) (int, int) {
	for p := 2; p*p <= n; p++ {
		if n%p == 0 {
			return p, n / p
		}
	}
	return 1, n
}

// RSABreak recovers plaintext m from ciphertext c given n and e.
func RSABreak(n, e, c int) int {
	p, q := FactorSmall(n)
	phi := (p - 1) * (q - 1)
	d := ModInv(e, phi)
	return ModPow(c, d, n)
}
