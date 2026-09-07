package main

import (
	"testing"

	"neon-grid/solutions/kit"
)

func TestPart1(t *testing.T) {
	want, _ := kit.Expected("203")
	if got := Part1(); got != want {
		t.Fatalf("Part1() =\n%s\nwant:\n%s", got, want)
	}
}

func TestPart2(t *testing.T) {
	_, want := kit.Expected("203")
	if got := Part2(); got != want {
		t.Fatalf("Part2() =\n%s\nwant:\n%s", got, want)
	}
}

func TestPart2Allocs(t *testing.T) {
	if n := testing.AllocsPerRun(30, func() { _ = Part2() }); n >= 20 {
		t.Fatalf("Part2() allocations = %v, want < 20", n)
	}
}
