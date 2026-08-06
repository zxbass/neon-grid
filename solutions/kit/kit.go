package kit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const marker = "\n=== PART 2 ===\n"

var rootDir string

func init() {
	rootDir = findRoot()
}

func findRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
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

// Data returns the path to a mission data file, relative to the module root.
func Data(id, name string) string {
	return filepath.Join(rootDir, "data", id, name)
}

// ReadFile reads a mission data file entirely.
func ReadFile(t *testing.T, id, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(Data(id, name))
	if err != nil {
		t.Fatalf("read %s: %v", Data(id, name), err)
	}
	return b
}

// ReadText reads a mission data text file without trailing newline.
func ReadText(t *testing.T, id, name string) string {
	t.Helper()
	return strings.TrimRight(string(ReadFile(t, id, name)), "\n")
}

// Expected returns part1 and part2 from the mission's expected.txt.
func Expected(id string) (part1, part2 string) {
	b, err := os.ReadFile(filepath.Join(rootDir, "data", id, "expected.txt"))
	if err != nil {
		panic("kit.Expected: " + err.Error())
	}
	parts := strings.SplitN(string(b), marker, 2)
	part1 = parts[0]
	if len(parts) == 2 {
		part2 = strings.TrimRight(parts[1], "\n")
	}
	return part1, part2
}
