package main

import (
	"fmt"
)

// Part1 читает tunnel.bin, разбирает фрагменты по заголовкам и склеивает:
// "SENT 4 fragments (len=4096)", "REASSEMBLED: 4096 bytes, order OK".
func Part1() string {
	panic("TODO")
}

// Part2 восстанавливает reordered.bin (перемешанные фрагменты с XOR) и
// возвращает "REORDERED: ..." и "TUNNEL OK (XOR+CRC verified)".
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
