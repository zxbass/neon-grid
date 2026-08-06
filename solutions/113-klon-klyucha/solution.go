package main

import (
	"fmt"
)

// Part1 читает intercept.bin (t, counter, CRC16CCITT) и восстанавливает ключ:
// "KEY(t=1700000000, c=7): ..." и "GEN OK: roundtrip".
func Part1() string {
	panic("TODO")
}

// Part2 проверяет соответствие перехвата по текущему времени и предсказывает
// следующий ключ: "MATCH: ...", "NEXT: ...".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
