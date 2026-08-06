package sol

// RunCPU emulates the tiny 8-bit CPU "Плесень-0" (mission 121).
// Opcodes: 0x00 HLT, 0x01 LDA #imm, 0x02 ADD #imm, 0x03 SUB #imm,
// 0x04 STA addr, 0x05 JMP addr, 0x06 JZ addr, 0x07 PRN, 0x08 INP addr.
// Instructions are variable-width: HLT and PRN take no operand (advance 1),
// the rest take one operand byte (advance 2).
func RunCPU(prog []byte) string {
	pc := 0
	var a byte
	mem := make([]byte, 256)
	out := make([]byte, 0, 16)
	opWidth := map[byte]int{0x00: 1, 0x01: 2, 0x02: 2, 0x03: 2, 0x04: 2, 0x05: 2, 0x06: 2, 0x07: 1, 0x08: 2}
	for steps := 0; pc < len(prog) && steps < 10000; steps++ {
		op := prog[pc]
		if op == 0x00 {
			return string(out)
		}
		w := opWidth[op]
		if w == 0 || pc+w > len(prog) {
			return string(out)
		}
		var opd byte
		if w == 2 {
			opd = prog[pc+1]
		}
		switch op {
		case 0x01:
			a = opd
		case 0x02:
			a += opd
		case 0x03:
			a -= opd
		case 0x04:
			mem[opd] = a
		case 0x05:
			pc = int(opd)
			continue
		case 0x06:
			if a == 0 {
				pc = int(opd)
				continue
			}
		case 0x07:
			out = append(out, a)
		case 0x08:
			a = mem[opd]
		}
		pc += w
	}
	return string(out)
}
