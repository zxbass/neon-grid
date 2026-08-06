package main

import (
	"fmt"
)

// Part1 разбирает DOS и PE-заголовок файла camera.exe и возвращает
// "PE32 x86  sections=5  entry=0x401000  subsys=GUI".
func Part1() string {
	panic("TODO")
}

// Part2 сканирует секцию .rdata на печатные строки и находит пароль
// вида pw_*/pass*.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
