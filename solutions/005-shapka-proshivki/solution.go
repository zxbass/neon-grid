package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"neon-grid/solutions/kit"
)

const (
	headerSize = 64
)

type Version struct {
	Major, Minor byte
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

type Header struct {
	Magic   string
	Ver     Version
	Name    string
	CodeLen uint32
	Crc     uint32
	Api     Version
	Hw      Version
	Desc    string
}

func (hdr Header) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "magic    : %s v%s\n", hdr.Magic, hdr.Ver)
	fmt.Fprintf(&sb, "name     : %s\n", hdr.Name)
	fmt.Fprintf(&sb, "code_len : %d\n", hdr.CodeLen)
	fmt.Fprintf(&sb, "crc      : 0x%X\n", hdr.Crc)
	fmt.Fprintf(&sb, "api      : %s\n", hdr.Api)
	fmt.Fprintf(&sb, "hw       : %s\n", hdr.Hw)
	fmt.Fprintf(&sb, "desc     : %s", hdr.Desc)

	return sb.String()
}

func readCString(buf []byte) string {
	nulIdx := bytes.IndexByte(buf, 0)

	if nulIdx == -1 {
		return string(buf)
	}

	return string(buf[:nulIdx])
}

func parseHeader(buf []byte) (*Header, error) {
	if len(buf) < headerSize {
		return nil, fmt.Errorf("заголовок короче %d байт", headerSize)
	}
	hdr := &Header{}
	hdr.Magic = string(buf[:2])
	hdr.Ver = Version{buf[2], buf[3]}
	hdr.Name = readCString(buf[4:12])
	hdr.CodeLen = binary.LittleEndian.Uint32(buf[12:16])
	hdr.Crc = binary.LittleEndian.Uint32(buf[16:20])
	hdr.Api = Version{buf[21], buf[20]}
	hdr.Hw = Version{buf[23], buf[22]}
	hdr.Desc = readCString(buf[24:64])

	return hdr, nil
}

// Part1 разбирает 64-байтный заголовок firmware.bin и возвращает
// строки "magic    : FW v1.4" и т.д.
func Part1() string {
	bin, err := os.ReadFile(kit.Data("005", "firmware.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	hdr, err := parseHeader(bin)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	if hdr.Magic != "FW" {
		return "BAD_MAGIC"
	}

	return hdr.String()
}

// Part2 сверяет контрольную сумму кода прошивки с заголовком.
func Part2() string {
	bin, err := os.ReadFile(kit.Data("005", "firmware.bin"))
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	hdr, err := parseHeader(bin)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder

	data := bin[64:]
	if len(data) < int(hdr.CodeLen) {
		hdr.CodeLen = uint32(len(data))
		fmt.Fprint(&sb, "TRUNCATED\n")
	}
	data = data[:hdr.CodeLen]
	var crc uint64
	for i := range hdr.CodeLen / 4 {
		crc += uint64(binary.LittleEndian.Uint32(data[i*4 : i*4+4]))
		crc %= 0xFFFFFFFF
	}
	if uint32(crc) == hdr.Crc {
		fmt.Fprintf(&sb, "CRC OK (0x%X)", crc)
	} else {
		fmt.Fprintf(&sb, "CRC MISMATCH (header: 0x%X data: 0x%X)", hdr.Crc, crc)
	}

	return sb.String()
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
