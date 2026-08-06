package main

import (
	"fmt"
)

// Part1 читает WAL-журнал wal.bin, применяет валидные записи к KV-стору
// и возвращает статистику и финальное состояние.
func Part1() string {
	panic("TODO")
}

// Part2 обрабатывает усечённый хвост журнала (краш), пишет snapshot.bin
// и возвращает количество применённых/пропущенных записей.
func Part2() string {
	panic("TODO")
}

func main() {
    fmt.Println("=== Part 1 ===")
    fmt.Println(Part1())
    fmt.Println("=== Part 2 ===")
    fmt.Println(Part2())
}
