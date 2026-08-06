package main

import (
	"fmt"
)

// Part1 распаковывает RLE-поток compressed.bin (пары count u16 LE + байт) и
// возвращает "LEN: <длина>\nDATA: <hex первых 64 байт>".
func Part1() string {
	panic("TODO")
}

// Part2 сжимает распакованные данные обратно и проверяет round-trip.
// Возвращает "PACKED: <hex упакованного потока>\nROUNDTRIP OK".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
