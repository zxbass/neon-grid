package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"sort"
	"strings"

	"neon-grid/sol"
)

// ---------------------------------------------------------------- 161 regex

type regexNode struct {
	kind string // "lit", "any", "class", "star", "plus", "qmark", "alt", "seq"
	ch   byte
	rng  [][2]byte
	neg  bool
	kids []*regexNode
}

type regexParser struct {
	re   string
	pos  int
	alt  *regexNode
}

func parseRegex(s string) *regexNode {
	p := &regexParser{re: s}
	seq := parseSeq(p)
	return seq
}

func parseSeq(p *regexParser) *regexNode {
	var kids []*regexNode
	for p.pos < len(p.re) {
		c := p.re[p.pos]
		if c == '|' || c == ')' {
			break
		}
		kids = append(kids, parseAtom(p))
	}
	if len(kids) == 1 {
		return kids[0]
	}
	return &regexNode{kind: "seq", kids: kids}
}

func parseAtom(p *regexParser) *regexNode {
	var n *regexNode
	switch p.re[p.pos] {
	case '(':
		p.pos++
		n = parseSeq(p)
		p.pos++ // ')'
	case '[':
		p.pos++
		neg := false
		if p.pos < len(p.re) && p.re[p.pos] == '^' {
			neg = true
			p.pos++
		}
		var rng [][2]byte
		for p.pos < len(p.re) && p.re[p.pos] != ']' {
			if p.pos+2 < len(p.re) && p.re[p.pos+1] == '-' {
				rng = append(rng, [2]byte{p.re[p.pos], p.re[p.pos+2]})
				p.pos += 3
			} else {
				c := p.re[p.pos]
				rng = append(rng, [2]byte{c, c})
				p.pos++
			}
		}
		p.pos++ // ']'
		n = &regexNode{kind: "class", rng: rng, neg: neg}
	case '.':
		p.pos++
		n = &regexNode{kind: "any"}
	case '\\':
		p.pos++
		n = &regexNode{kind: "lit", ch: p.re[p.pos]}
		p.pos++
	default:
		n = &regexNode{kind: "lit", ch: p.re[p.pos]}
		p.pos++
	}
	// quantifier
	if p.pos < len(p.re) {
		switch p.re[p.pos] {
		case '*':
			n = &regexNode{kind: "star", kids: []*regexNode{n}}
			p.pos++
		case '+':
			n = &regexNode{kind: "plus", kids: []*regexNode{n}}
			p.pos++
		case '?':
			n = &regexNode{kind: "qmark", kids: []*regexNode{n}}
			p.pos++
		}
	}
	return n
}

func regexMatch(node *regexNode, s string) bool {
	var rec func(n *regexNode, pos int) []int
	rec = func(n *regexNode, pos int) []int {
		switch n.kind {
		case "lit":
			if pos < len(s) && s[pos] == n.ch {
				return []int{pos + 1}
			}
		case "any":
			if pos < len(s) {
				return []int{pos + 1}
			}
		case "class":
			if pos < len(s) {
				in := false
				for _, r := range n.rng {
					if s[pos] >= r[0] && s[pos] <= r[1] {
						in = true
						break
					}
				}
				if in != n.neg {
					return []int{pos + 1}
				}
			}
		case "seq":
			states := []int{pos}
			for _, k := range n.kids {
				var next []int
				for _, st := range states {
					next = append(next, rec(k, st)...)
				}
				states = next
				if len(states) == 0 {
					return nil
				}
			}
			return states
		case "star", "plus":
			min := 0
			if n.kind == "plus" {
				min = 1
			}
			reachable := map[int]bool{pos: true}
			depth := map[int]int{pos: 0}
			queue := []int{pos}
			for len(queue) > 0 {
				st := queue[0]
				queue = queue[1:]
				for _, e := range rec(n.kids[0], st) {
					if !reachable[e] {
						reachable[e] = true
						depth[e] = depth[st] + 1
						queue = append(queue, e)
					}
				}
			}
			var ends []int
			for e := range reachable {
				if e != pos && depth[e] >= min {
					ends = append(ends, e)
				}
			}
			if n.kind == "star" {
				ends = append(ends, pos)
			}
			return ends
		case "qmark":
			return append(rec(n.kids[0], pos), pos)
		}
		return nil
	}
	ends := rec(node, 0)
	for _, e := range ends {
		if e == len(s) {
			return true
		}
	}
	return false
}

