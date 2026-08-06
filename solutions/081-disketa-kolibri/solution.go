package main

import (
	"fmt"
)

// Part1 разбирает BPB образа floppy.img и возвращает строку параметров
// "FAT12 1440KB  bytes/sector=512  clusters=2847  root=224".
func Part1() string {
	panic("TODO")
}

// Part2 находит в корневом каталоге SECRET.TXT, проходит цепочку кластеров
// через FAT12 и возвращает его содержимое.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
