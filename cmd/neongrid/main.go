package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ---------------------------------------------------------------- app

type app struct {
	root     string
	missions []*Mission
	stats    *Stats

	view     string // list | detail | stats | test
	sel      int    // index в отфильтрованном списке
	scroll   int
	filter   string
	filtered []*Mission

	outputTitle string
	output      []string

	statusLine string
}

func (a *app) visible() []*Mission {
	if a.filter == "" {
		return a.missions
	}
	var out []*Mission
	for _, m := range a.missions {
		needle := strings.ToLower(a.filter)
		if strings.Contains(m.ID, needle) || strings.Contains(strings.ToLower(m.Slug), needle) {
			out = append(out, m)
		}
	}
	return out
}

func (a *app) selected() *Mission {
	if a.sel < 0 || a.sel >= len(a.filtered) {
		return nil
	}
	return a.filtered[a.sel]
}

// ---------------------------------------------------------------- actions

func (a *app) runTest(m *Mission) string {
	a.stats.touch(m.ID)
	cmd := exec.Command("go", "test", "./"+m.Dir, "-count=1")
	cmd.Dir = a.root
	out, err := cmd.CombinedOutput()
	if err == nil {
		a.stats.complete(m.ID)
	}
	a.stats.save()
	return string(out)
}

func (a *app) runMission(m *Mission) string {
	a.stats.touch(m.ID)
	a.stats.save()
	cmd := exec.Command("go", "run", "./"+m.Dir)
	cmd.Dir = a.root
	out, _ := cmd.CombinedOutput()
	return string(out)
}

func (a *app) verifyAll() string {
	a.statusLine = "Проверяю все миссии... (Esc — прервать)"
	results := make([]string, len(a.missions))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	done := 0
	var mu sync.Mutex
	for i, m := range a.missions {
		wg.Add(1)
		go func(i int, m *Mission) {
			defer wg.Done()
			sem <- struct{}{}
			out := a.runTest(m)
			<-sem
			mu.Lock()
			ok := !strings.Contains(out, "FAIL") && strings.Contains(out, "ok ") && !strings.Contains(out, "panic")
			status := "✗"
			if ok {
				status = "✓"
			}
			results[i] = fmt.Sprintf("%s %s %s", status, m.ID, m.Slug)
			done++
			mu.Unlock()
		}(i, m)
	}
	doneCh := make(chan struct{})
	go func() { wg.Wait(); close(doneCh) }()
	for {
		select {
		case <-doneCh:
			a.statusLine = ""
			return "ИТОГ ПРОВЕРКИ:\n" + strings.Join(results, "\n")
		default:
			a.render()
			k, _ := readKey()
			if k == keyEsc {
				a.statusLine = "Проверка прервана"
				return "Проверка прервана пользователем."
			}
		}
	}
}

// ---------------------------------------------------------------- entry

func main() {
	root := findRoot()
	term := os.Stdin
	fd := int(term.Fd())
	if !isTerminal(fd) || !isTerminal(int(os.Stdout.Fd())) {
		plain(root)
		return
	}
	termRaw(fd)
	defer termRestore(fd)
	defer fmt.Print("\x1b[2J\x1b[H\x1b[?25h\x1b[0m")

	a := &app{root: root, view: "list"}
	a.missions = loadMissions(root)
	a.stats = loadStats(root)
	if len(a.missions) == 0 {
		fmt.Println("Миссии не найдены. Запусти из корня проекта neon-grid.")
		return
	}
	a.filtered = a.missions
	fmt.Print("\x1b[?25l")
	a.loop()
}

func (a *app) loop() {
	for {
		a.render()
		k, r := readKey()
		switch a.view {
		case "list":
			a.handleList(k, r)
		case "detail":
			a.handleDetail(k, r)
		case "stats":
			a.handleStats(k, r)
		case "test":
			a.handleTest(k, r)
		}
	}
}

