package main

import (
	"fmt"
)

// Part1 читает заголовки из packet.bin (3 бита version, 5 type, 8 flags,
// 16 length, MSB-first) и возвращает строки "VERSION=.. TYPE=.. FLAGS=..
// LENGTH=..".
func Part1() string {
	panic("TODO")
}

// Part2 записывает прочитанные поля обратно и возвращает строки
// "WRITTEN: <hex байт>" для каждого заголовка.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
