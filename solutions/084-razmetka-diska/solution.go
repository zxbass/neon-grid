package main

import (
	"fmt"
)

// Part1 читает MBR из disk.img и возвращает таблицу партиций
// ("PT 0: type=0x0C (FAT32 LBA) start=2048 ...").
func Part1() string {
	panic("TODO")
}

// Part2 разбирает GPT-заголовок в LBA1 и возвращает диапазон
// партиции MERCURY.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
