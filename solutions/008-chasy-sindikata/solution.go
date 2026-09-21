package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"neon-grid/solutions/kit"
)

// readInputLine читает из input.txt строку с номером line (с нуля)
// и разбирает её hex-значения как 32-битные числа.
func readInputLine(line int) ([]uint32, error) {
	fd, err := os.Open(kit.Data("008", "input.txt"))
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	scn := bufio.NewScanner(fd)
	for range line {
		if !scn.Scan() {
			if err := scn.Err(); err != nil {
				return nil, fmt.Errorf("пропуск строки: %w", err)
			}
			return nil, fmt.Errorf("в файле нет строки #%d", line)
		}
	}

	if !scn.Scan() {
		if err := scn.Err(); err != nil {
			return nil, fmt.Errorf("чтение строки #%d: %w", line, err)
		}
		return nil, fmt.Errorf("в файле нет строки #%d", line)
	}

	fields := strings.Fields(scn.Text())
	nums := make([]uint32, 0, len(fields))
	for _, s := range fields {
		n, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("разбор %q: %w", s, err)
		}
		nums = append(nums, uint32(n))
	}

	return nums, nil
}

type DOSTime struct {
	sec, min, hour, day, mon byte
	year                     uint32
}

// parseDOSTimestamp раскладывает 32-битный DOS-таймстамп по битовым полям.
func parseDOSTimestamp(ts uint32) DOSTime {
	return DOSTime{
		sec:  byte(ts&0x1F) * 2,       // биты 0–4: секунды / 2
		min:  byte((ts >> 5) & 0x3F),  // биты 5–10: минуты
		hour: byte((ts >> 11) & 0x1F), // биты 11–15: часы
		day:  byte((ts >> 16) & 0x1F), // биты 16–20: день месяца
		mon:  byte((ts >> 21) & 0x0F), // биты 21–24: месяц
		year: 1980 + (ts>>25)&0x7F,    // биты 25–31: год с 1980
	}
}

func (dt DOSTime) String() string {
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", dt.year, dt.mon, dt.day, dt.hour, dt.min, dt.sec)
}

// Unix возвращает Unix-секунды. time.Date сам учитывает длины месяцев и
// високосные годы, поэтому это надёжнее ручной арифметики.
func (dt DOSTime) Unix() int64 {
	t := time.Date(int(dt.year), time.Month(dt.mon), int(dt.day), int(dt.hour), int(dt.min), int(dt.sec), 0, time.UTC)
	return t.Unix()
}

// Part1 декодирует DOS-таймстампы из первой строки input.txt.
func Part1() string {
	nums, err := readInputLine(0)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	lines := make([]string, 0, len(nums))
	for _, n := range nums {
		lines = append(lines, parseDOSTimestamp(n).String())
	}

	return strings.Join(lines, "\n")
}

// Part2 находит самый свежий таймстамп из второй строки и разницу в секундах.
// Битовые поля идут от младшего (секунды) к старшему (год), поэтому числовой
// максимум упакованного значения — это и есть самый поздний момент.
func Part2() string {
	nums, err := readInputLine(1)
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}
	if len(nums) == 0 {
		return "ERROR: нет таймстампов"
	}

	latest := parseDOSTimestamp(slices.Max(nums))
	earliest := parseDOSTimestamp(slices.Min(nums))

	return fmt.Sprintf("LATEST: %s\nDIFF: %d", latest, latest.Unix()-earliest.Unix())
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