func (a *app) handleList(k key, r rune) {
	switch k {
	case keyUp:
		if a.sel > 0 {
			a.sel--
		}
	case keyDown:
		if a.sel < len(a.filtered)-1 {
			a.sel++
		}
	case keyPgUp:
		a.sel -= 10
		if a.sel < 0 {
			a.sel = 0
		}
	case keyPgDn:
		a.sel += 10
		if a.sel >= len(a.filtered) {
			a.sel = len(a.filtered) - 1
		}
	case keyHome:
		a.sel = 0
	case keyEnd:
		a.sel = len(a.filtered) - 1
	case keyEnter:
		m := a.selected()
		if m == nil {
			return
		}
		a.outputTitle = "ТЕСТ  " + m.ID + " " + m.Slug
		a.output = strings.Split(a.runTest(m), "\n")
		a.scroll = 0
		a.view = "test"
	case keyRune:
		a.handleListRune(r)
	}
}

func (a *app) handleListRune(r rune) {
	switch r {
	case 'r':
		m := a.selected()
		if m == nil {
			return
		}
		a.outputTitle = "ЗАПУСК  " + m.ID + " " + m.Slug
		a.output = strings.Split(a.runMission(m), "\n")
		a.scroll = 0
		a.view = "test"
	case 't':
		m := a.selected()
		if m == nil {
			return
		}
		a.outputTitle = "ЗАДАНИЕ  " + m.ID + " " + m.Slug
		a.output = strings.Split(a.taskText(m), "\n")
		a.scroll = 0
		a.view = "test"
	case 's':
		a.view = "stats"
	case 'v':
		out := a.verifyAll()
		a.outputTitle = "ПРОВЕРКА ВСЕХ МИССИЙ"
		a.output = strings.Split(out, "\n")
		a.scroll = 0
		a.view = "test"
	case '/':
		fmt.Print("\x1b[?25h")
		val, ok := readLine("Фильтр: ")
		fmt.Print("\x1b[?25l")
		if ok {
			a.filter = strings.TrimSpace(val)
			a.filtered = a.visible()
			a.sel = 0
		}
	case 'g':
		fmt.Print("\x1b[?25h")
		val, ok := readLine("Перейти к миссии: ")
		fmt.Print("\x1b[?25l")
		if ok {
			for i, m := range a.filtered {
				if m.ID == strings.TrimSpace(val) {
					a.sel = i
					break
				}
			}
		}
	case 'q':
		a.quit()
	}
}

func (a *app) handleDetail(k key, r rune) {
	switch k {
	case keyEsc, keyTab, keyLeft:
		a.view = "list"
	case keyRune:
		if r == 't' || r == 'q' {
			a.view = "list"
		}
	case keyUp:
		if a.scroll > 0 {
			a.scroll--
		}
	case keyDown:
		a.scroll++
	case keyPgUp:
		a.scroll -= 10
		if a.scroll < 0 {
			a.scroll = 0
		}
	case keyPgDn:
		a.scroll += 10
	}
}

func (a *app) handleStats(k key, r rune) {
	switch k {
	case keyEsc, keyTab, keyLeft:
		a.view = "list"
	case keyRune:
		if r == 's' || r == 'q' {
			a.view = "list"
		}
	case keyUp:
		a.scroll--
		if a.scroll < 0 {
			a.scroll = 0
		}
	case keyDown:
		a.scroll++
	}
}

func (a *app) handleTest(k key, r rune) {
	switch k {
	case keyEsc, keyTab, keyLeft:
		a.view = "list"
	case keyRune:
		if r == 'q' {
			a.view = "list"
		}
	case keyUp:
		if a.scroll > 0 {
			a.scroll--
		}
	case keyDown:
		a.scroll++
	case keyPgUp:
		a.scroll -= 10
		if a.scroll < 0 {
			a.scroll = 0
		}
	case keyPgDn:
		a.scroll += 10
	case keyHome:
		a.scroll = 0
	case keyEnd:
		a.scroll = len(a.output)
	}
}

