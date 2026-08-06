package main

import (
	"fmt"
)

// Part1 читает строки из input.txt и возвращает "NEON: 0x..." и
// "GRID: 0x..." — CRC-32 IEEE для каждой строки.
func Part1() string {
	panic("TODO")
}

// Part2 считает CRC-32 табличным методом для blob.bin и его повреждённой
// копии, сверяя с побитовым. Возвращает "TABLE CRC: 0x...\nBITWISE MATCH:
// OK\nCORRUPTED CRC: 0x...".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
