package main

import (
	"fmt"
)

// Part1 генерирует 64 значения состояния 16-битного LFSR (отводы битов
// 0,2,3,5) от начального состояния 0xACE1 и возвращает каждое как
// "0x%04X", по одному на строку.
func Part1() string {
	panic("TODO")
}

// Part2 читает emitted.txt (байты выдачи — старший байт состояния после
// сдвига), перебором находит начальное состояние LFSR и возвращает
// следующие 5 байт выдачи: "NEXT: <hex через пробел>".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
