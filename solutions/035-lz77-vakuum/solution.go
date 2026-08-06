package main

import (
	"fmt"
)

// Part1 распаковывает LZ77-поток packed.bin (флаги, литералы, пары
// offset/length) и возвращает "DECODED: <текст>".
func Part1() string {
	panic("TODO")
}

// Part2 сжимает raw.bin жадным LZ77-компрессором, возвращает
// "PACKED: <hex потока>\nROUNDTRIP OK (input=<длина> output=<длина>)".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
