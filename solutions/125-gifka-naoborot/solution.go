package main

import (
	"fmt"
)

// Part1 строит одиночный GIF87a из pixels.bin и проверяет его декодером:
// "GIF87a 32x24 colors=16", "bytes=0x...", "ROUNDTRIP: OK".
func Part1() string {
	panic("TODO")
}

// Part2 строит из frames.bin полную и диффовую анимацию и сравнивает размеры:
// "FULL: N bytes", "DIFF: M bytes   <- экономнее в X раза".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
