# AGENTS.md

NEON//GRID: a Go "coding game" with 200 missions. Each mission is a stub the player implements;
tests compare `Part1()`/`Part2()` output against `data/NNN/expected.txt`. All docs/comments/expected
output are in Russian. Go 1.26.5, stdlib only, no external deps.

## Commands

```bash
go build ./... && go vet ./...   # must stay clean
go test ./sol/...                # reference algorithms — must stay green
go run ./cmd/gendata             # regenerate data/NNN + expected.txt (flag: -ids 101-110)
go run ./cmd/neon list|run|test <NNN>   # run from module root (relative paths)
go run ./cmd/neongrid            # TUI progress/statistics terminal (raw mode via ioctl, stdlib only)
go test ./solutions/<NNN-slug>/  # test one mission
```

## Critical gotchas

- **`go test ./solutions/...` failing is expected.** All 200 stubs are `panic("TODO")` by design; a
  green run only means the mission is implemented. Never "fix" stubs into passing — players fill
  them in.
- **Every solution dir is `package main`** (not `package missionNNN`) so both `go test` and
  `go run ./solutions/NNN-slug/` work. Each has `solution.go` (`Part1()`, `Part2()`, `main()`
  printing `=== Part 1 ===`/`=== Part 2 ===`) + `solution_test.go` in the same package.
- **`expected.txt` format** (enforced by `kit.Expected`): `part1` + `"\n=== PART 2 ===\n"` +
  `part2`. `Part1()`/`Part2()` must return strings with NO trailing newline (test compare against
  trimmed values). Note: header in tests is `=== PART 2 ===` (uppercase), while `main()` prints
  `=== Part 2 ===` — only return values are tested, so main() format is irrelevant.
- **50 missions have no data dirs and no gendata generators**: 041–080 (all 40), plus 104, 109, 112,
  114, 117, 118, 120, 124, 129, 130. Their tests panic in `kit.Expected` ("open
  data/041/expected.txt"), never reaching `Part1`. To implement one of these, you must also add a
  generator in `cmd/gendata/` (self-register via `register(id, fn)` in an `init()`) or create
  `data/NNN/expected.txt` manually.
- **Generators are the source of truth for expected output.** Missions 131–200 (plus 101–103,
  105–108, 110) were built generator-first: the generator's simulation defines the expected result,
  and the task file (`tasks/NNN-slug.md`) spells out the exact algorithm the player must reproduce
  (tie-breaks, rounding, float formats). If you change a generator, regenerate data; if you change
  expected semantics, update the task text too.
- **`kit` path resolution**: `kit.Data/ReadFile/ReadText/Expected` resolve from module root by
  walking up from CWD to `go.mod`. Works from any directory inside the module; tests and `neon` CLI
  assume run from root.
- `cmd/neon` and `cmd/gen` use relative paths (`solutions/`), so they must run from the repo root.
- **`cmd/neongrid` TUI**: stdlib-only raw terminal (termios via `syscall.SYS_IOCTL`,
  `tcgets=0x5401`/`tcsets=0x5402`). Status = `panic("TODO")` scan of `solution.go` + persisted
  `completed` flag in `.neongrid/stats.json` (gitignored). Verified-on-demand: `Enter` runs
  `go test ./solutions/NNN-slug/`. Works from repo root (finds `go.mod` by walking up). Non-tty
  invocation prints a plain list. Keys: ↑↓/PgUp/PgDn nav, Enter test, r run, t task, s stats, v
  verify-all, g jump, / filter, q/Esc exit/back.

## Layout

- `solutions/NNN-slug/` — mission stubs (player code); `solutions/kit/` — test helpers
- `tasks/NNN-slug.md` — mission descriptions (all 200 exist)
- `data/NNN/` — input files + `expected.txt` (150 populated)
- `sol/` — reference algorithm implementations + `sol_test.go`
- `cmd/gendata/` — data/expected generators (001–040, 081–110 partial, 111–130 partial, 131–200
  all); `cmd/gen/` — scratch demos; `cmd/neon/` — mission runner CLI
- Mission packs: 001–040 basics, 081–100 media, 111–113+115+116+119+121–123+125–128 finale arc,
  131–140 crypto attacks, 141–150 file formats II, 151–160 OS internals, 161–170 algorithms, 171–180
  terminal art, 181–190 packets/sniffers, 191–200 artifacts

## Generator internals worth reusing

- `gen_131_140.go`: `enigmaCrypt` (simplified Enigma), `sha256State`/`sha256WithState` (length
  extension), `crcAppendForce` (GF(2) CRC collision solver), `rc4`, MT19937 + `untemper`
- `gen_141_150.go`: `buildZip`, `makePNG` (in main.go), TTF/PDF/ISO/OLE builders
- `gen_151_160.go`: `cpuStepTrace` (Плесень-0 disasm/trace), buddy allocator sim
- `gen_161_170.go`: mini regex engine (`parseRegex`/`regexMatch`), B-tree
- `gen_191_200.go`: `goertzel`, `wavOf`, `rot13`, `bencodeString`
