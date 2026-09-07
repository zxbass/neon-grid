package sol

// AStar finds the shortest path on a grid with 4-directional moves R/L/U/D.
// Used in mission 166. Rules (task 166):
//   - open set: min f = g + h; ties: min g, then earlier insertion;
//   - h is Manhattan distance to goal;
//   - neighbors expanded in order R, L, U, D;
//   - on equal g the first parent is kept.
// It returns the move sequence from start to goal, or "" if unreachable.
func AStar(grid []string, start, goal [2]int) string {
	H, W := len(grid), len(grid[0])
	dirs := [4][3]int{{0, 1, 'R'}, {0, -1, 'L'}, {-1, 0, 'U'}, {1, 0, 'D'}}
	manh := func(p [2]int) int {
		dr := p[0] - goal[0]
		if dr < 0 {
			dr = -dr
		}
		dc := p[1] - goal[1]
		if dc < 0 {
			dc = -dc
		}
		return dr + dc
	}
	type node struct {
		f, g, seq int
		p         [2]int
	}
	g := map[[2]int]int{start: 0}
	parent := map[[2]int][2]int{}
	closed := map[[2]int]bool{}
	heap := []node{{manh(start), 0, 0, start}}
	seq := 1
	for len(heap) > 0 {
		// pop min (f, g, seq)
		best := 0
		for i := 1; i < len(heap); i++ {
			a, b := heap[i], heap[best]
			if a.f < b.f || (a.f == b.f && (a.g < b.g || (a.g == b.g && a.seq < b.seq))) {
				best = i
			}
		}
		cur := heap[best]
		heap = append(heap[:best], heap[best+1:]...)
		if closed[cur.p] {
			continue
		}
		if cur.p == goal {
			var rev []byte
			p := cur.p
			for p != start {
				pp := parent[p]
				switch {
				case p[1] > pp[1]:
					rev = append(rev, 'R')
				case p[1] < pp[1]:
					rev = append(rev, 'L')
				case p[0] > pp[0]:
					rev = append(rev, 'D')
				default:
					rev = append(rev, 'U')
				}
				p = pp
			}
			for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
				rev[i], rev[j] = rev[j], rev[i]
			}
			return string(rev)
		}
		closed[cur.p] = true
		for _, d := range dirs {
			nb := [2]int{cur.p[0] + d[0], cur.p[1] + d[1]}
			if nb[0] < 0 || nb[0] >= H || nb[1] < 0 || nb[1] >= W || grid[nb[0]][nb[1]] == '#' {
				continue
			}
			ng := cur.g + 1
			if old, ok := g[nb]; ok && ng >= old {
				continue
			}
			g[nb] = ng
			parent[nb] = cur.p
			heap = append(heap, node{ng + manh(nb), ng, seq, nb})
			seq++
		}
	}
	return ""
}

// TetrisBoard simulates the falling of pieces (mission 172): field 6x10,
// pieces spawn at the top with left column x=3 and fall until blocked;
// full rows are cleared and counted. Shapes are given as rows of columns,
// e.g. 'I' = {{0,1,2,3}}, 'O' = {{0,1},{0,1}}. It returns the final field
// and the number of cleared lines.
func TetrisBoard(pieces string, shapes map[byte][][]int) ([]string, int) {
	const W, H = 10, 6
	grid := make([][]byte, H)
	for r := range grid {
		grid[r] = make([]byte, W)
	}
	lines := 0
	for i := 0; i < len(pieces); i++ {
		p := pieces[i]
		shape := shapes[p]
		y := 0
		for {
			ok := true
			for r, cols := range shape {
				for _, c := range cols {
					ny, nx := y+r, 3+c
					if ny >= H || grid[ny][nx] != 0 {
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
		for r, cols := range shape {
			for _, c := range cols {
				grid[y+r][3+c] = p
			}
		}
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
	out := make([]string, H)
	for r := 0; r < H; r++ {
		row := make([]byte, W)
		for c := 0; c < W; c++ {
			if grid[r][c] == 0 {
				row[c] = '.'
			} else {
				row[c] = grid[r][c]
			}
		}
		out[r] = string(row)
	}
	return out, lines
}