package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ---------------------------------------------------------------- 171 2048

func slide2048(row []int) ([]int, int) {
	// move all to the left, merge pairs once
	var vals []int
	for _, v := range row {
		if v != 0 {
			vals = append(vals, v)
		}
	}
	out := make([]int, len(row))
	score := 0
	i := 0
	for j := 0; j < len(vals); j++ {
		if j+1 < len(vals) && vals[j] == vals[j+1] {
			out[i] = vals[j] * 2
			score += vals[j] * 2
			j++
		} else {
			out[i] = vals[j]
		}
		i++
	}
	return out, score
}

func move2048(board [4][4]int, dir string) ([4][4]int, int) {
	score := 0
	var cols [4][4]int
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			cols[r][c] = board[c][r]
		}
	}
	get := func(rows bool) [4][4]int {
		if rows {
			return board
		}
		return cols
	}
	_ = get
	switch dir {
	case "UP", "DOWN":
		trans := [4][4]int{}
		for c := 0; c < 4; c++ {
			col := [4]int{board[0][c], board[1][c], board[2][c], board[3][c]}
			if dir == "DOWN" {
				reverse4(col[:])
			}
			sl, sc := slide2048(col[:])
			score += sc
			if dir == "DOWN" {
				reverse4(sl)
			}
			for r := 0; r < 4; r++ {
				trans[r][c] = sl[r]
			}
		}
		board = trans
	case "LEFT", "RIGHT":
		for r := 0; r < 4; r++ {
			row := [4]int{board[r][0], board[r][1], board[r][2], board[r][3]}
			if dir == "RIGHT" {
				reverse4(row[:])
			}
			sl, sc := slide2048(row[:])
			score += sc
			if dir == "RIGHT" {
				reverse4(sl)
			}
			copy(board[r][:], sl)
		}
	}
	return board, score
}

func reverse4(a []int) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func (s *set) gen171() {
	id := "171"
	board := [4][4]int{
		{2, 0, 0, 2},
		{0, 4, 0, 0},
		{2, 0, 2, 0},
		{0, 0, 0, 2},
	}
	moves := []string{"UP", "LEFT", "RIGHT", "DOWN", "UP", "LEFT", "DOWN", "RIGHT"}
	s.text(id, "moves.txt", strings.Join(moves, "\n"))
	var lines []string
	for r := 0; r < 4; r++ {
		var row []string
		for c := 0; c < 4; c++ {
			row = append(row, fmt.Sprint(board[r][c]))
		}
		lines = append(lines, strings.Join(row, " "))
	}
	s.text(id, "board.txt", strings.Join(lines, "\n"))
	total := 0
	for _, m := range moves {
		nb, sc := move2048(board, m)
		board = nb
		total += sc
	}
	var p1 []string
	for r := 0; r < 4; r++ {
		var row []string
		for c := 0; c < 4; c++ {
			row = append(row, fmt.Sprint(board[r][c]))
		}
		p1 = append(p1, strings.Join(row, " "))
	}
	p2 := fmt.Sprintf("SCORE: %d", total)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 172 tetris

var tetrisShapes = map[byte][][]int{
	'I': {{0, 1, 2, 3}},
	'O': {{0, 1}, {0, 1}},
	'T': {{0, 1, 2}, {1}},
	'S': {{1, 2}, {0, 1}},
	'Z': {{0, 1}, {1, 2}},
	'J': {{0}, {0, 1, 2}},
	'L': {{2}, {0, 1, 2}},
}

func (s *set) gen172() {
	id := "172"
	pieces := []byte(strings.Replace("I O T S Z J L I", " ", "", -1))
	s.text(id, "pieces.txt", "I O T S Z J L I")
	const W, H = 10, 6
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = make([]byte, W)
	}
	spawnX := 3
	lines := 0
	for _, p := range pieces {
		shape := tetrisShapes[p]
		rows := len(shape)
		// spawn at top, fall until collision
		y := 0
		for {
			ok := true
			for r := 0; r < rows; r++ {
				for _, c := range shape[r] {
					ny := y + r
					nx := spawnX + c
					if ny >= H || (ny >= 0 && grid[ny][nx] != 0) {
						ok = false
					}
				}
			}
			if !ok {
				break
			}
			y++
		}
		y--
		if y < 0 {
			y = 0
		}
		// lock
		for r := 0; r < rows; r++ {
			for _, c := range shape[r] {
				grid[y+r][spawnX+c] = p
			}
		}
		// clear full rows
		for r := 0; r < H; r++ {
			full := true
			for c := 0; c < W; c++ {
				if grid[r][c] == 0 {
					full = false
				}
			}
			if full {
				lines++
				copy(grid[1:r+1], grid[0:r])
				grid[0] = make([]byte, W)
			}
		}
	}
	var p1 []string
	for r := 0; r < H; r++ {
		var row []byte
		for c := 0; c < W; c++ {
			if grid[r][c] == 0 {
				row = append(row, '.')
			} else {
				row = append(row, grid[r][c])
			}
		}
		p1 = append(p1, string(row))
	}
	p2 := fmt.Sprintf("LINES CLEARED: %d\nPIECES: %d", lines, len(pieces))
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 173 game of life

func lifeStep(grid [][]byte) [][]byte {
	h, w := len(grid), len(grid[0])
	out := make([][]byte, h)
	for r := range out {
		out[r] = make([]byte, w)
		for c := range out[r] {
			out[r][c] = '.'
		}
	}
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			n := 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					if dr == 0 && dc == 0 {
						continue
					}
					rr, cc := r+dr, c+dc
					if rr >= 0 && rr < h && cc >= 0 && cc < w && grid[rr][cc] == '#' {
						n++
					}
				}
			}
			if grid[r][c] == '#' {
				if n == 2 || n == 3 {
					out[r][c] = '#'
				}
			} else if n == 3 {
				out[r][c] = '#'
			}
		}
	}
	return out
}

