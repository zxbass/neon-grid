package main

import (
	"fmt"
)

// Part1 читает input.txt (строки "seed: <ASCII>" и "n: <число>") и
// возвращает токен t[n] хэш-цепочки: "TOKEN: <64 hex>".
func Part1() string {
	panic("TODO")
}

// Part2 читает input.txt (строки "known:", "steps:" и "expected:"),
// проверяет цепочку прогоном steps хэшей от known и возвращает
// "CHAIN OK\nORIGIN: <64 hex>" (или "CHAIN BROKEN").
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
