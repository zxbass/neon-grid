package main

import (
	"fmt"
	"io"
	"os"

	"neon-grid/solutions/kit"
)

const (
	headSize = 16
	bootSize = 512
)

func readBootInBuffer(buf []byte) error {
	dataFilePath := kit.Data("001", "boot.bin")
	dataFile, err := os.Open(dataFilePath)
	if err != nil {
		return fmt.Errorf("error while opening file: %v", err)
	}
	defer dataFile.Close()

	_, err = io.ReadFull(dataFile, buf)
	if err != nil {
		return fmt.Errorf("error while reading data: %v", err)
	}

	return nil
}

// Part1 возвращает первые 16 байт boot.bin как строку hex (заглавные,
// через пробел). Миссия: tasks/001-pervyi-dzhek-in.md.
func Part1() string {
	dataHead := make([]byte, headSize)
	err := readBootInBuffer(dataHead)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	return fmt.Sprintf("% X", dataHead)
}

// Part2 возвращает проверку загрузчика: "BOOT OK" и контрольную сумму
// всех байт boot.bin.
func Part2() string {
	data := make([]byte, bootSize)
	err := readBootInBuffer(data)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sum byte
	for i := range data {
		sum += data[i]
	}

	bootStatus := "BAD"
	if data[510] == 0x55 && data[511] == 0xAA {
		bootStatus = "OK"
	}

	return fmt.Sprintf("BOOT %s\nSUM: 0x%X", bootStatus, sum)
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
