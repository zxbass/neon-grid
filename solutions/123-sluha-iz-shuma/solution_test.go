package main

import (
	"testing"

	"neon-grid/solutions/kit"
)

func TestPart1(t *testing.T) {
	want, _ := kit.Expected("123")
	if got := Part1(); got != want {
		t.Fatalf("Part1() =\n%s\nwant:\n%s", got, want)
	}
}

func TestPart2(t *testing.T) {
	_, want := kit.Expected("123")
	if got := Part2(); got != want {
		t.Fatalf("Part2() =\n%s\nwant:\n%s", got, want)
	}
}
