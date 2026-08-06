package sol

import (
	"bytes"
	"testing"
)

// ---------------------------------------------------------------- 038

func mustFrame(t *testing.T, s string) []byte {
	t.Helper()
	b, err := Hex(s)
	if err != nil {
		t.Fatalf("bad frame %q: %v", s, err)
	}
	return b
}

func TestMission038Frame(t *testing.T) {
	// Frame 1: pad 0xFF inside counted data; checksum is over real data.
	// Parser fails, then the fix rule recovers "HELLO".
	d, fixed, ok := Frame(mustFrame(t, "7e 06 00 48 45 4c 4c 4f ff 42 7e"))
	if !ok || !fixed || string(d) != "HELLO" {
		t.Fatalf("038 frame1: data=%q fixed=%v ok=%v", d, fixed, ok)
	}
	// Frame 2: already good.
	d, fixed, ok = Frame(mustFrame(t, "7e 04 00 47 52 49 44 18 7e"))
	if !ok || fixed || string(d) != "GRID" {
		t.Fatalf("038 frame2: data=%q fixed=%v ok=%v", d, fixed, ok)
	}
	// Frame 3: corrupt and not repairable.
	_, _, ok = Frame(mustFrame(t, "7e 04 00 41 42 43 44 55 7e"))
	if ok {
		t.Fatal("038 frame3: expected corrupt")
	}
	// Good frame from part 1 (XOR of ABC is 0x40, not 0x00!).
	d, fixed, ok = Frame(mustFrame(t, "7e 03 00 41 42 43 40 7e"))
	if !ok || fixed || !bytes.Equal(d, []byte("ABC")) {
		t.Fatalf("038 part1: data=%q fixed=%v ok=%v", d, fixed, ok)
	}
}
