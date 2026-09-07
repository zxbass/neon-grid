package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ---------------------------------------------------------------- 071 multichat

func (s *set) gen071() {
	id := "071"
	chat := "12:00:00 CONNECT crow\n" +
		"12:00:01 CONNECT zen\n" +
		"12:00:02 CONNECT ghost\n" +
		"12:00:05 [crow] ready\n" +
		"12:00:07 [zen] target locked\n" +
		"12:00:09 [ghost] copy that\n" +
		"12:00:12 [crow] reading map\n" +
		"12:00:14 [zen] found exit\n" +
		"12:00:16 [ghost] too slow\n" +
		"12:00:20 [ghost] left\n"
	s.text(id, "chat.log", chat)

	conns, msgs := 0, 0
	first := ""
	for _, l := range strings.Split(strings.TrimSpace(chat), "\n") {
		if strings.Contains(l, " CONNECT ") {
			conns++
			continue
		}
		if msgs == 0 {
			first = strings.SplitN(l, " ", 2)[1]
		}
		msgs++
	}
	p1 := fmt.Sprintf("CONNECTIONS: %d  MESSAGES: %d  DELIVERED: %d",
		conns, msgs, msgs*(conns-1))
	p2 := fmt.Sprintf("LOG: %d lines, first: %s", msgs, first)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 072 worker pool

type gen072job struct {
	id   string
	cost int
}

// simulatePool: greedy earliest-free assignment, ties to lowest worker index.
func simulatePool(list []gen072job, w int) int {
	load := make([]int, w)
	for _, t := range list {
		best := 0
		for i := 1; i < w; i++ {
			if load[i] < load[best] {
				best = i
			}
		}
		load[best] += t.cost
	}
	max := 0
	for _, l := range load {
		if l > max {
			max = l
		}
	}
	return max
}

func (s *set) gen072() {
	id := "072"
	tasks := "a 100\nb 20\nc 20\nd 20\ne 20\nf 20\ng 20\nh 20\ni 20\nj 20\n"
	s.text(id, "tasks.txt", tasks)

	var list []gen072job
	single := 0
	for _, l := range strings.Split(strings.TrimSpace(tasks), "\n") {
		f := strings.Fields(l)
		cost, _ := strconv.Atoi(f[1])
		list = append(list, gen072job{f[0], cost})
		single += cost
	}

	const w = 3
	dyn := simulatePool(list, w)

	static := 0
	ceil := (len(list) + w - 1) / w
	pos := 0
	for i := 0; i < w; i++ {
		n := ceil
		if i >= len(list)%w {
			n = ceil - 1
		}
		sum := 0
		for j := pos; j < pos+n; j++ {
			sum += list[j].cost
		}
		if sum > static {
			static = sum
		}
		pos += n
	}

	better := "static"
	if dyn < static {
		better = "dynamic"
	}
	p1 := fmt.Sprintf("W=%d  N=%d  T=%dms  (single=%dms, speedup=%.1fx)",
		w, len(list), dyn, single, float64(single)/float64(dyn))
	p2 := fmt.Sprintf("STATIC: %dms  DYNAMIC: %dms  (better: %s)", static, dyn, better)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 073 scanner

func (s *set) gen073() {
	id := "073"
	scan := "range 1 1024\nconcurrency 50\npulse 4\ndelay 200\n" +
		"22 OPEN\n80 OPEN\n443 FILTERED\n1200 OPEN\n8080 OPEN\n"
	s.text(id, "scan.txt", scan)

	n, c, pulse, delay := 0, 0, 0, 0
	var open []string
	for _, l := range strings.Split(strings.TrimSpace(scan), "\n") {
		f := strings.Fields(l)
		switch f[0] {
		case "range":
			start, _ := strconv.Atoi(f[1])
			end, _ := strconv.Atoi(f[2])
			n = end - start + 1
		case "concurrency":
			c, _ = strconv.Atoi(f[1])
		case "pulse":
			pulse, _ = strconv.Atoi(f[1])
		case "delay":
			delay, _ = strconv.Atoi(f[1])
		default:
			if f[1] == "OPEN" {
				open = append(open, f[0])
			}
		}
	}
	waves := (n + c - 1) / c
	p1 := fmt.Sprintf("OPEN: %s  C=%d: %d ports in %dms",
		strings.Join(open, " "), c, n, waves*delay)
	p2 := fmt.Sprintf("OPEN: %d  RATE: %d/s  QUEUED: %d  TIME: %dms",
		len(open), 1000/pulse, n-1, (n-1)*pulse+delay)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 074 race

func (s *set) gen074() {
	id := "074"
	runs := "5221913\n7999998\n6110254\n7883421\n5438765\n7999999\n" +
		"5777711\n7000123\n5554432\n7777888\n"
	s.text(id, "runs.txt", runs)
	fix := "LOCK 214\nATOMIC 31\nCHAN 198\n"
	s.text(id, "fix.txt", fix)

	var vals []int
	for _, l := range strings.Split(strings.TrimSpace(runs), "\n") {
		v, _ := strconv.Atoi(l)
		vals = append(vals, v)
	}
	min, max := vals[0], vals[0]
	for _, v := range vals[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	const expect = 8 * 1000000
	p1 := fmt.Sprintf("MIN=%d MAX=%d EXPECTED=%d  (гонка!)", min, max, expect)

	var p2 []string
	for _, l := range strings.Split(strings.TrimSpace(fix), "\n") {
		f := strings.Fields(l)
		p2 = append(p2, fmt.Sprintf("%s: %d in %sms", f[0], expect, f[1]))
	}
	s.expected(id, p1, strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 075 balancer

func (s *set) gen075() {
	id := "075"
	rr := "12:00:01 crow 1200\n" +
		"12:00:02 zen 1201\n" +
		"12:00:03 ghost 1202\n" +
		"12:00:04 reave 1200\n" +
		"12:00:05 miko 1201\n" +
		"12:00:06 pix 1202\n" +
		"12:00:07 crow 1200\n" +
		"12:00:08 zen 1201\n" +
		"12:00:09 ghost 1202\n"
	s.text(id, "rr.log", rr)
	health := "t=0 1200 UP\nt=0 1201 UP\nt=0 1202 UP\nt=2000 1201 DOWN\nt=4000 1201 UP\n"
	s.text(id, "health.log", health)
	conns := "1000\n1500\n2500\n3000\n3500\n4500\n5000\n5500\n6000\n"
	s.text(id, "conns.txt", conns)

	counts := map[string]int{}
	var order []string
	for _, l := range strings.Split(strings.TrimSpace(rr), "\n") {
		be := strings.Fields(l)[2]
		if counts[be] == 0 {
			order = append(order, be)
		}
		counts[be]++
	}
	var p1 []string
	for _, be := range order {
		p1 = append(p1, fmt.Sprintf("%s:%d", be, counts[be]))
	}

	type hcheck struct {
		t      int
		be     string
		status bool
	}
	var checks []hcheck
	for _, l := range strings.Split(strings.TrimSpace(health), "\n") {
		f := strings.Fields(l)
		t, _ := strconv.Atoi(strings.TrimPrefix(f[0], "t="))
		checks = append(checks, hcheck{t, f[1], f[2] == "UP"})
	}
	var times []int
	for _, l := range strings.Split(strings.TrimSpace(conns), "\n") {
		t, _ := strconv.Atoi(l)
		times = append(times, t)
	}

	backends := []string{"1200", "1201", "1202"}
	up := map[string]bool{}
	dist := map[string]int{}
	ptr, ci := 0, 0
	for _, t := range times {
		for ci < len(checks) && checks[ci].t <= t {
			up[checks[ci].be] = checks[ci].status
			ci++
		}
		for {
			be := backends[ptr%len(backends)]
			ptr++
			if up[be] {
				dist[be]++
				break
			}
		}
	}
	var p2 []string
	for _, be := range backends {
		p2 = append(p2, fmt.Sprintf("%s:%d", be, dist[be]))
	}
	s.expected(id, "BACKENDS: "+strings.Join(p1, " "), "DISTRIBUTION: "+strings.Join(p2, " "))
}

// ---------------------------------------------------------------- 076 scheduler

func (s *set) gen076() {
	id := "076"
	tasks := "t1 5 20\nt2 3 40\nt3 8 20\nt4 1 30\nt5 5 30\n"
	s.text(id, "tasks.txt", tasks)

	type task struct {
		id   string
		prio int
		ms   int
	}
	var list []task
	total := 0
	for _, l := range strings.Split(strings.TrimSpace(tasks), "\n") {
		f := strings.Fields(l)
		prio, _ := strconv.Atoi(f[1])
		ms, _ := strconv.Atoi(f[2])
		list = append(list, task{f[0], prio, ms})
		total += ms
	}

	var fifo, pr []string
	for _, t := range list {
		fifo = append(fifo, t.id)
	}
	byPrio := append([]task{}, list...)
	sort.SliceStable(byPrio, func(i, j int) bool { return byPrio[i].prio > byPrio[j].prio })
	for _, t := range byPrio {
		pr = append(pr, t.id)
	}
	p1 := fmt.Sprintf("fifo: %s  total=%dms\nprio: %s  total=%dms",
		strings.Join(fifo, " "), total, strings.Join(pr, " "), total)

	// RR with W=2: quantum 10ms, prio>=7 runs two quanta in a row.
	const quantum = 10
	type rrTask struct {
		id    string
		prio  int
		rem   int
		delay int
	}
	order := make([]*rrTask, len(list))
	queue := make([]*rrTask, len(list))
	for i := range list {
		order[i] = &rrTask{list[i].id, list[i].prio, list[i].ms, 0}
		queue[i] = order[i]
	}
	type wk struct {
		task   *rrTask
		quanta int
	}
	workers := []*wk{{nil, 0}, {nil, 0}}
	now := 0
	for len(queue) > 0 || workers[0].task != nil || workers[1].task != nil {
		now += quantum
		for _, w := range workers {
			if w.task == nil {
				continue
			}
			w.task.rem -= quantum
			w.quanta--
			if w.task.rem == 0 {
				w.task.delay = now
				w.task = nil
			} else if w.quanta == 0 {
				queue = append(queue, w.task)
				w.task = nil
			}
		}
		for _, w := range workers {
			if w.task != nil || len(queue) == 0 {
				continue
			}
			t := queue[0]
			queue = queue[1:]
			w.task = t
			w.quanta = 1
			if t.prio >= 7 {
				w.quanta = 2
			}
		}
	}
	delayParts := make([]string, len(list))
	sum := 0
	for i, t := range order {
		delayParts[i] = fmt.Sprintf("%s=%d", t.id, t.delay)
		sum += t.delay
	}
	p2 := fmt.Sprintf("W=2 RR: avg_delay=%.1fms (%s)",
		float64(sum)/float64(len(order)), strings.Join(delayParts, " "))
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 077 deadline

func (s *set) gen077() {
	id := "077"
	pairs := "100 500\n1000 500\n250 300\n"
	budget := "1 5000 4242\n1 1000 9999\n2 3000 7777\n"
	s.text(id, "pairs.txt", pairs)
	s.text(id, "budget.txt", budget)

	var p1 []string
	for _, l := range strings.Split(strings.TrimSpace(pairs), "\n") {
		f := strings.Fields(l)
		ms, _ := strconv.Atoi(f[0])
		d, _ := strconv.Atoi(f[1])
		st := "OK"
		if ms > d {
			st = "TIMEOUT"
		}
		p1 = append(p1, f[0]+" "+f[1]+"  -> "+st)
	}

	var p2 []string
	for _, l := range strings.Split(strings.TrimSpace(budget), "\n") {
		f := strings.Fields(l)
		t, _ := strconv.Atoi(f[0])
		b, _ := strconv.Atoi(f[1])
		valid, _ := strconv.Atoi(f[2])
		last := b/t - 1
		if valid <= last {
			p2 = append(p2, fmt.Sprintf("UNLOCKED %04d", valid))
		} else {
			p2 = append(p2, fmt.Sprintf("BOOM %04d", last))
		}
	}
	s.expected(id, strings.Join(p1, "\n"), strings.Join(p2, "\n"))
}

// ---------------------------------------------------------------- 078 replication

func (s *set) gen078() {
	id := "078"
	ops := "1 A SET a 1\n" +
		"2 B SET b 2\n" +
		"3 C SET a 3\n" +
		"4 A DEL b\n" +
		"5 B SET c 5\n" +
		"6 C SET a 6\n" +
		"7 A SET b 7\n" +
		"8 B DEL c\n" +
		"9 C SET a 9\n"
	s.text(id, "ops.log", ops)
	logA := "1 A SET k 10\n3 A SET m 1\n5 A SET x 1\n6 A DEL m\n"
	logB := "1 A SET k 10\n2 B SET m 2\n4 B SET z 99\n5 B SET x 2\n7 B SET x 7\n"
	s.text(id, "logA.txt", logA)
	s.text(id, "logB.txt", logB)

	apply := func(state map[string]string, keys *[]string, line string) {
		f := strings.Fields(line)
		key := f[3]
		if _, ok := state[key]; !ok {
			*keys = append(*keys, key)
		}
		if f[2] == "SET" {
			state[key] = f[4]
		} else {
			state[key] = "DELETED"
		}
	}
	render := func(state map[string]string, keys []string) string {
		sort.Strings(keys)
		var parts []string
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s=%s", k, state[k]))
		}
		return strings.Join(parts, " ")
	}

	state := map[string]string{}
	var keys []string
	n := 0
	for _, l := range strings.Split(strings.TrimSpace(ops), "\n") {
		apply(state, &keys, l)
		n++
	}
	p1 := fmt.Sprintf("FINAL: %s\nJOURNAL: %d entries", render(state, keys), n)

	type op struct {
		seq  int
		node string
		line string
	}
	var all []op
	for _, l := range strings.Split(strings.TrimSpace(logA), "\n") {
		f := strings.Fields(l)
		seq, _ := strconv.Atoi(f[0])
		all = append(all, op{seq, f[1], l})
	}
	for _, l := range strings.Split(strings.TrimSpace(logB), "\n") {
		f := strings.Fields(l)
		seq, _ := strconv.Atoi(f[0])
		all = append(all, op{seq, f[1], l})
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].seq != all[j].seq {
			return all[i].seq < all[j].seq
		}
		return all[i].node < all[j].node
	})
	state2 := map[string]string{}
	var keys2 []string
	for _, o := range all {
		apply(state2, &keys2, o.line)
	}
	p2 := "CONVERGED: " + render(state2, keys2)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 079 retry queue

func (s *set) gen079() {
	id := "079"
	msgs := "M1 OK\n" +
		"M2 FAIL OK\n" +
		"M3 OK\n" +
		"M4 FAIL FAIL OK\n" +
		"M5 FAIL FAIL FAIL FAIL FAIL\n" +
		"M6 FAIL OK\n" +
		"M7 FAIL FAIL FAIL OK\n" +
		"M8 OK\n" +
		"M9 FAIL FAIL FAIL FAIL OK\n" +
		"M10 FAIL FAIL FAIL FAIL FAIL\n"
	s.text(id, "msgs.txt", msgs)
	prio := "P1 9 1000 OK\n" +
		"P2 1 3000 FAIL FAIL FAIL FAIL FAIL\n" +
		"P3 5 1200 FAIL OK\n" +
		"P4 7 800 FAIL FAIL OK\n" +
		"P5 3 300 FAIL FAIL FAIL FAIL FAIL\n"
	s.text(id, "prio.txt", prio)

	delivered, dead, maxPause := 0, 0, 0
	for _, l := range strings.Split(strings.TrimSpace(msgs), "\n") {
		outcomes := strings.Fields(l)[1:]
		ok := false
		for i, o := range outcomes {
			if o == "OK" {
				ok = true
				break
			}
			if i < len(outcomes)-1 {
				pause := 100 << i
				if pause > maxPause {
					maxPause = pause
				}
			}
		}
		if ok {
			delivered++
		} else {
			dead++
		}
	}
	p1 := fmt.Sprintf("DELIVERED: %d  DEAD: %d\nMAX_PAUSE_USED: %dms",
		delivered, dead, maxPause)

	type msg struct {
		id       string
		prio     int
		ttl      int
		outcomes []string
		next     int
		fails    int
	}
	var msgs2 []*msg
	for _, l := range strings.Split(strings.TrimSpace(prio), "\n") {
		f := strings.Fields(l)
		p, _ := strconv.Atoi(f[1])
		ttl, _ := strconv.Atoi(f[2])
		msgs2 = append(msgs2, &msg{f[0], p, ttl, f[3:], 0, 0})
	}
	const step = 100
	del, d, exp := 0, 0, 0
	for t := 0; ; t += step {
		var pick *msg
		for _, m := range msgs2 {
			if m.outcomes == nil || m.next > t {
				continue
			}
			if pick == nil || m.prio > pick.prio {
				pick = m
			}
		}
		if pick == nil {
			alive := false
			for _, m := range msgs2 {
				if m.outcomes != nil {
					alive = true
					break
				}
			}
			if !alive {
				break
			}
			continue
		}
		if t > pick.ttl {
			pick.outcomes = nil
			exp++
			continue
		}
		if pick.outcomes[0] == "OK" {
			pick.outcomes = nil
			del++
			continue
		}
		pick.outcomes = pick.outcomes[1:]
		pick.fails++
		if len(pick.outcomes) == 0 {
			pick.outcomes = nil
			d++
			continue
		}
		pick.next = t + 100*(1<<(pick.fails-1))
	}
	p2 := fmt.Sprintf("DELIVERED: %d  DEAD: %d  EXPIRED: %d", del, d, exp)
	s.expected(id, p1, p2)
}

// ---------------------------------------------------------------- 080 log parser

func (s *set) gen080() {
	id := "080"
	log := "2049-12-01 03:14:15 10.0.0.7 USER crow ACTION login OK\n" +
		"2049-12-01 03:14:16 10.0.0.8 USER zen ACTION read plan.doc OK\n" +
		"2049-12-01 03:14:17 10.0.0.9 USER ghost ACTION sudo rm -rf /tmp/x OK\n" +
		"2049-12-01 03:14:18 10.0.0.7 USER crow ACTION read map.bin OK\n" +
		"2049-12-01 03:14:19 10.0.0.8 USER zen ACTION write log.raw OK\n" +
		"2049-12-01 03:14:20 10.0.0.10 USER reave ACTION delete data:users.db OK\n" +
		"2049-12-01 03:14:21 10.0.0.7 USER crow ACTION write vault.key OK\n" +
		"2049-12-01 03:14:22 10.0.0.8 USER zen ACTION read plan.doc OK\n" +
		"2049-12-01 03:14:23 10.0.0.9 USER ghost ACTION login OK\n" +
		"2049-12-01 03:14:24 10.0.0.10 USER reave ACTION scan 10.0.0.0/24 OK\n" +
		"2049-12-01 03:14:25 10.0.0.7 USER crow ACTION delete old.tmp OK\n" +
		"2049-12-01 03:14:26 10.0.0.11 USER miko ACTION login OK\n" +
		"2049-12-01 03:14:27 10.0.0.8 USER zen ACTION sudo apt update OK\n" +
		"2049-12-01 03:14:28 10.0.0.9 USER ghost ACTION read ghost.log OK\n" +
		"2049-12-01 03:14:29 10.0.0.7 USER crow ACTION login OK\n" +
		"2049-12-01 03:14:30 10.0.0.10 USER reave ACTION write shadow.db OK\n" +
		"2049-12-01 03:14:31 10.0.0.8 USER zen ACTION read plan.doc OK\n" +
		"2049-12-01 03:14:32 10.0.0.9 USER ghost ACTION rm ghost.old OK\n" +
		"2049-12-01 03:14:33 10.0.0.11 USER miko ACTION read mail.txt OK\n" +
		"2049-12-01 03:14:34 10.0.0.7 USER crow ACTION sudo scan 10.0.0.0/24 OK\n" +
		"2049-12-01 03:14:35 10.0.0.7 USER crow ACTION read grid.map OK\n" +
		"2049-12-01 03:14:36 10.0.0.8 USER zen ACTION login OK\n" +
		"2049-12-01 03:14:37 10.0.0.9 USER ghost ACTION write stash.bin OK\n" +
		"2049-12-01 03:14:38 10.0.0.7 USER crow ACTION write manifest.txt OK\n"
	s.text(id, "access.log", log)

	counts := map[string]int{}
	var susp []string
	rows := 0
	for _, l := range strings.Split(strings.TrimSpace(log), "\n") {
		rows++
		f := strings.Fields(l)
		user := f[4]
		action := strings.Join(f[6:len(f)-1], " ")
		counts[user]++
		la := strings.ToLower(action)
		if strings.Contains(la, "sudo") || strings.Contains(la, "rm") ||
			strings.Contains(la, "delete") {
			susp = append(susp, fmt.Sprintf("  %s: %s", user, action))
		}
	}
	type uc struct {
		user  string
		count int
	}
	var users []uc
	for u, c := range counts {
		users = append(users, uc{u, c})
	}
	sort.Slice(users, func(i, j int) bool {
		if users[i].count != users[j].count {
			return users[i].count > users[j].count
		}
		return users[i].user < users[j].user
	})
	var acts []string
	var top []string
	for i, u := range users {
		acts = append(acts, fmt.Sprintf("%s:%d", u.user, u.count))
		if i < 3 {
			top = append(top, u.user)
		}
	}
	p1 := fmt.Sprintf("ACTIONS: %s\nTOP: %s", strings.Join(acts, " "), strings.Join(top, " "))
	p2 := fmt.Sprintf("SUSPICIOUS: %d\n%s\nMERGE CHECK: parts=4 rows=%d matches_single=%d",
		len(susp), strings.Join(susp[:2], "\n"), rows, len(susp))
	s.expected(id, p1, p2)
}

func init() {
	register("071", (*set).gen071)
	register("072", (*set).gen072)
	register("073", (*set).gen073)
	register("074", (*set).gen074)
	register("075", (*set).gen075)
	register("076", (*set).gen076)
	register("077", (*set).gen077)
	register("078", (*set).gen078)
	register("079", (*set).gen079)
	register("080", (*set).gen080)
}