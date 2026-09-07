package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		listMissions()
		return
	}

	switch os.Args[1] {
	case "list", "ls":
		listMissions()
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: neon run <NNN>")
			os.Exit(1)
		}
		runMission(os.Args[2])
	case "test":
		if len(os.Args) < 3 {
			fmt.Println("Usage: neon test <NNN>")
			os.Exit(1)
		}
		testMission(os.Args[2])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Commands: list, run <NNN>, test <NNN>")
	}
}

func banner() {
	fmt.Print(`
   ███╗   ██╗███████╗ ██████╗ ███╗   ██╗
   ████╗  ██║██╔════╝██╔═══██╗████╗  ██║
   ██╔██╗ ██║█████╗  ██║   ██║██╔██╗ ██║
   ██║╚██╗██║██╔══╝  ██║   ██║██║╚██╗██║
   ██║ ╚████║███████╗╚██████╔╝██║ ╚████║
   ╚═╝  ╚═══╝╚══════╝ ╚═════╝ ╚═╝  ╚═══╝
       GRID // 2049 // КАРТА ВЗЛОМА
`)
}

type mission struct {
	num  string
	name string
}

func listMissions() {
	banner()
	fmt.Println(" 200 миссий для нетраннера")
	fmt.Println(strings.Repeat("─", 56))
	fmt.Println()

	categories := []struct {
		title string
		range_ string
	}{
		{"Бинарные файлы и байты", "001-010"},
		{"Форматы файлов", "011-020"},
		{"Криптография", "021-030"},
		{"Кодирование и сжатие", "031-040"},
		{"Сокеты и сети, база", "041-050"},
		{"Сети, глубокий уровень", "051-060"},
		{"TUI-интерфейсы", "061-070"},
		{"Параллельность и асинхронность", "071-080"},
		{"Файловые системы и форензика", "081-090"},
		{"Мультимедиа и сигналы", "091-100"},
		{"Реверс-инжиниринг", "101-110"},
		{"Финальный взлом", "111-120"},
		{"После финала", "121-130"},
		{"Крипто-атаки", "131-140"},
		{"Форматы, том II", "141-150"},
		{"Внутри ОС", "151-160"},
		{"Алгоритмы", "161-170"},
		{"Терминал-арт", "171-180"},
		{"Сети: пакеты и снифферы", "181-190"},
		{"Артефакты Сетки", "191-200"},
	}

	for _, cat := range categories {
		fmt.Printf(" \033[36m%s\033[0m  [%s]\n", cat.title, cat.range_)
	}

	fmt.Println()
	fmt.Println(" \033[33mUsage:\033[0m")
	fmt.Println("   neon list              показать все миссии")
	fmt.Println("   neon run <NNN>         запустить миссию")
	fmt.Println("   neon test <NNN>        протестировать миссию")
	fmt.Println()
	fmt.Println(" \033[90mПример: go run ./cmd/neon run 001\033[0m")
	fmt.Println()
}

func findMissionDir(num string) string {
	padded := fmt.Sprintf("%03s", num)
	if len(num) == 1 {
		padded = "00" + num
	} else if len(num) == 2 {
		padded = "0" + num
	} else {
		padded = num
	}

	entries, err := os.ReadDir("solutions")
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), padded+"-") {
			return filepath.Join("solutions", e.Name())
		}
	}
	// Try just number prefix
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), num+"-") {
			return filepath.Join("solutions", e.Name())
		}
	}
	return ""
}

func runMission(num string) {
	dir := findMissionDir(num)
	if dir == "" {
		fmt.Printf("Mission %s not found\n", num)
		os.Exit(1)
	}
	banner()
	fmt.Printf(" ▶ Running mission %s...\n\n", num)
	cmd := exec.Command("go", "run", "./"+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n ✗ Mission %s failed: %v\n", num, err)
		os.Exit(1)
	}
}

func testMission(num string) {
	dir := findMissionDir(num)
	if dir == "" {
		fmt.Printf("Mission %s not found\n", num)
		os.Exit(1)
	}
	banner()
	fmt.Printf(" ▶ Testing mission %s...\n\n", num)
	cmd := exec.Command("go", "test", "-v", "./"+dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("\n ✗ Mission %s tests failed: %v\n", num, err)
		os.Exit(1)
	}
}
