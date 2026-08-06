package sol

// XorBytes XORs two byte slices element-wise (must be equal length).
// Used by mission 029 (P1^P2 = C1^C2).
func XorBytes(a, b []byte) []byte {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]byte, n)
	for i := 0; i < n; i++ {
		out[i] = a[i] ^ b[i]
	}
	return out
}
