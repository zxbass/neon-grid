package main

import (
	"fmt"
)

// Part1 читает cipher.bin (поток RC4-подобного шифра, ключ MERCURY) и
// возвращает расшифрованный текст: "PLAINTEXT: <текст>".
func Part1() string {
	panic("TODO")
}

// Part2 читает frames.bin (3 фрейма по 4 байта заголовка; если байт 0
// после XOR с текущим байтом потока равен 0x80, поток сбрасывается) и
// возвращает содержимое фреймов после заголовка:
// "FRAME 0: ...\nFRAME 1: ...\nFRAME 2: ...".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
