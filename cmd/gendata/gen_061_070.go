package main

// Generators for missions 061-070 (terminal TUI conversions). All
// generators self-register below. Every mission is deterministic:
// inputs are text files under data/NNN/, expected.txt is derived from
// the same parsing/simulation rules a solution must implement.

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------- 061 dashboard

func (s *set) gen061() {
	id := "061"
	s.text(id, "progress.txt", "40\n")

	// part 1: single progress bar, width 20
	v := 40
	filled := v * 20 / 100
	p1 := "[" + strings.Repeat("#", filled) + strings.Repeat(".", 20-filled) + "] " + strconv.Itoa(v) + "%"

	// part 2: the fixed final dashboard frame
	dash := "ICE breach 98\nData leak 48\nTrace level 64\n"
	s.text(id, "dash.txt", dash)
	var rows []string
	for _, l := range strings.Split(strings.TrimSpace(dash), "\n") {
		f := strings.Fields(l)
		name, pct := strings.Join(f[:len(f)-1], " "), 0
		for _, ch := range f[len(f)-1] {
			pct = pct*10 + int(ch-'0')
		}
		fill := pct * 20 / 100
		rows = append(rows, fmt.Sprintf("| %-11s [%s] %2d%% |",
			name, strings.Repeat("#", fill)+strings.Repeat(".", 20-fill), pct))
	}
	border := "+" + strings.Repeat("-", 40) + "+"
	p2 := border + "\n" + strings.Join(rows, "\n") + "\n" + border + "\nMISSION 061 COMPLETE"
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 062 snake

func (s *set) gen062() {
	id := "062"
	const W, H = 20, 10
	keys := "dddddsssaaaawwwwwddddddddd" // 26 keys, one per tick
	s.text(id, "test_keys.txt", strings.Join(strings.Split(keys, ""), "\n")+"\n")
	s.text(id, "food.txt", "14 4\n10 7\n10 2\n0 0\n")

	type pt struct{ x, y int }
	dirs := map[byte]pt{'w': {0, -1}, 's': {0, 1}, 'a': {-1, 0}, 'd': {1, 0}}
	body := []pt{{9, 4}, {8, 4}, {7, 4}} // head first; heading right
	dir := pt{1, 0}
	foods := []pt{{14, 4}, {10, 7}, {10, 2}, {0, 0}}
	fi := 0
	food := foods[0]
	score := 0
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = bytes.Repeat([]byte{'.'}, W)
	}
	grid[food.y][food.x] = '@'

	over := ""
	for tick := 0; tick < 200; tick++ {
		if tick < len(keys) {
			if nd, ok := dirs[keys[tick]]; ok && !(nd.x == -dir.x && nd.y == -dir.y) {
				dir = nd
			}
		}
		head := pt{body[0].x + dir.x, body[0].y + dir.y}
		if head.x < 0 || head.x >= W || head.y < 0 || head.y >= H {
			over = "GAME OVER"
			break
		}
		eat := head == food
		if !eat {
			body = body[:len(body)-1] // tail moves away
		}
		collide := false
		for _, b := range body {
			if b == head {
				collide = true
			}
		}
		if collide {
			over = "GAME OVER"
			break
		}
		body = append([]pt{head}, body...)
		if eat {
			grid[head.y][head.x] = '.' // clear the eaten food
			score++
			if fi+1 < len(foods) {
				fi++
				food = foods[fi]
				grid[food.y][food.x] = '@'
			}
		}
	}
	if over == "" {
		over = "SURVIVED 200 ticks"
	}
	for _, b := range body {
		grid[b.y][b.x] = '#'
	}
	grid[body[0].y][body[0].x] = 'H'

	var rows []string
	for _, r := range grid {
		rows = append(rows, string(r))
	}
	p1 := strings.Join(rows, "\n")
	p2 := fmt.Sprintf("%s: score=%d", over, score)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 063 battleship

func (s *set) gen063() {
	id := "063"
	// board.txt: one ship per line "NAME SIZE CELLS..."; cells are "A1".."J10"
	board := "BS 4 B1 C1 D1 E1\n" +
		"C3 3 G1 H1 I1\n" +
		"S3 3 A4 B4 C4\n" +
		"D2 2 E6 F6\n" +
		"P2 2 H7 I7\n" +
		"K2 2 C9 D9\n" +
		"A1 1 J3\n" +
		"G1 1 G9\n" +
		"B1 1 B7\n" +
		"E1 1 A10\n"
	shots := "F1\nF4\nA5\nJ10\nB1\nC1\nD1\nE1\nG1\nH1\nI1\nA4\nB4\nC4\nE6\nF6\nH7\nI7\nC9\nD9\nJ3\nG9\nB7\nA10\n"
	s.text(id, "board.txt", board)
	s.text(id, "shots.txt", shots)

	parse := func(cell string) (int, int) { // row, col
		row, _ := strconv.Atoi(cell[1:])
		return row - 1, int(cell[0] - 'A')
	}
	type ship struct {
		name  string
		cells [][2]int
	}
	var ships []ship
	all := map[[2]int]string{}
	for _, l := range strings.Split(strings.TrimSpace(board), "\n") {
		f := strings.Fields(l)
		var cells [][2]int
		for _, c := range f[2:] {
			r, col := parse(c)
			cells = append(cells, [2]int{r, col})
			all[[2]int{r, col}] = f[0]
		}
		ships = append(ships, ship{f[0], cells})
	}
	// validate: ships must not touch, not even diagonally
	for i := range ships {
		for j := i + 1; j < len(ships); j++ {
			for _, a := range ships[i].cells {
				for _, b := range ships[j].cells {
					dr, dc := a[0]-b[0], a[1]-b[1]
					if dr < 0 {
						dr = -dr
					}
					if dc < 0 {
						dc = -dc
					}
					if dr <= 1 && dc <= 1 {
						panic("063: ships touch: " + ships[i].name + " " + ships[j].name)
					}
				}
			}
		}
	}

	// part 1: render the 10x10 field
	field := make([][]byte, 10)
	for r := range field {
		field[r] = bytes.Repeat([]byte{'~'}, 10)
	}
	for cell := range all {
		field[cell[0]][cell[1]] = '#'
	}
	var rows []string
	rows = append(rows, "   "+strings.Join(strings.Fields("A B C D E F G H I J"), " "))
	for r := 0; r < 10; r++ {
		rows = append(rows, fmt.Sprintf("%2d %s", r+1, strings.Join(strings.Fields(string(field[r])), " ")))
	}
	p1 := strings.Join(rows, "\n")

	// part 2: play the shot sequence
	var p2 []string
	hits, misses := 0, 0
	for _, l := range strings.Split(strings.TrimSpace(shots), "\n") {
		r, col := parse(l)
		if _, ok := all[[2]int{r, col}]; ok {
			hits++
			p2 = append(p2, l+" HIT")
		} else {
			misses++
			p2 = append(p2, l+" MISS")
		}
	}
	p2 = append(p2, fmt.Sprintf("VICTORY after %d shots (hits %d, misses %d)", hits+misses, hits, misses))
	s.expected(id, p1, strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 064 cursor artist

func (s *set) gen064() {
	id := "064"
	const W, H = 40, 20
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = bytes.Repeat([]byte{'.'}, W)
	}
	// a neon diamond: '#' outline, '*' fill
	for k := -5; k <= 5; k++ {
		r := 9 + k
		lo, hi := 20-k, 20+k
		if lo > hi {
			lo, hi = hi, lo
		}
		grid[r][lo], grid[r][hi] = '#', '#'
		for c := lo + 1; c < hi; c++ {
			grid[r][c] = '*'
		}
	}
	grid[3][33] = '@'
	grid[16][12] = '+'
	grid[1][20], grid[2][20] = '|', '|'
	grid[11][7] = '<'
	grid[11][32] = '>'
	grid[11][34], grid[11][35] = '-', '-'
	var art []string
	for _, r := range grid {
		art = append(art, string(r))
	}
	s.text(id, "art.txt", strings.Join(art, "\n")+"\n")

	p1 := strings.Join(art, "\n")

	// part 2: draw the cursor line on row 0: 19 dashes + '>'
	for c := 0; c < 19; c++ {
		grid[0][c] = '-'
	}
	grid[0][19] = '>'
	var out []string
	for _, r := range grid {
		out = append(out, string(r))
	}
	p2 := "SAVED (40x20)\n" + strings.Join(out, "\n")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 065 monitor

func (s *set) gen065() {
	id := "065"
	statText := "cpu 1000 0 0 8000 0 0 0 0\ncpu 1123 0 0 8877 0 0 0 0\n"
	memText := "MemTotal:       16777216 kB\nMemAvailable:   12373212 kB\n"
	loadText := "0.34 0.21 0.18 1/567 2345\n"
	procs1Text := "1 systemd 100 50\n204 chrome 500 300\n777 neongrid 200 100\n999 sshd 50 20\n" +
		"1234 kworker 10 5\n4321 docker 300 150\n5555 node 5 2\n8888 nginx 15 7\n"
	procs2Text := "1 systemd 101 50\n204 chrome 532 300\n777 neongrid 245 100\n999 sshd 62 20\n" +
		"1234 kworker 33 5\n4321 docker 310 150\n5555 node 25 2\n8888 nginx 40 7\n"
	s.text(id, "stat.txt", statText)
	s.text(id, "meminfo.txt", memText)
	s.text(id, "loadavg.txt", loadText)
	s.text(id, "procs1.txt", procs1Text)
	s.text(id, "procs2.txt", procs2Text)

	sum := func(line string) int {
		total := 0
		for _, f := range strings.Fields(line)[1:] {
			v, _ := strconv.Atoi(f)
			total += v
		}
		return total
	}
	stat := strings.Split(strings.TrimSpace(statText), "\n")
	deltaTotal := sum(stat[1]) - sum(stat[0])
	field := func(line string, i int) int {
		f := strings.Fields(line)
		v, _ := strconv.Atoi(f[i])
		return v
	}
	deltaIdle := field(stat[1], 4) - field(stat[0], 4) // idle = field index 4
	cpu := 100.0 * float64(deltaTotal-deltaIdle) / float64(deltaTotal)

	mem := strings.Split(strings.TrimSpace(memText), "\n")
	totalK, _ := strconv.Atoi(strings.Fields(mem[0])[1])
	availK, _ := strconv.Atoi(strings.Fields(mem[1])[1])
	usedK := totalK - availK
	memPct := usedK * 100 / totalK

	load := strings.Fields(strings.TrimSpace(loadText))[:3]

	p1 := fmt.Sprintf("CPU:  %.1f%%\nMEM:  %.1f GiB / %.0f GiB (%d%%)\nLOAD: %s",
		cpu, float64(usedK)/1048576, float64(totalK)/1048576, memPct, strings.Join(load, " "))

	// part 2: top-5 processes by %cpu
	procs := strings.Split(strings.TrimSpace(procs1Text), "\n")
	procs2 := strings.Split(strings.TrimSpace(procs2Text), "\n")
	type proc struct {
		pid int
		name string
		cpu float64
	}
	var list []proc
	for i, l := range procs {
		f := strings.Fields(l)
		f2 := strings.Fields(procs2[i])
		pid, _ := strconv.Atoi(f[0])
		u1, _ := strconv.Atoi(f[2])
		u2, _ := strconv.Atoi(f2[2])
		list = append(list, proc{pid, f[1], 100.0 * float64(u2-u1) / float64(deltaTotal)})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].cpu != list[j].cpu {
			return list[i].cpu > list[j].cpu
		}
		return list[i].pid < list[j].pid
	})
	var tab []string
	tab = append(tab, fmt.Sprintf("%-4s %-16s %s", "PID", "NAME", "CPU%"))
	for _, p := range list[:5] {
		tab = append(tab, fmt.Sprintf("%-4d %-16s %.1f", p.pid, p.name, p.cpu))
	}
	p2 := fmt.Sprintf("CPU:  %.1f%%  MEM: %d%%  LOAD: %s\n%s",
		cpu, memPct, strings.Join(load, " "), strings.Join(tab, "\n"))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 066 screensaver

func (s *set) gen066() {
	id := "066"
	s.text(id, "bounce.txt", "W 40\nH 20\nstart 5 5\ndx 1\ndy 2\nsteps 50\n")

	W, H := 40, 20
	x, y := 5, 5
	dx, dy := 1, 2
	steps := 50
	var pos [][2]int
	for i := 0; i < steps; i++ {
		x += dx
		y += dy
		if x < 0 {
			x = -x
			dx = -dx
		} else if x >= W {
			x = 2*(W-1) - x
			dx = -dx
		}
		if y < 0 {
			y = -y
			dy = -dy
		} else if y >= H {
			y = 2*(H-1) - y
			dy = -dy
		}
		pos = append(pos, [2]int{x, y})
	}
	var p1 []string
	for _, p := range pos {
		p1 = append(p1, fmt.Sprintf("%d,%d", p[0], p[1]))
	}

	// part 2: final frame, trail of the last 4 positions
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = bytes.Repeat([]byte{'.'}, W)
	}
	for i, p := range pos[len(pos)-4:] {
		ch := byte('o')
		if i == 3 {
			ch = '*'
		}
		grid[p[1]][p[0]] = ch
	}
	var rows []string
	for _, r := range grid {
		rows = append(rows, string(r))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(rows, "\n"))
}

// ---------------------------------------------------------------- 067 quest

func (s *set) gen067() {
	id := "067"
	quest := "# 1\n" +
		"name: Neon Lounge\n" +
		"desc: You are in the Neon Lounge.\n" +
		"exits: n=2 s=3\n" +
		"items: ram\n" +
		"\n" +
		"# 2\n" +
		"name: Back Alley\n" +
		"desc: A narrow alley behind the Neon Lounge.\n" +
		"exits: s=1\n" +
		"items: keycard\n" +
		"\n" +
		"# 3\n" +
		"name: Server Room\n" +
		"desc: Rows of servers hum in the dark.\n" +
		"exits: n=1 e=4\n" +
		"need: keycard\n" +
		"\n" +
		"# 4\n" +
		"name: Black Gate Terminal\n" +
		"desc: The Black Gate terminal looms before you.\n" +
		"exits: w=3\n"
	commands := "go s\ngo n\ntake keycard\ngo s\ngo s\ngo e\nwin\n"
	s.text(id, "quest.txt", quest)
	s.text(id, "commands.txt", commands)

	type room struct {
		id    int
		name   string
		desc   string
		exits  map[string]int
		items  []string
		need   string
	}
	rooms := map[int]*room{}
	var cur *room
	for _, l := range strings.Split(strings.TrimSpace(quest), "\n") {
		switch {
		case strings.HasPrefix(l, "# "):
			n, _ := strconv.Atoi(strings.TrimPrefix(l, "# "))
			cur = &room{id: n, exits: map[string]int{}}
			rooms[n] = cur
		case strings.HasPrefix(l, "name: "):
			cur.name = strings.TrimPrefix(l, "name: ")
		case strings.HasPrefix(l, "desc: "):
			cur.desc = strings.TrimPrefix(l, "desc: ")
		case strings.HasPrefix(l, "exits: "):
			for _, e := range strings.Fields(strings.TrimPrefix(l, "exits: ")) {
				d, t, _ := strings.Cut(e, "=")
				n, _ := strconv.Atoi(t)
				cur.exits[d] = n
			}
		case strings.HasPrefix(l, "items: "):
			cur.items = strings.Fields(strings.TrimPrefix(l, "items: "))
		case strings.HasPrefix(l, "need: "):
			cur.need = strings.TrimPrefix(l, "need: ")
		}
	}
	start := rooms[1]
	var exits []string
	for d := range start.exits {
		exits = append(exits, d)
	}
	sort.Strings(exits)
	p1 := start.desc + "\nExits: " + strings.Join(exits, " ") + "\nItems: [" + strings.Join(start.items, ", ") + "]"

	// part 2: simulate commands
	inv := []string{}
	here := start
	moves := 0
	winLine := ""
	for _, cmd := range strings.Split(strings.TrimSpace(commands), "\n") {
		moves++
		f := strings.Fields(cmd)
		switch f[0] {
		case "go":
			next, ok := here.exits[f[1]]
			if !ok {
				continue
			}
			dst := rooms[next]
			if dst.need != "" && !containsStr(inv, dst.need) {
				continue // LOCKED
			}
			here = dst
		case "take":
			if containsStr(here.items, f[1]) {
				inv = append(inv, f[1])
			}
		case "win":
			winLine = "WIN: You jacked into the Black Gate terminal."
		}
	}
	p2 := winLine + "\nMoves: " + strconv.Itoa(moves) + "  Items: " + strings.Join(inv, ", ")
	s.expected(id, p1, p2)
}

func containsStr(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------- 068 browser

func (s *set) gen068() {
	id := "068"
	pages := map[string]string{
		"index.html": "<html>\n<head><title>OMEGA-DYNE // Corporate Hub</title></head>\n" +
			"<body>\n<p>Welcome, citizen. Trust us.</p>\n" +
			"<a href=\"about.html\">About us</a>\n" +
			"<a href=\"login.html\">Employee login</a>\n" +
			"<a href=\"flag.html\">Restricted</a>\n</body>\n</html>\n",
		"about.html": "<html>\n<head><title>About OMEGA-DYNE</title></head>\n" +
			"<body>\n<p>We are the grid behind the grid.</p>\n" +
			"<a href=\"index.html\">Home</a>\n" +
			"<a href=\"flag.html\">The Gate</a>\n</body>\n</html>\n",
		"login.html": "<html>\n<head><title>Employee login</title></head>\n" +
			"<body>\n<p>Enter your credentials.</p>\n" +
			"<a href=\"index.html\">Home</a>\n" +
			"<a href=\"flag.html\">The Gate</a>\n</body>\n</html>\n",
		"flag.html": "<html>\n<head><title>FLAG PAGE</title></head>\n" +
			"<body>\n<p>FLAG{TUI_SURFING_IS_PURE}</p>\n</body>\n</html>\n",
	}
	for name, html := range pages {
		s.text(id, name, html)
	}
	commands := "1\n0\nback\n1\nquit\n"
	s.text(id, "commands.txt", commands)

	// parse a page: title, body text (links excluded), links in order
	type page struct {
		title, body string
		links       []string
	}
	parsePage := func(html string) page {
		var p page
		if i := strings.Index(html, "<title>"); i >= 0 {
			rest := html[i+len("<title>"):]
			if j := strings.Index(rest, "</title>"); j >= 0 {
				p.title = rest[:j]
			}
		}
		body := html
		if i := strings.Index(body, "<body>"); i >= 0 {
			body = body[i+len("<body>"):]
			if j := strings.Index(body, "</body>"); j >= 0 {
				body = body[:j]
			}
		}
		rest := body
		for {
			i := strings.Index(rest, "<a href=\"")
			if i < 0 {
				break
			}
			rest = rest[i+len("<a href=\""):]
			if j := strings.Index(rest, "\""); j >= 0 {
				p.links = append(p.links, rest[:j])
				rest = rest[j:]
			}
		}
		body = stripATags(body)
		body = stripTags(body)
		p.body = strings.Join(strings.Fields(body), " ")
		return p
	}
	docs := map[string]page{}
	for name, html := range pages {
		docs[name] = parsePage(html)
	}

	// part 1: index page
	idx := docs["index.html"]
	p1 := "TITLE: " + idx.title + "\nBODY: " + idx.body

	// part 2: navigation per commands.txt
	var out []string
	cur := "index.html"
	var hist []string
	for _, cmd := range strings.Split(strings.TrimSpace(commands), "\n") {
		if cmd == "quit" {
			break
		}
		out = append(out, "> "+cmd)
		if cmd == "back" {
			if len(hist) > 0 {
				cur = hist[len(hist)-1]
				hist = hist[:len(hist)-1]
			}
			continue
		}
		n, err := strconv.Atoi(cmd)
		if err != nil || n >= len(docs[cur].links) {
			continue
		}
		hist = append(hist, cur)
		cur = docs[cur].links[n]
	}
	final := docs[cur]
	out = append(out, "TITLE: "+final.title, "BODY: "+final.body)
	s.expected(id, p1, strings.Join(out, "\n"))
}

func stripATags(html string) string {
	var b strings.Builder
	for {
		i := strings.Index(html, "<a ")
		if i < 0 {
			b.WriteString(html)
			return b.String()
		}
		b.WriteString(html[:i])
		j := strings.Index(html[i:], "</a>")
		if j < 0 {
			return b.String()
		}
		html = html[i+j+len("</a>"):]
	}
}

func stripTags(s string) string {
	var b strings.Builder
	for {
		i := strings.Index(s, "<")
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		j := strings.Index(s[i:], ">")
		if j < 0 {
			return b.String()
		}
		s = s[i+j+1:]
	}
}

// ---------------------------------------------------------------- 069 editor

func (s *set) gen069() {
	id := "069"
	bufferText := "LINE1\nLINE2\nLINE3\n"
	commands := "iA\nd\niB\nu\nu\nr\nr\nj\np\n"
	s.text(id, "buffer.txt", bufferText)
	s.text(id, "commands.txt", commands)

	render := func(lines []string, cur int) string {
		var rows []string
		for i, l := range lines {
			if i == cur {
				rows = append(rows, "> "+l)
			} else {
				rows = append(rows, "  "+l)
			}
		}
		return strings.Join(rows, "\n")
	}
	lines := strings.Split(strings.TrimSpace(bufferText), "\n")
	p1 := render(lines, 0)

	// part 2: undo/redo simulation (snapshot-based)
	type snap struct {
		lines []string
		cur   int
	}
	do := func(state snap) []string { return append([]string{}, state.lines...) }
	var undo, redo []snap
	cur := 0
	for _, cmd := range strings.Split(strings.TrimSpace(commands), "\n") {
		switch {
		case strings.HasPrefix(cmd, "i") && len(cmd) > 1:
			undo = append(undo, snap{do(snap{lines, cur}), cur})
			lines[cur] += cmd[1:]
			redo = nil
		case cmd == "d":
			undo = append(undo, snap{do(snap{lines, cur}), cur})
			lines = append(lines[:cur], lines[cur+1:]...)
			if cur >= len(lines) {
				cur = len(lines) - 1
			}
			redo = nil
		case cmd == "j" && cur < len(lines)-1:
			cur++
		case cmd == "k" && cur > 0:
			cur--
		case strings.HasPrefix(cmd, "n") && len(cmd) > 1:
			undo = append(undo, snap{do(snap{lines, cur}), cur})
			lines = append(lines, "")
			copy(lines[cur+2:], lines[cur+1:])
			lines[cur+1] = cmd[1:]
			cur++
			redo = nil
		case cmd == "u" && len(undo) > 0:
			st := undo[len(undo)-1]
			undo = undo[:len(undo)-1]
			redo = append(redo, snap{do(snap{lines, cur}), cur})
			lines, cur = st.lines, st.cur
		case cmd == "r" && len(redo) > 0:
			st := redo[len(redo)-1]
			redo = redo[:len(redo)-1]
			undo = append(undo, snap{do(snap{lines, cur}), cur})
			lines, cur = st.lines, st.cur
		}
	}
	p2 := render(lines, cur)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 070 panel

func (s *set) gen070() {
	id := "070"
	const key = byte(0x5A)
	codes := "4A 0C 33 4C 0A 2B\n70 5D 92 A5 5B 00\n5A 5B 58 59 5E 5F\n"
	answers := "16 86 105 22 80 113\n42 7 200 255 1 90\n0 1 2 99 4 5\n"
	rounds := "4A 0C 33 4C 0A 2B|16 86 105 22 80 113|28\n" +
		"70 5D 92 A5 5B 00|42 7 200 255 1 90|31\n" +
		"5A 5B 58 59 5E 5F|0 1 2 3 4 5|12\n" +
		"4A 0C 33 4C 0A 2B|16 86 105 22 80 113|29\n" +
		"70 5D 92 A5 5B 00|0 0 0 0 0 0|21\n"
	s.text(id, "codes.txt", codes)
	s.text(id, "answers.txt", answers)
	s.text(id, "rounds.txt", rounds)

	decode := func(hex string) string {
		var dec []string
		for _, h := range strings.Fields(hex) {
			v, _ := strconv.ParseUint(h, 16, 8)
			dec = append(dec, strconv.Itoa(int(byte(v) ^ key)))
		}
		return strings.Join(dec, " ")
	}

	// part 1: decode + check each code
	var p1 []string
	codeLines := strings.Split(strings.TrimSpace(codes), "\n")
	ansLines := strings.Split(strings.TrimSpace(answers), "\n")
	for i, c := range codeLines {
		dec := decode(c)
		status := "MATCH"
		if dec != ansLines[i] {
			status = "FAIL"
		}
		p1 = append(p1, "CODE: "+c+" -> DECRYPTED: "+dec)
		p1 = append(p1, "CHECK: "+ansLines[i]+" -> "+status)
	}

	// part 2: rounds — pass if the typed answer matches AND ticks <= 30
	var p2 []string
	passed := 0
	for i, l := range strings.Split(strings.TrimSpace(rounds), "\n") {
		hex, rest, _ := strings.Cut(l, "|")
		ans, ticksStr, _ := strings.Cut(rest, "|")
		t, _ := strconv.Atoi(ticksStr)
		okCode := decode(hex) == ans
		okTime := t <= 30
		status := "ACCESS GRANTED"
		if !okCode {
			status = "TRACED (wrong code)"
		} else if !okTime {
			status = fmt.Sprintf("TRACED (%d ticks)", t)
		} else {
			passed++
		}
		p2 = append(p2, fmt.Sprintf("ROUND %d: %s", i+1, status))
	}
	final := "TRACED"
	if passed*2 >= 5 {
		final = "ACCESS GRANTED"
	}
	p2 = append(p2, fmt.Sprintf("ROUNDS: %d/5 passed", passed), "FINAL: "+final)
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

func init() {
	register("061", (*set).gen061)
	register("062", (*set).gen062)
	register("063", (*set).gen063)
	register("064", (*set).gen064)
	register("065", (*set).gen065)
	register("066", (*set).gen066)
	register("067", (*set).gen067)
	register("068", (*set).gen068)
	register("069", (*set).gen069)
	register("070", (*set).gen070)
}