func (a *app) quit() {
	a.stats.save()
	a.view = "quit"
}

func (a *app) taskText(m *Mission) string {
	b, err := os.ReadFile(a.root + "/tasks/" + m.ID + "-" + m.Slug + ".md")
	if err != nil {
		return "задание не найдено"
	}
	return string(b)
}

// ---------------------------------------------------------------- render

func (a *app) render() {
	w, h := termSize(int(os.Stdout.Fd()))
	if h < 10 {
		return
	}
	var sb strings.Builder
	sb.WriteString("\x1b[2J\x1b[H")
	switch a.view {
	case "list":
		a.renderList(&sb, w, h)
	case "stats":
		a.renderStats(&sb, w, h)
	case "test":
		a.renderOutput(&sb, w, h)
	}
	sb.WriteString("\x1b[?25l")
	fmt.Print(sb.String())
}

func (a *app) renderList(sb *strings.Builder, w, h int) {
	visRows := h - 4
	listW := w*2/3 - 3
	if listW > 48 {
		listW = 48
	}
	detailW := w - listW - 3
	if detailW < 20 {
		detailW = 20
	}

	type row struct {
		text string
		sel  bool
	}
	var rows []row
	lastPack := -1
	selIdx := -1
	for i, m := range a.filtered {
		if m.Pack != lastPack {
			rows = append(rows, row{text: escBold + escMag + " " + packs[m.Pack].Name + escReset + escDim + " [" + fmt.Sprintf("%03d-%03d", packs[m.Pack].Range[0], packs[m.Pack].Range[1]) + "]" + escReset})
			lastPack = m.Pack
		}
		rows = append(rows, row{
			text: fmt.Sprintf(" %s [%s] %s", m.ID, iconOf(m, a.stats), renderLine(m.Title, listW-16)),
			sel:  i == a.sel,
		})
		if i == a.sel {
			selIdx = len(rows) - 1
		}
	}

	scroll := 0
	if selIdx >= visRows {
		scroll = selIdx - visRows + 1
	}

	detail := a.detailPanel(a.selected(), detailW)

	for li, r := range rows {
		if li < scroll || li >= scroll+visRows {
			continue
		}
		left := " " + r.text
		if r.sel {
			left = escRev + renderLine(" "+r.text, listW) + escReset
		}
		right := ""
		if li-scroll < len(detail) {
			right = detail[li-scroll]
		}
		sb.WriteString("\n" + renderLine(left, listW) + escDim + "│" + escReset + renderLine(right, detailW))
	}

	// прогресс
	total := len(a.missions)
	done := 0
	for _, m := range a.missions {
		if a.stats.get(m.ID).Completed {
			done++
		}
	}
	sb.WriteString("\n\n " + bar(done, total, 24) + fmt.Sprintf("  %d/%d  %d%%", done, total, done*100/total))
	sb.WriteString("\n" + escDim + " ↑↓ — навигация · Enter — тест · r — запуск · t — задание · s — статистика · v — проверить все · g — переход · / — фильтр · q — выход" + escReset)
	if a.filter != "" {
		sb.WriteString("\n" + escYellow + " Фильтр: " + a.filter + escReset)
	}
	if a.statusLine != "" {
		sb.WriteString("\n" + a.statusLine)
	}
}

func (a *app) detailPanel(m *Mission, width int) []string {
	if m == nil {
		return nil
	}
	st := a.stats.get(m.ID)
	label, color := statusOf(m, a.stats)
	var lines []string
	lines = append(lines,
		escBold+escCyan+" "+m.ID+" "+m.Slug+escReset,
		" "+renderLine(m.Title, width-2),
		"",
		" Пакет: "+packs[m.Pack].Name,
		" Уровень: "+strings.Repeat("★", m.Stars)+strings.Repeat("☆", 5-m.Stars),
		" Статус: "+color+label+escReset,
		" Данные: "+yesno(m.HasData),
		" Попытки: "+itoa(st.Attempts),
	)
	if st.Completed {
		lines = append(lines, escGreen+"  Пройдена: "+st.CompletedAt+escReset)
	}
	lines = append(lines, "")
	lines = append(lines, escDim+" — задание —"+escReset)
	for _, l := range strings.Split(a.taskText(m), "\n") {
		lines = append(lines, renderLine(" "+l, width-2))
	}
	return lines
}

