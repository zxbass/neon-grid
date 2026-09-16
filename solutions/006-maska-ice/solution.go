package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"neon-grid/solutions/kit"
)

const (
	Read = iota
	Write
	Trace
	Break
	Inject
	Tunnel
	Scan
)

const PassMask uint32 = 1<<Read | 1<<Trace | 1<<Tunnel

var BitNames = [7]string{
	Read:   "READ",
	Write:  "WRITE",
	Trace:  "TRACE",
	Break:  "BREAK",
	Inject: "INJECT",
	Tunnel: "TUNNEL",
	Scan:   "SCAN",
}

func readInput() ([]uint32, error) {
	data, err := os.ReadFile(kit.Data("006", "input.txt"))
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(string(data))
	nums := make([]uint32, 0, len(fields))

	for _, sn := range fields {
		n, err := strconv.ParseUint(sn, 16, 32)
		if err != nil {
			return nil, err
		}

		nums = append(nums, uint32(n))
	}

	return nums, nil
}

// Part1 декодирует 32-битные слова флагов из input.txt в списки прав
// вида "00000013: READ WRITE INJECT".
func Part1() string {
	nums, err := readInput()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder
	for _, n := range nums {
		fmt.Fprintf(&sb, "%08x:", n)
		for i := range 32 {
			if (n>>i)&1 == 1 {
				sb.WriteByte(' ')
				if i < len(BitNames) {
					sb.WriteString(BitNames[i])
				} else {
					fmt.Fprintf(&sb, "bit%d(reserved)", i)
				}
			}
		}
		sb.WriteByte('\n')
	}

	return strings.TrimSpace(sb.String())
}

// Part2 возвращает ключ от "парадной двери" ICE.
func Part2() string {
	return fmt.Sprintf("KEY: 0x%x", PassMask)
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
