package main

import (
	"encoding/binary"
	"fmt"
	"neon-grid/solutions/kit"
	"os"
)

func rotl32(x uint32, n int) uint32 {
	return (x << n) | (x >> (32 - n))
}

// - acc = 0x5A5A5A5A
func seal(buf []byte) uint32 {
	var acc uint32 = 0x5a5a5a5a

	for _, b := range buf {
		acc = rotl32(acc, 5) ^ uint32(b)
	}

	return acc
}

// Part1 читает cargo.bin и возвращает печать "SEAL: 0x<hex>".
func Part1() string {
	cargo, err := os.ReadFile(kit.Data("009", "cargo.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	return fmt.Sprintf("SEAL: 0x%08X", seal(cargo))
}

// Part2 проверяет печати грузов в batch.bin и возвращает
// "SEAL BAD (expected ..., got ...)".
func Part2() string {
	batch, err := os.ReadFile(kit.Data("009", "batch.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	batchSeal := binary.LittleEndian.Uint32(batch)
	realSeal := seal(batch[4:])

	if batchSeal == realSeal {
		return fmt.Sprintf("SEAL OK (0x%08X)", realSeal)
	}
	return fmt.Sprintf("SEAL BAD (expected 0x%08X, got 0x%08X)", batchSeal, realSeal)
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
