package main

import (
	"fmt"
)

// Part1 разбирает ELF64-заголовок файла omega.elf и возвращает
// "ELF64 LE  type=ET_DYN  machine=x86-64  entry=0x401000  shdr=13".
func Part1() string {
	panic("TODO")
}

// Part2 находит секции .rodata и .text через .shstrtab и возвращает
// флаг из .rodata и первые 8 байт кода .text.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
