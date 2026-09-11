package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"neon-grid/solutions/kit"
)

// reads maximun lineCount lines from a text file, starting with startLine.
// if startLine = 0, it means from the beginning, lines count from 0.
func readFromTextFile(path string, startLine, lineCount int) ([]string, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	if startLine < 0 {
		startLine = 0
	}

	lines := make([]string, 0, lineCount)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		if startLine > 0 {
			startLine--
			continue
		}

		if lineCount == 0 {
			break
		}
		lineCount--

		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func scanInputBytes() ([]byte, error) {
	byteLines, err := readFromTextFile(kit.Data("002", "input.txt"), 0, 1)
	if err != nil {
		return nil, err
	}

	var byteLine string
	if len(byteLines) > 0 {
		byteLine = byteLines[0]
	}

	parts := strings.Fields(byteLine)
	bytes := make([]byte, 0, len(parts))

	for _, part := range parts {
		b, err := strconv.ParseUint(part, 16, 8)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, byte(b))
	}

	return bytes, nil
}

// Part1 декодирует байты из первой строки input.txt в строки вида
// "SIZE=42 FAST=no IMPORT=yes".
func Part1() string {
	_, err := scanInputBytes()
	if err != nil {
		log.Fatal(err)
	}

	return "foo"
}

// Part2 упаковывает тройки "SIZE FAST IMPORT" из input.txt в hex-байт
// или возвращает SIZE_OVERFLOW.
func Part2() string {
	panic("TODO")
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
