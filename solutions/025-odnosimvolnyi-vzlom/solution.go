package main

import (
	"fmt"
)

// Part1 перебирает все 256 однобайтовых ключей для cipher.bin, выбирает
// ключ с максимальной долей печатных ASCII + пробел/перевод строки и
// возвращает "KEY: 0x%02X\nTEXT: <первые 200 символов>".
func Part1() string {
	panic("TODO")
}

// Part2 читает cipher8.bin (повторяющийся ключ длины 8), подбирает байт
// ключа для каждой колонки i mod 8 по максимальной доле печатных символов
// и возвращает "KEY: <ключ>\nTEXT: <полный текст>".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
