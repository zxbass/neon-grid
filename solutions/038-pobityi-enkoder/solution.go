package main

import (
	"fmt"
)

// Part1 разбирает кадры из frame.bin (0x7E | len u16 LE | данные | XOR |
// 0x7E) и возвращает "FRAME OK" или "FRAME BAD" для каждого.
func Part1() string {
	panic("TODO")
}

// Part2 обрабатывает кадры из frames.bin с «лишним» байтом 0xFF в конце и
// возвращает "FIXED: <данные>", "OK: <данные>" или "CORRUPT: <hex>".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
