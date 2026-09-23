package main

import (
	"fmt"
	"os"
	"strings"

	"neon-grid/solutions/kit"

	"github.com/zxbass/bt"
)

const (
	IdSize       = 4
	NicknameSize = 8
	LevelSize    = 1
	KarmaSize    = 1
	CreditsSize  = 4
	LastIpSize   = 8
	RecordSize   = IdSize + NicknameSize + LevelSize + KarmaSize + CreditsSize + LastIpSize
	endMark      = 0xFFFFFFFF
)

type Record struct {
	Id           uint32
	Nickname     string
	Level, Karma byte
	Credits      uint32
	LastIp       string
}

func (r Record) String() string {
	return fmt.Sprintf("ID=%08X nick=%-8s lvl=%d karma=%d cred=%d ip=%s",
		r.Id, r.Nickname, r.Level, r.Karma, r.Credits, r.LastIp)
}

func parseDump() ([]Record, bool, error) {
	data, err := os.ReadFile(kit.Data("007", "memdump.bin"))
	if err != nil {
		return nil, false, err
	}

	recs := make([]Record, 0, len(data)/RecordSize)
	cur := bt.NewCursor(data)
	finished := false
	for cur.BytesLeft() >= IdSize {
		id := cur.U32LE()
		if id == endMark {
			finished = true
			break
		}
		if cur.BytesLeft() < RecordSize {
			break
		}
		r := Record{}
		r.Id = id
		r.Nickname = cur.StrOrRest(NicknameSize)
		r.Level = cur.U8()
		r.Karma = cur.U8()
		r.Credits = cur.U32LE()
		r.LastIp = cur.StrOrRest(LastIpSize)
		recs = append(recs, r)
	}

	return recs, !finished || len(recs) == 0, nil
}

// Part1 разбирает записи memdump.bin и возвращает строки
// "ID=... nick=... lvl=... karma=... cred=... ip=...".
func Part1() string {
	recs, trunc, err := parseDump()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder
	for _, r := range recs {
		fmt.Fprintf(&sb, "%s\n", r)
	}
	if trunc {
		fmt.Fprint(&sb, "TRUNCATED")
	}

	return strings.TrimSpace(sb.String())
}

// Part2 возвращает подозрительных операторов (SUSPECT) и средний
// кредит (AVG).
func Part2() string {
	recs, trunc, err := parseDump()
	if err != nil {
		return fmt.Sprintf("ERROR: %v", err)
	}

	var sb strings.Builder
	var sumCredits int
	clean := true
	for _, r := range recs {
		if r.Level >= 5 && r.Karma < 20 {
			clean = false
			fmt.Fprintf(&sb, "SUSPECT: %08X %s\n", r.Id, r.Nickname)
		}
		sumCredits += int(r.Credits)
	}
	if clean && !trunc {
		fmt.Fprint(&sb, "CLEAN\n")
	}
	if trunc {
		fmt.Fprint(&sb, "TRUNCATED\n")
	} else {
		fmt.Fprintf(&sb, "AVG=%d", sumCredits/len(recs))
	}

	return strings.TrimSpace(sb.String())
}

func main() {
	fmt.Println("=== Part 1 ===")
	fmt.Println(Part1())
	fmt.Println("=== Part 2 ===")
	fmt.Println(Part2())
}