func (s *set) gen173() {
	id := "173"
	start := []string{
		"....................",
		"....##..............",
		"....##..............",
		"..........##........",
		"..........##........",
		"............##......",
		"............##......",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
		"....................",
	}
	grid := make([][]byte, len(start))
	for i, l := range start {
		grid[i] = []byte(l)
	}
	s.text(id, "grid.txt", strings.Join(start, "\n"))
	var counts []string
	for g := 1; g <= 10; g++ {
		grid = lifeStep(grid)
		live := 0
		for _, row := range grid {
			for _, ch := range row {
				if ch == '#' {
					live++
				}
			}
		}
		counts = append(counts, fmt.Sprintf("GEN %d: %d", g, live))
	}
	var p1 []string
	for _, row := range grid {
		p1 = append(p1, string(row))
	}
	p2 := strings.Join(counts, "\n")
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 174 dungeon

func (s *set) gen174() {
	id := "174"
	const W, H = 40, 20
	lcg := uint32(42)
	next := func(n int) int {
		lcg = 1664525*lcg + 1013904223
		return int(lcg % uint32(n))
	}
	type room struct{ x, y, w, h int }
	var rooms []room
	overlap := func(r room) bool {
		for _, o := range rooms {
			if r.x <= o.x+o.w && r.x+r.w >= o.x && r.y <= o.y+o.h && r.y+r.h >= o.y {
				return true
			}
		}
		return false
	}
	for i := 0; i < 6; i++ {
		for tries := 0; tries < 100; tries++ {
			r := room{next(W - 8), next(H - 6), 4 + next(5), 3 + next(4)}
			if !overlap(r) {
				rooms = append(rooms, r)
				break
			}
		}
	}
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = make([]byte, W)
		for c := range grid[r] {
			grid[r][c] = '#'
		}
	}
	floor := func(x, y int) {
		if x >= 0 && x < W && y >= 0 && y < H {
			grid[y][x] = '.'
		}
	}
	for _, r := range rooms {
		for y := r.y; y <= r.y+r.h; y++ {
			for x := r.x; x <= r.x+r.w; x++ {
				floor(x, y)
			}
		}
	}
	for i := 1; i < len(rooms); i++ {
		a, b := rooms[i-1], rooms[i]
		cx1, cy1 := a.x+a.w/2, a.y+a.h/2
		cx2, cy2 := b.x+b.w/2, b.y+b.h/2
		// L-образный коридор
		for x := min(cx1, cx2); x <= max(cx1, cx2); x++ {
			floor(x, cy1)
		}
		for y := min(cy1, cy2); y <= max(cy1, cy2); y++ {
			floor(cx2, y)
		}
	}
	var p1 []string
	for i, r := range rooms {
		p1 = append(p1, fmt.Sprintf("ROOM %d: x=%d y=%d w=%d h=%d", i, r.x, r.y, r.w, r.h))
	}
	var p2 []string
	for _, row := range grid {
		p2 = append(p2, string(row))
	}
	s.expected(id, strings.Join(p1, "\n"), "MAP:\n"+strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 175 raycasting

func (s *set) gen175() {
	id := "175"
	world := []string{
		"################",
		"#..............#",
		"#..............#",
		"#..####........#",
		"#..#....######.#",
		"#..#..........#",
		"#.....##.......#",
		"#......#.......#",
		"################",
	}
	var p2s []string
	for _, l := range world {
		p2s = append(p2s, l)
	}
	s.text(id, "map.txt", strings.Join(p2s, "\n"))
	px, py := 4.5, 4.5
	facing := 0.0
	const rays = 24
	fov := 60.0
	startAng := facing - fov/2
	var dists []string
	var strip []byte
	for i := 0; i < rays; i++ {
		ang := (startAng + fov*float64(i)/float64(rays-1)) * math.Pi / 180
		dx, dy := math.Cos(ang), math.Sin(ang)
		// DDA
		mapX, mapY := int(px), int(py)
		deltaX := math.Abs(1 / dx)
		deltaY := math.Abs(1 / dy)
		var stepX, stepY int
		var sideX, sideY float64
		if dx < 0 {
			stepX = -1
			sideX = (px - float64(mapX)) * deltaX
		} else {
			stepX = 1
			sideX = (float64(mapX+1) - px) * deltaX
		}
		if dy < 0 {
			stepY = -1
			sideY = (py - float64(mapY)) * deltaY
		} else {
			stepY = 1
			sideY = (float64(mapY+1) - py) * deltaY
		}
		hit := false
		side := 0
		for !hit {
			if sideX < sideY {
				sideX += deltaX
				mapX += stepX
				side = 0
			} else {
				sideY += deltaY
				mapY += stepY
				side = 1
			}
			if world[mapY][mapX] == '#' {
				hit = true
			}
		}
		var dist float64
		if side == 0 {
			dist = sideX - deltaX
		} else {
			dist = sideY - deltaY
		}
		dist = math.Max(dist, 0.001)
		dists = append(dists, fmt.Sprintf("%.2f", dist))
		// render strip
		height := int(5.0 / dist)
		if height > 8 {
			height = 8
		}
		space := 8 - height
		strip = append(strip, strings.Repeat(" ", space)...)
		strip = append(strip, strings.Repeat("#", height)...)
	}
	p1 := strings.Join(dists, " ")
	p2 := string(strip)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 176 chiptune

var noteFreq = map[string]float64{
	"C4": 261.63, "D4": 293.66, "E4": 329.63, "F4": 349.23,
	"G4": 392.00, "A4": 440.00, "B4": 493.88,
	"C5": 523.25, "D5": 587.33, "E5": 659.25, "G5": 783.99,
}

func (s *set) gen176() {
	id := "176"
	notes := []struct {
		name  string
		secs  float64
	}{
		{"C4", 0.25}, {"E4", 0.25}, {"G4", 0.5},
	}
	var lines []string
	total := 0.0
	for _, n := range notes {
		lines = append(lines, fmt.Sprintf("%s %.2f", n.name, n.secs))
		total += n.secs
	}
	s.text(id, "notes.txt", strings.Join(lines, "\n"))
	const sr = 8000
	samples := int(total * sr)
	f := noteFreq[notes[0].name]
	var first []string
	for i := 0; i < 8; i++ {
		t := float64(i) / sr
		phase := math.Sin(2 * math.Pi * f * t)
		var v float64
		if phase >= 0 {
			v = 0.3
		} else {
			v = -0.3
		}
		first = append(first, fmt.Sprint(int(v*32767+0.5)))
	}
	p1 := fmt.Sprintf("NOTES: %d DURATION: %.3fs\nSAMPLES: %d", len(notes), total, samples)
	p2 := "S0..S7: " + strings.Join(first, " ")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 177 tile editor

func (s *set) gen177() {
	id := "177"
	tile := [8][8]int{
		{0, 0, 0, 0, 1, 1, 0, 0},
		{0, 0, 0, 1, 2, 1, 0, 0},
		{0, 0, 1, 2, 2, 1, 0, 0},
		{0, 1, 2, 2, 2, 1, 0, 0},
		{1, 2, 2, 2, 1, 0, 0, 0},
		{0, 1, 2, 1, 0, 0, 0, 0},
		{0, 0, 1, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}
	var planes [2][]byte
	for pl := 0; pl < 2; pl++ {
		for y := 0; y < 8; y++ {
			var b byte
			for x := 0; x < 8; x++ {
				if tile[y][x]&(1<<pl) != 0 {
					b |= 1 << (7 - uint(x))
				}
			}
			planes[pl] = append(planes[pl], b)
		}
	}
	var bin []byte
	bin = append(bin, planes[0]...)
	bin = append(bin, planes[1]...)
	s.file(id, "tile.bin", bin)
	s.text(id, "palette.txt", "000000\n555555\nAAAAAA\nFFFFFF\n")
	chars := ".o#@"
	used := map[int]bool{}
	var p1 []string
	for y := 0; y < 8; y++ {
		var row []byte
		for x := 0; x < 8; x++ {
			pix := tile[y][x]
			used[pix] = true
			row = append(row, chars[pix])
		}
		p1 = append(p1, string(row))
	}
	var us []int
	for i := 0; i < 4; i++ {
		if used[i] {
			us = append(us, i)
		}
	}
	p2 := fmt.Sprintf("USED: %s", strings.Join(intsToStrings(us), ","))
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 178 matrix rain

func (s *set) gen178() {
	id := "178"
	const W, H = 20, 10
	drops := 6
	steps := 30
	seed := uint32(42)
	next := func(n int) int {
		seed = 1664525*seed + 1013904223
		return int(seed % uint32(n))
	}
	var xs []int
	for i := 0; i < drops; i++ {
		xs = append(xs, next(W))
	}
	ys := make([]int, drops)
	for s := 0; s < steps; s++ {
		for i := range ys {
			ys[i]++
			if ys[i] >= H {
				ys[i] = 0
				xs[i] = next(W)
			}
		}
	}
	frame := make([][]byte, H)
	for r := range frame {
		frame[r] = make([]byte, W)
		for c := range frame[r] {
			frame[r][c] = '.'
		}
	}
	for i := range ys {
		frame[ys[i]][xs[i]] = '#'
	}
	var lines []string
	for _, row := range frame {
		lines = append(lines, string(row))
	}
	s.text(id, "config.txt", fmt.Sprintf("WIDTH: %d\nHEIGHT: %d\nDROPS: %d\nSTEPS: %d\nSEED: 42", W, H, drops, steps))
	var pos []string
	for i := range xs {
		pos = append(pos, fmt.Sprintf("%d,%d", xs[i], ys[i]))
	}
	sort.Strings(pos)
	p1 := "FRAME 30:\n" + strings.Join(lines, "\n")
	p2 := "HEADS: " + strings.Join(pos, " ")
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 179 ascii video

func rleDecode(s string) string {
	var out []byte
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		n := 1
		if j > i {
			fmt.Sscanf(s[i:j], "%d", &n)
		}
		ch := byte(' ')
		if j < len(s) {
			ch = s[j]
		}
		for k := 0; k < n; k++ {
			out = append(out, ch)
		}
		i = j + 1
	}
	return string(out)
}

func (s *set) gen179() {
	id := "179"
	f1 := []string{"8.", "8.", "8.", "3.2#3.", "2.4#2.", "8.", "8.", "8."}
	f2 := []string{"8.", "8.", "8.", "8.", "8.", "8.", "4.3#.", "8."}
	s.text(id, "frames.txt", strings.Join(f1, "\n")+"\n\n"+strings.Join(f2, "\n"))
	d1 := make([][]byte, 8)
	for i, l := range f1 {
		d1[i] = []byte(rleDecode(l))
	}
	d2 := make([][]byte, 8)
	for i, l := range f2 {
		d2[i] = []byte(rleDecode(l))
	}
	changed := 0
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if d1[r][c] != d2[r][c] {
				changed++
			}
		}
	}
	var p1 []string
	for _, row := range d1 {
		p1 = append(p1, string(row))
	}
	p2 := fmt.Sprintf("CHANGED: %d", changed)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 180 terminal pong

func (s *set) gen180() {
	id := "180"
	s.text(id, "config.txt", "FIELD: 40x20\nPADDLE: 3\nLEFT_Y: 10\nRIGHT_Y: 10\nBALL: 20,10 VX=1 VY=1\n")
	W, H := 40, 20
	bx, by := 20.0, 10.0
	vx, vy := 1.0, 1.0
	leftY, rightY := 10, 10
	scoreL, scoreR := 0, 0
	bounces := 0
	for step := 0; step < 10000; step++ {
		bx += vx
		by += vy
		if by <= 0 || by >= float64(H-1) {
			vy = -vy
			by += vy
		}
		if int(bx) == 1 {
			if int(by) >= leftY-1 && int(by) <= leftY+1 {
				vx = -vx
				bounces++
			} else {
				scoreR++
				bx, by = 20, 10
				vx, vy = -vx, -vy
			}
		}
		if int(bx) == W-2 {
			if int(by) >= rightY-1 && int(by) <= rightY+1 {
				vx = -vx
				bounces++
			} else {
				scoreL++
				bx, by = 20, 10
				vx, vy = -vx, -vy
			}
		}
		if scoreL >= 5 || scoreR >= 5 {
			break
		}
	}
	p1 := fmt.Sprintf("SCORE: %d-%d", scoreL, scoreR)
	p2 := fmt.Sprintf("BOUNCES: %d\nWINNER: %s", bounces, map[bool]string{true: "LEFT", false: "RIGHT"}[scoreL > scoreR])
	s.expected(id, p1, p2)
}

func init() {
	register("171", (*set).gen171)
	register("172", (*set).gen172)
	register("173", (*set).gen173)
	register("174", (*set).gen174)
	register("175", (*set).gen175)
	register("176", (*set).gen176)
	register("177", (*set).gen177)
	register("178", (*set).gen178)
	register("179", (*set).gen179)
	register("180", (*set).gen180)
}
