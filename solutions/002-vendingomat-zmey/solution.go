package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"neon-grid/solutions/kit"
)

const (
	SIZE_MASK  = 0b111111
	FAST_MASK  = 0x40
	IMPORT_BIT = 0x80
)

type Triplet struct {
	Size, Fast, Imp uint64
}

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

	var lines []string
	if lineCount < 0 {
		lines = make([]string, 0, 4)
	} else {
		lines = make([]string, 0, lineCount)
	}
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

func scanTriplets() ([]Triplet, error) {
	lines, err := readFromTextFile(kit.Data("002", "input.txt"), 1, -1)
	if err != nil {
		return nil, err
	}

	triplets := make([]Triplet, 0, len(lines))

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 3 {
			return nil, errors.New("data triplet is wrong - must be 3 parts")
		}

		size, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing size column: %v", err)
		}
		fast, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing fast column: %v", err)
		}
		imp, err := strconv.ParseUint(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing imp column: %v", err)
		}

		triplets = append(triplets, Triplet{Size: size, Fast: fast, Imp: imp})
	}

	return triplets, nil
}

// Part1 декодирует байты из первой строки input.txt в строки вида
// "SIZE=42 FAST=no IMPORT=yes".
func Part1() string {
	bytes, err := scanInputBytes()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var out strings.Builder

	for _, b := range bytes {
		vol := b & SIZE_MASK
		fast := "no"
		if b&FAST_MASK != 0 {
			fast = "yes"
		}
		imp := "no"
		if b&IMPORT_BIT != 0 {
			imp = "yes"
		}
		fmt.Fprintf(&out, "SIZE=%d FAST=%s IMPORT=%s\n", vol, fast, imp)
	}

	return strings.TrimSpace(out.String())
}

// Part2 упаковывает тройки "SIZE FAST IMPORT" из input.txt в hex-байт
// или возвращает SIZE_OVERFLOW.
func Part2() string {
	triplets, err := scanTriplets()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	var out strings.Builder

	for _, triplet := range triplets {
		if triplet.Size > SIZE_MASK || triplet.Fast > 1 || triplet.Imp > 1 {
			fmt.Fprintf(&out, "SIZE_OVERFLOW\n")
			continue
		}
		packed := (triplet.Imp << 7) + (triplet.Fast << 6) + triplet.Size
		fmt.Fprintf(&out, "0x%X\n", packed)
	}
	return strings.TrimSpace(out.String())
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