func (s *set) gen161() {
	id := "161"
	pattern := "NE[OA]N[0-9]+"
	tests := []string{"NEON2049", "NEAN7", "NEON", "XNEON1", "NEMO", "N3ON99"}
	s.text(id, "pattern.txt", pattern)
	s.text(id, "strings.txt", strings.Join(tests, "\n"))
	re := parseRegex(pattern)
	var p1 []string
	n := 0
	for _, t := range tests {
		if regexMatch(re, t) {
			p1 = append(p1, "YES")
			n++
		} else {
			p1 = append(p1, "NO")
		}
	}
	p2 := fmt.Sprintf("MATCHES: %d", n)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 162 B-tree

type btree struct {
	order int
	root  *btNode
}

type btNode struct {
	keys    []int
	children []*btNode
	leaf    bool
}

func newBTree(order int) *btree {
	return &btree{order: order, root: &btNode{leaf: true}}
}

func (t *btree) insert(key int) {
	root := t.root
	if len(root.keys) == 2*t.order-1 {
		s := &btNode{leaf: false}
		s.children = []*btNode{root}
		splitChild(s, 0, root, t.order)
		t.root = s
	}
	insertNonFull(t.root, key, t.order)
}

func splitChild(parent *btNode, i int, child *btNode, order int) {
	mid := order - 1
	median := child.keys[mid]
	z := &btNode{leaf: child.leaf}
	z.keys = append([]int{}, child.keys[mid+1:]...)
	child.keys = child.keys[:mid]
	if !child.leaf {
		z.children = append([]*btNode{}, child.children[mid+1:]...)
		child.children = child.children[:mid+1]
	}
	parent.keys = append(parent.keys, 0)
	copy(parent.keys[i+1:], parent.keys[i:])
	parent.keys[i] = median
	parent.children = append(parent.children, nil)
	copy(parent.children[i+2:], parent.children[i+1:])
	parent.children[i+1] = z
}

func insertNonFull(n *btNode, key, order int) {
	i := len(n.keys) - 1
	if n.leaf {
		n.keys = append(n.keys, 0)
		for i >= 0 && key < n.keys[i] {
			n.keys[i+1] = n.keys[i]
			i--
		}
		n.keys[i+1] = key
		return
	}
	for i >= 0 && key < n.keys[i] {
		i--
	}
	i++
	if len(n.children[i].keys) == 2*order-1 {
		splitChild(n, i, n.children[i], order)
		if key > n.keys[i] {
			i++
		}
	}
	insertNonFull(n.children[i], key, order)
}

func btreeFind(n *btNode, key int) bool {
	i := 0
	for i < len(n.keys) && key > n.keys[i] {
		i++
	}
	if i < len(n.keys) && key == n.keys[i] {
		return true
	}
	if n.leaf {
		return false
	}
	return btreeFind(n.children[i], key)
}

func (s *set) gen162() {
	id := "162"
	inserts := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 5, 15}
	queries := []int{40, 55, 100, 5, 99, 15}
	s.text(id, "inserts.txt", strings.TrimSpace(strings.Join(intsToStrings(inserts), " ")))
	s.text(id, "queries.txt", strings.TrimSpace(strings.Join(intsToStrings(queries), " ")))
	t := newBTree(4)
	for _, k := range inserts {
		t.insert(k)
	}
	var p1 []string
	for _, q := range queries {
		if btreeFind(t.root, q) {
			p1 = append(p1, fmt.Sprintf("%d: FOUND", q))
		} else {
			p1 = append(p1, fmt.Sprintf("%d: NOT FOUND", q))
		}
	}
	var rootKeys []string
	for _, k := range t.root.keys {
		rootKeys = append(rootKeys, fmt.Sprint(k))
	}
	height := 1
	for n := t.root; len(n.children) > 0; n = n.children[0] {
		height++
	}
	p2 := fmt.Sprintf("ROOT: [%s]\nHEIGHT: %d", strings.Join(rootKeys, " "), height)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

func intsToStrings(v []int) []string {
	out := make([]string, len(v))
	for i, x := range v {
		out[i] = fmt.Sprint(x)
	}
	return out
}

// ---------------------------------------------------------------- 163 LRU cache

func (s *set) gen163() {
	id := "163"
	ops := []string{
		"PUT a 1",
		"GET a",
		"GET b",
		"PUT b 2",
		"GET a",
		"PUT c 3",
		"GET c",
		"GET b",
	}
	s.text(id, "ops.txt", strings.Join(ops, "\n"))
	capacity := 2
	var order []string
	vals := map[string]string{}
	var p1 []string
	for _, op := range ops {
		parts := strings.Fields(op)
		if parts[0] == "PUT" {
			if _, ok := vals[parts[1]]; !ok && len(order) == capacity {
				evict := order[0]
				order = order[1:]
				delete(vals, evict)
			}
			if _, ok := vals[parts[1]]; !ok {
				order = append(order, parts[1])
			} else {
				for i, k := range order {
					if k == parts[1] {
						order = append(order[:i], order[i+1:]...)
						break
					}
				}
				order = append(order, parts[1])
			}
			vals[parts[1]] = parts[2]
		} else {
			if v, ok := vals[parts[1]]; ok {
				p1 = append(p1, fmt.Sprintf("%s: HIT %s", parts[1], v))
				for i, k := range order {
					if k == parts[1] {
						order = append(order[:i], order[i+1:]...)
						break
					}
				}
				order = append(order, parts[1])
			} else {
				p1 = append(p1, fmt.Sprintf("%s: MISS", parts[1]))
			}
		}
	}
	var mru []string
	for i := len(order) - 1; i >= 0; i-- {
		mru = append(mru, fmt.Sprintf("%s=%s", order[i], vals[order[i]]))
	}
	p2 := "CACHE (MRU first): " + strings.Join(mru, " ")
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 164 Merkle

func h(pairs ...[]byte) []byte {
	if len(pairs) == 1 {
		return pairs[0]
	}
	var buf []byte
	for _, p := range pairs {
		buf = append(buf, p...)
	}
	sum := sha256.Sum256(buf)
	return sum[:]
}

func (s *set) gen164() {
	id := "164"
	leaves := []string{"aa", "bb", "cc", "dd"}
	var lines []string
	var lh [][]byte
	for i, l := range leaves {
		sum := sha256.Sum256([]byte(l))
		lh = append(lh, sum[:])
		lines = append(lines, fmt.Sprintf("leaf%d: %x", i, sum))
	}
	s.text(id, "leaves.txt", strings.Join(lines, "\n"))
	n1 := h(lh[0], lh[1])
	n2 := h(lh[2], lh[3])
	root := h(n1, n2)
	// proof for leaf 2 (index 1): siblings n0 sibling = lh[0] combined as h(lh0,lh1)=n1; then root sibling = n2
	var p1 []string
	p1 = append(p1, fmt.Sprintf("ROOT: %x", root))
	p1 = append(p1, fmt.Sprintf("N1: %x\nN2: %x", n1, n2))
	s.expected(id, strings.Join(p1[:1], "\n"), strings.Join(p1[1:], "\n")+"\nVERIFY: OK")
}

// ---------------------------------------------------------------- 165 sorting network

func (s *set) gen165() {
	id := "165"
	net := []string{"0-1", "2-3", "0-2", "1-3", "1-2"}
	input := []int{3, 1, 4, 2}
	s.text(id, "net.txt", strings.Join(net, "\n"))
	s.text(id, "input.txt", strings.TrimSpace(strings.Join(intsToStrings(input), " ")))
	arr := append([]int{}, input...)
	depth := 0
	layers := [][][2]int{}
	for _, c := range net {
		var i, j int
		fmt.Sscanf(c, "%d-%d", &i, &j)
		// place in first layer that doesn't conflict
		placed := false
		for li := range layers {
			conflict := false
			for _, pair := range layers[li] {
				if pair[0] == i || pair[1] == j || pair[0] == j || pair[1] == i {
					conflict = true
					break
				}
			}
			if !conflict {
				layers[li] = append(layers[li], [2]int{i, j})
				placed = true
				break
			}
		}
		if !placed {
			layers = append(layers, [][2]int{{i, j}})
		}
	}
	for _, layer := range layers {
		for _, pair := range layer {
			i, j := pair[0], pair[1]
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
		depth++
	}
	sorted := true
	for i := 1; i < len(arr); i++ {
		if arr[i-1] > arr[i] {
			sorted = false
		}
	}
	p1 := fmt.Sprintf("OUTPUT: %s\nSORTED: %s", strings.Join(intsToStrings(arr), " "), map[bool]string{true: "YES", false: "NO"}[sorted])
	p2 := fmt.Sprintf("NETWORK LAYERS: %d\nCOMPARATORS: %d", depth, len(net))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 166 A*

func (s *set) gen166() {
	id := "166"
	// maze with a unique shortest path (7x12), path found by A* (sol.AStar)
	grid := []string{
		"S...........",
		"###########.",
		"#...........",
		"#.##########",
		"..........##",
		"#########...",
		".........E..",
	}
	s.text(id, "map.txt", strings.Join(grid, "\n"))
	path := sol.AStar(grid, [2]int{0, 0}, [2]int{6, 9})
	if path == "" {
		panic("166: no path")
	}
	p1 := fmt.Sprintf("LEN: %d", len(path))
	p2 := "PATH: " + path
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 167 bloom

func hashFNV1a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

func hashDJB2(s string) uint32 {
	h := uint32(5381)
	for i := 0; i < len(s); i++ {
		h = h*33 + uint32(s[i])
	}
	return h
}

func hashSDBM(s string) uint32 {
	h := uint32(0)
	for i := 0; i < len(s); i++ {
		h = uint32(s[i]) + (h << 6) + (h << 16) - h
	}
	return h
}

func (s *set) gen167() {
	id := "167"
	words := []string{"crow", "zen", "grid", "neon", "ice", "mercury"}
	queries := []string{"crow", "zen", "alley", "grid", "mercury", "black", "gate", "neon"}
	s.text(id, "words.txt", strings.Join(words, "\n"))
	s.text(id, "queries.txt", strings.Join(queries, "\n"))
	const m = 32
	var bits [32]bool
	positions := func(w string) []int {
		return []int{int(hashFNV1a(w) % m), int(hashDJB2(w) % m), int(hashSDBM(w) % m)}
	}
	for _, w := range words {
		for _, p := range positions(w) {
			bits[p] = true
		}
	}
	var p1 []string
	for _, q := range queries {
		all := true
		for _, p := range positions(q) {
			if !bits[p] {
				all = false
			}
		}
		if all {
			p1 = append(p1, fmt.Sprintf("%s: MAYBE", q))
		} else {
			p1 = append(p1, fmt.Sprintf("%s: NO", q))
		}
	}
	set := 0
	var hexBytes []string
	for i := 0; i < 4; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << (7 - uint(j))
			}
		}
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				set++
			}
		}
		hexBytes = append(hexBytes, fmt.Sprintf("%02X", b))
	}
	p2 := fmt.Sprintf("BITS SET: %d\nMAP: %s", set, strings.Join(hexBytes, " "))
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 168 k-means

func (s *set) gen168() {
	id := "168"
	points := [][2]float64{
		{1, 1}, {1, 2}, {2, 1}, {2, 2},
		{10, 10}, {10, 11}, {11, 10}, {11, 11},
		{5, 20}, {6, 20}, {5, 21}, {6, 21},
	}
	var lines []string
	for _, p := range points {
		lines = append(lines, fmt.Sprintf("%d %d", int(p[0]), int(p[1])))
	}
	s.text(id, "points.txt", strings.Join(lines, "\n"))
	centroids := [][2]float64{{0, 0}, {10, 10}, {5, 20}}
	s.text(id, "centroids.txt", "0 0\n10 10\n5 20")
	for iter := 0; iter < 10; iter++ {
		cluster := make([][]int, 3)
		for i, p := range points {
			best, bd := 0, math.Inf(1)
			for k := range centroids {
				d := (p[0]-centroids[k][0])*(p[0]-centroids[k][0]) + (p[1]-centroids[k][1])*(p[1]-centroids[k][1])
				if d < bd {
					bd = d
					best = k
				}
			}
			cluster[best] = append(cluster[best], i)
		}
		for k := range centroids {
			var sx, sy float64
			for _, i := range cluster[k] {
				sx += points[i][0]
				sy += points[i][1]
			}
			if len(cluster[k]) > 0 {
				centroids[k] = [2]float64{sx / float64(len(cluster[k])), sy / float64(len(cluster[k]))}
			}
		}
	}
	var p1 []string
	var p2 []string
	for k := range centroids {
		p1 = append(p1, fmt.Sprintf("C%d: (%.1f, %.1f)", k, centroids[k][0], centroids[k][1]))
	}
	// counts by final assignment
	var cnt [3]int
	for _, p := range points {
		best := 0
		bd := math.Inf(1)
		for k := range centroids {
			d := (p[0]-centroids[k][0])*(p[0]-centroids[k][0]) + (p[1]-centroids[k][1])*(p[1]-centroids[k][1])
			if d < bd {
				bd = d
				best = k
			}
		}
		cnt[best]++
	}
	for k := range centroids {
		p2 = append(p2, fmt.Sprintf("CLUSTER %d: %d", k, cnt[k]))
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 169 naive bayes

func (s *set) gen169() {
	id := "169"
	spam := []string{"free money tonight", "free credit offer", "win free prize now"}
	ham := []string{"meeting at noon", "report ready", "lunch tomorrow"}
	tests := []string{"free lunch now", "report ready tonight", "win money"}
	s.text(id, "spam.txt", strings.Join(spam, "\n"))
	s.text(id, "ham.txt", strings.Join(ham, "\n"))
	s.text(id, "tests.txt", strings.Join(tests, "\n"))

	vocab := map[string]bool{}
	words := func(s string) []string { return strings.Fields(s) }
	spamCnt := map[string]int{}
	hamCnt := map[string]int{}
	for _, d := range spam {
		for _, w := range words(d) {
			spamCnt[w]++
			vocab[w] = true
		}
	}
	for _, d := range ham {
		for _, w := range words(d) {
			hamCnt[w]++
			vocab[w] = true
		}
	}
	V := len(vocab)
	totalSpam := 0
	for _, c := range spamCnt {
		totalSpam += c
	}
	totalHam := 0
	for _, c := range hamCnt {
		totalHam += c
	}
	// Laplace smoothing, compare P(spam|d) vs P(ham|d)
	var p1 []string
	for _, d := range tests {
		ls, lh := math.Log(0.5), math.Log(0.5)
		for _, w := range words(d) {
			ls += math.Log(float64(spamCnt[w]+1) / float64(totalSpam+V))
			lh += math.Log(float64(hamCnt[w]+1) / float64(totalHam+V))
		}
		if ls >= lh {
			p1 = append(p1, "SPAM")
		} else {
			p1 = append(p1, "HAM")
		}
	}
	p2 := fmt.Sprintf("VOCAB: %d\nSPAM DOCS: %d HAM DOCS: %d\nTOTAL WORDS: %d+%d", V, len(spam), len(ham), totalSpam, totalHam)
	s.expected(id, strings.Join(p1, "\n"), p2)
}

// ---------------------------------------------------------------- 170 markov

func (s *set) gen170() {
	id := "170"
	seed := "the grid remembers the grid never dies the grid is alive"
	s.text(id, "seed.txt", seed)
	counts := map[[2]byte]int{}
	for i := 0; i+1 < len(seed); i++ {
		counts[[2]byte{seed[i], seed[i+1]}]++
	}
	var gen []byte
	cur := byte('t')
	for i := 0; i < 40; i++ {
		gen = append(gen, cur)
		best, bc := byte(0), -1
		for c := byte(0); c < 128; c++ {
			if n := counts[[2]byte{cur, c}]; n > 0 && n > bc {
				bc = n
				best = c
			}
		}
		if bc <= 0 {
			break
		}
		cur = best
	}
	var keys [][2]byte
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return string(keys[i][:]) < string(keys[j][:])
	})

	var t3 []string
	for i := 0; i < 3 && i < len(keys); i++ {
		t3 = append(t3, fmt.Sprintf("%q x%d", string(keys[i][:]), counts[keys[i]]))
	}
	p1 := "GEN: " + string(gen)
	p2 := "TOP BIGRAMS: " + strings.Join(t3, ", ")
	s.expected(id, p1, p2)
}

func init() {
	register("161", (*set).gen161)
	register("162", (*set).gen162)
	register("163", (*set).gen163)
	register("164", (*set).gen164)
	register("165", (*set).gen165)
	register("166", (*set).gen166)
	register("167", (*set).gen167)
	register("168", (*set).gen168)
	register("169", (*set).gen169)
	register("170", (*set).gen170)
}
