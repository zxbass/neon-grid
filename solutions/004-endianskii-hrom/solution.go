package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"neon-grid/solutions/kit"
)

func readNums() ([]byte, error) {
	df, err := os.Open(kit.Data("004", "nums.bin"))
	if err != nil {
		return nil, err
	}

	bytes, err := io.ReadAll(df)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Part1 читает nums.bin как little-endian u32 и возвращает строки
// "value=<n> BE=0x<hex>".
func Part1() string {
	bytes, err := readNums()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder

	sizeIn32Bits := len(bytes) / 4
	for i := range sizeIn32Bits {
		num := binary.LittleEndian.Uint32(bytes[i*4:])
		fmt.Fprintf(&sb, "value=%-5d BE=0x%08X\n", num, num)
	}

	return strings.TrimSpace(sb.String())
}

// Part2 возвращает развёрнутые ASCII-подписи узлов из nums.bin
// (слова, где все 4 байта печатные).
func Part2() string {
	panic("TODO")
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
