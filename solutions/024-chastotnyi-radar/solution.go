package main

import (
	"fmt"
)

// Part1 читает cipher.txt (длинный шифротекст A–Z), считает частоты символов
// и возвращает символы в порядке убывания частоты (при равенстве — по алфавиту):
// "MOST COMMON: <символы через пробел>".
func Part1() string {
	panic("TODO")
}

// Part2 расшифровывает cipher.txt методом частотного анализа (по эталону +
// эвристика слов-якорей THE/AND/ING) и возвращает "DECRYPTED: <текст>".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
