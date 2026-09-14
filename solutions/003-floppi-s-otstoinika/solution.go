package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"neon-grid/solutions/kit"
)

const (
	DUMP_LINE_SIZE = 16
)

func buildAsciiFromBytes(data []byte) string {
	var sb strings.Builder

	for _, b := range data {
		if b < 0x20 || b > 0x7e {
			sb.WriteRune('.')
			continue
		}
		sb.WriteByte(b)
	}

	return sb.String()
}

func dumpFile(df *os.File, filter string) string {
	var sb strings.Builder
	buf := make([]byte, DUMP_LINE_SIZE)
	reader := bufio.NewReader(df)
	address := 0
	filteredOut := false
	for ; ; address += DUMP_LINE_SIZE {
		n, err := io.ReadFull(reader, buf)
		if err != nil && n == 0 {
			break
		}
		s := buildAsciiFromBytes(buf[:n])

		if len(filter) > 0 && !strings.Contains(s, filter) {
			filteredOut = true
			continue
		}

		if filteredOut {
			fmt.Fprint(&sb, "....\n")
		}

		halfN := min(n, 8)

		fmt.Fprintf(&sb, "%08x  % x  % x |%s|\n", address, buf[:halfN], buf[halfN:n], s)

		filteredOut = false
	}

	return sb.String()
}

// Part1 возвращает полный hex-дамп disk.img.
func Part1() string {
	fd, err := os.Open(kit.Data("003", "disk.img"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	defer fd.Close()
	return strings.TrimSpace(dumpFile(fd, ""))
}

// Part2 возвращает отфильтрованный дамп: только строки, содержащие
// "GRID", с подавлением повторов (строка "*").
func Part2() string {
	fd, err := os.Open(kit.Data("003", "disk.img"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	defer fd.Close()
	return strings.TrimSpace(dumpFile(fd, "GRID"))
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