func yesno(b bool) string {
	if b {
		return escGreen + "есть" + escReset
	}
	return escRed + "нет" + escReset
}

func (a *app) renderStats(sb *strings.Builder, w, h int) {
	total := len(a.missions)
	done := 0
	byStars := map[int]int{}
	byStarsDone := map[int]int{}
	attempts := 0
	for _, m := range a.missions {
		st := a.stats.get(m.ID)
		attempts += st.Attempts
		byStars[m.Stars]++
		if st.Completed {
			done++
			byStarsDone[m.Stars]++
		}
	}
	sb.WriteString(escBold + escCyan + " СТАТИСТИКА" + escReset + "\n\n")
	sb.WriteString(fmt.Sprintf(" Пройдено: %s%d%s/%d  ·  попыток: %d\n\n", escGreen, done, escReset, total, attempts))
	sb.WriteString(" Общий прогресс:\n")
	sb.WriteString(" " + bar(done, total, 40) + fmt.Sprintf("  %d%%\n\n", done*100/total))
	sb.WriteString(" По пакетам:\n")
	for pi, p := range packs {
		pd, pt := 0, 0
		for _, m := range a.missions {
			if m.Pack == pi {
				pt++
				if a.stats.get(m.ID).Completed {
					pd++
				}
			}
		}
		if pt == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("  %-34s %s %2d/%2d\n", renderLine(p.Name, 34), bar(pd, pt, 20), pd, pt))
	}
	sb.WriteString("\n По сложности:\n")
	for s := 5; s >= 1; s-- {
		t := byStars[s]
		d := byStarsDone[s]
		label := strings.Repeat("★", s) + strings.Repeat("☆", 5-s)
		sb.WriteString(fmt.Sprintf("  %s %s %d/%d\n", label, bar(d, t, 20), d, t))
	}
	sb.WriteString("\n" + escDim + " Esc — назад" + escReset)
}

func bar(done, total, width int) string {
	if total == 0 {
		return strings.Repeat(" ", width)
	}
	filled := done * width / total
	return escCyan + strings.Repeat("█", filled) + escDim + strings.Repeat("░", width-filled) + escReset
}

func (a *app) renderOutput(sb *strings.Builder, w, h int) {
	sb.WriteString(escBold + escCyan + " " + a.outputTitle + escReset + "\n")
	body := h - 3
	for i := a.scroll; i < len(a.output) && i-a.scroll < body; i++ {
		line := renderLine(a.output[i], w-2)
		sb.WriteString("\n " + line)
	}
	sb.WriteString("\n\n" + escDim + " ↑↓ — прокрутка  Esc — назад" + escReset)
}

// ---------------------------------------------------------------- plain mode

func plain(root string) {
	missions := loadMissions(root)
	st := loadStats(root)
	total := len(missions)
	done := 0
	for _, m := range missions {
		if st.get(m.ID).Completed {
			done++
		}
	}
	fmt.Printf("NEON//GRID — %d миссий, пройдено %d\n\n", total, done)
	lastPack := -1
	for _, m := range missions {
		if m.Pack != lastPack {
			fmt.Printf("\n%s [%03d-%03d]\n", packs[m.Pack].Name, packs[m.Pack].Range[0], packs[m.Pack].Range[1])
			lastPack = m.Pack
		}
		icon := "○"
		if st.get(m.ID).Completed {
			icon = "✔"
		} else if !m.Todo {
			icon = "◐"
		}
		mark := ""
		if !m.HasData {
			mark = "  (нет данных)"
		}
		fmt.Printf("  %s %s %s%s\n", icon, m.ID, m.Title, mark)
	}
}
