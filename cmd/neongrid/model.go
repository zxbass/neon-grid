package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------- mission model

type Mission struct {
	ID      string `json:"id"`
	Slug    string `json:"slug"`
	Dir     string `json:"dir"`
	Title   string `json:"title"`
	Stars   int    `json:"stars"`
	HasData bool   `json:"hasData"`
	Todo    bool   `json:"todo"` // стаб с panic("TODO")
	Pack    int    `json:"pack"` // индекс пакета
}

type Pack struct {
	Name  string
	Range [2]int
}

var packs = []Pack{
	{"Бинарные файлы и байты", [2]int{1, 10}},
	{"Форматы файлов", [2]int{11, 20}},
	{"Криптография", [2]int{21, 30}},
	{"Кодирование и сжатие", [2]int{31, 40}},
	{"Сокеты и сети, база", [2]int{41, 50}},
	{"Сети, глубокий уровень", [2]int{51, 60}},
	{"TUI-интерфейсы", [2]int{61, 70}},
	{"Параллельность и асинхронность", [2]int{71, 80}},
	{"Файловые системы и форензика", [2]int{81, 90}},
	{"Мультимедиа и сигналы", [2]int{91, 100}},
	{"Реверс-инжиниринг", [2]int{101, 110}},
	{"Финальный взлом", [2]int{111, 120}},
	{"После финала", [2]int{121, 130}},
	{"Крипто-атаки", [2]int{131, 140}},
	{"Форматы, том II", [2]int{141, 150}},
	{"Внутри ОС", [2]int{151, 160}},
	{"Алгоритмы", [2]int{161, 170}},
	{"Терминал-арт", [2]int{171, 180}},
	{"Сети: пакеты и снифферы", [2]int{181, 190}},
	{"Артефакты Сетки", [2]int{191, 200}},
}

func packOf(id int) int {
	for i, p := range packs {
		if id >= p.Range[0] && id <= p.Range[1] {
			return i
		}
	}
	return 0
}

func findRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}

var titleRe = regexp.MustCompile(`^#\s+(?:Миссия\s+)?(\d{3})\s*[-—]\s*(.+)$`)
var starsRe = regexp.MustCompile("`([★☆]{5})`")
var todoRe = regexp.MustCompile(`panic\("TODO"\)`)

func loadMissions(root string) []*Mission {
	entries, err := os.ReadDir(filepath.Join(root, "solutions"))
	if err != nil {
		return nil
	}
	var out []*Mission
	for _, e := range entries {
		if !e.IsDir() || len(e.Name()) < 4 || e.Name()[:3] < "001" || e.Name()[:3] > "200" {
			continue
		}
		id := e.Name()[:3]
		slug := e.Name()[4:]
		m := &Mission{
			ID:    id,
			Slug:  slug,
			Dir:   filepath.Join("solutions", e.Name()),
			Title: slug,
			Stars: 3,
		}
		n, _ := strAtoi(id)
		m.Pack = packOf(n)
		if _, err := os.Stat(filepath.Join(root, "data", id)); err == nil {
			m.HasData = true
		}
		// задача
		if b, err := os.ReadFile(filepath.Join(root, "tasks", id+"-"+slug+".md")); err == nil {
			text := string(b)
			firstLine := strings.SplitN(text, "\n", 2)[0]
			if m2 := titleRe.FindStringSubmatch(firstLine); m2 != nil {
				m.Title = strings.TrimSpace(m2[2])
			}
			if m3 := starsRe.FindStringSubmatch(text); m3 != nil {
				m.Stars = strings.Count(m3[1], "★")
			}
		}
		// стаб?
		if b, err := os.ReadFile(filepath.Join(root, m.Dir, "solution.go")); err == nil {
			m.Todo = todoRe.Match(b)
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func strAtoi(s string) (int, error) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("bad number %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// ---------------------------------------------------------------- statistics

type MissionStats struct {
	Attempts    int    `json:"attempts"`
	LastRun     string `json:"lastRun,omitempty"`
	Completed   bool   `json:"completed"`
	CompletedAt string `json:"completedAt,omitempty"`
}

type Stats struct {
	mu       sync.Mutex
	Missions map[string]*MissionStats `json:"missions"`
	file     string
}

func loadStats(root string) *Stats {
	st := &Stats{Missions: map[string]*MissionStats{}}
	p := filepath.Join(root, ".neongrid", "stats.json")
	st.file = p
	if b, err := os.ReadFile(p); err == nil {
		_ = json.Unmarshal(b, st)
	}
	if st.Missions == nil {
		st.Missions = map[string]*MissionStats{}
	}
	return st
}

func (s *Stats) get(id string) *MissionStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Missions[id] == nil {
		s.Missions[id] = &MissionStats{}
	}
	return s.Missions[id]
}

func (s *Stats) save() {
	s.mu.Lock()
	defer s.mu.Unlock()
	dir := filepath.Dir(s.file)
	_ = os.MkdirAll(dir, 0o755)
	b, _ := json.MarshalIndent(s, "", "  ")
	_ = os.WriteFile(s.file, b, 0o644)
}

func now() string { return time.Now().UTC().Format(time.RFC3339) }

func (s *Stats) touch(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.Missions[id]
	if m == nil {
		m = &MissionStats{}
		s.Missions[id] = m
	}
	m.Attempts++
	m.LastRun = now()
}

func (s *Stats) complete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m := s.Missions[id]
	if m == nil {
		m = &MissionStats{}
		s.Missions[id] = m
	}
	m.Completed = true
	m.CompletedAt = now()
}

// ---------------------------------------------------------------- helpers

func statusOf(m *Mission, st *Stats) (label string, color string) {
	if s := st.get(m.ID); s.Completed {
		return "ПРОЙДЕНА", escGreen
	}
	if !m.Todo {
		return "реализована", escYellow
	}
	if !m.HasData {
		return "нет данных", escRed
	}
	return "стаб", escDim
}

func iconOf(m *Mission, st *Stats) string {
	if s := st.get(m.ID); s.Completed {
		return escGreen + "✔" + escReset
	}
	if !m.Todo {
		return escYellow + "◐" + escReset
	}
	return escDim + "○" + escReset
}
