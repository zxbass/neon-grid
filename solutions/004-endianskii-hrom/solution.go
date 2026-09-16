package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"neon-grid/solutions/kit"
)

// Part1 читает nums.bin как little-endian u32 и возвращает строки
// "value=<n> BE=0x<hex>".
func Part1() string {
	bytes, err := os.ReadFile(kit.Data("004", "nums.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder

	numCount := len(bytes) / 4
	for i := range numCount {
		num := binary.LittleEndian.Uint32(bytes[i*4 : i*4+4])
		fmt.Fprintf(&sb, "value=%-5d BE=0x%08X\n", num, num)
	}

	return strings.TrimSpace(sb.String())
}

// Part2 возвращает развёрнутые ASCII-подписи узлов из nums.bin
// (слова, где все 4 байта печатные).
func Part2() string {
	bytes, err := os.ReadFile(kit.Data("004", "nums.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder

	numCount := len(bytes) / 4
outer:
	for i := range numCount {
		st := i * 4
		for k := range 4 {
			if bytes[st+k] < 0x20 || bytes[st+k] > 0x7e {
				continue outer
			}
		}
		sb.WriteByte(bytes[st+3])
		sb.WriteByte(bytes[st+2])
		sb.WriteByte(bytes[st+1])
		sb.WriteByte(bytes[st])
		sb.WriteByte('\n')
	}

	return strings.TrimSpace(sb.String())
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
