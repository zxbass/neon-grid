// Reference algorithms for the Performance pack (missions 201-210).
package sol

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"
)

// ---------------------------------------------------------------- 201

// PerfProfileHot returns the function with the highest CPU share.
func PerfProfileHot(profile []byte) string {
	best, bestCpu := "", -1.0
	for _, ln := range strings.Split(strings.TrimSpace(string(profile)), "\n") {
		f := strings.Fields(ln)
		var cpu float64
		fmt.Sscanf(f[1], "%f", &cpu)
		if cpu > bestCpu {
			best, bestCpu = f[0], cpu
		}
	}
	return "HOT: " + best
}

// PerfProfileAlloc returns the function with the most allocations.
func PerfProfileAlloc(profile []byte) string {
	best, bestAlloc := "", -1
	for _, ln := range strings.Split(strings.TrimSpace(string(profile)), "\n") {
		f := strings.Fields(ln)
		var alloc int
		fmt.Sscanf(f[2], "%d", &alloc)
		if alloc > bestAlloc {
			best, bestAlloc = f[0], alloc
		}
	}
	return "ALLOC: " + best
}

// ---------------------------------------------------------------- 202

// PerfRecordsDump lists variable-length records (u8 type, u16 LE size, payload).
func PerfRecordsDump(records []byte) string {
	var sb strings.Builder
	i, n := 0, 0
	for i < len(records) {
		if i+3 > len(records) {
			break
		}
		t := records[i]
		size := int(binary.LittleEndian.Uint16(records[i+1 : i+3]))
		if i+3+size > len(records) {
			break
		}
		i += 3 + size
		n++
		fmt.Fprintf(&sb, "R%d: T=0x%02X LEN=%d\n", n, t, size)
	}
	fmt.Fprintf(&sb, "TOTAL: %d", n)
	return sb.String()
}

// PerfRecordsSum returns the byte sum of all payloads.
func PerfRecordsSum(records []byte) string {
	var sum byte
	i, n := 0, 0
	for i < len(records) {
		if i+3 > len(records) {
			break
		}
		size := int(binary.LittleEndian.Uint16(records[i+1 : i+3]))
		if i+3+size > len(records) {
			break
		}
		for j := 0; j < size; j++ {
			sum += records[i+3+j]
		}
		i += 3 + size
		n++
	}
	return fmt.Sprintf("SUM: 0x%02X\nTOTAL: %d", sum, n)
}

// ---------------------------------------------------------------- 203

// PerfPairsDump serializes ID VALUE pairs (decimal, one per line) to 10 LE bytes.
func PerfPairsDump(values []byte) string {
	var sb strings.Builder
	for i, ln := range strings.Split(strings.TrimSpace(string(values)), "\n") {
		id, v := parseDecimalPair(ln)
		var b [10]byte
		binary.LittleEndian.PutUint16(b[0:2], id)
		binary.LittleEndian.PutUint64(b[2:10], v)
		fmt.Fprintf(&sb, "P%d: %X\n", i+1, b)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// PerfPairsHex returns all serialized pairs concatenated as one hex string.
func PerfPairsHex(values []byte) string {
	var out strings.Builder
	for _, ln := range strings.Split(strings.TrimSpace(string(values)), "\n") {
		id, v := parseDecimalPair(ln)
		var b [10]byte
		binary.LittleEndian.PutUint16(b[0:2], id)
		binary.LittleEndian.PutUint64(b[2:10], v)
		fmt.Fprintf(&out, "%X", b)
	}
	return "HEX: " + out.String()
}

func parseDecimalPair(line string) (uint16, uint64) {
	p := 0
	for p < len(line) && line[p] != ' ' {
		p++
	}
	var id uint64
	for i := 0; i < p; i++ {
		id = id*10 + uint64(line[i]-'0')
	}
	var v uint64
	for i := p + 1; i < len(line); i++ {
		v = v*10 + uint64(line[i]-'0')
	}
	return uint16(id), v
}

// ---------------------------------------------------------------- 204

const perfMatrixN = 128

// PerfMatrixSum returns the total sum of a 128x128 int32 LE matrix.
func PerfMatrixSum(matrix []byte) string {
	var total int64
	for i := 0; i+4 <= len(matrix); i += 4 {
		total += int64(int32(binary.LittleEndian.Uint32(matrix[i : i+4])))
	}
	return fmt.Sprintf("SUM: %d", total)
}

// PerfMatrixCols returns per-column sums (row-major traversal).
func PerfMatrixCols(matrix []byte) string {
	col := make([]int64, perfMatrixN)
	for i := 0; i < perfMatrixN; i++ {
		for j := 0; j < perfMatrixN; j++ {
			v := int32(binary.LittleEndian.Uint32(matrix[(i*perfMatrixN+j)*4:]))
			col[j] += int64(v)
		}
	}
	var sb strings.Builder
	for j := 0; j < perfMatrixN; j++ {
		fmt.Fprintf(&sb, "C%d: %d\n", j, col[j])
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ---------------------------------------------------------------- 205

// PerfCRC8 computes CRC-8 (poly 0x07, init 0, MSB-first).
func PerfCRC8(payload []byte) byte {
	c := byte(0)
	for _, b := range payload {
		c ^= b
		for k := 0; k < 8; k++ {
			if c&0x80 != 0 {
				c = (c << 1) ^ 0x07
			} else {
				c <<= 1
			}
		}
	}
	return c
}

// PerfPacketsCRC lists CRC8 of each packet (u16 LE size + payload).
func PerfPacketsCRC(packets []byte) string {
	var sb strings.Builder
	i, n := 0, 0
	for i < len(packets) {
		l := int(binary.LittleEndian.Uint16(packets[i : i+2]))
		if i+2+l > len(packets) {
			break
		}
		n++
		fmt.Fprintf(&sb, "P%d: CRC8=0x%02X\n", n, PerfCRC8(packets[i+2:i+2+l]))
		i += 2 + l
	}
	return strings.TrimRight(sb.String(), "\n")
}

// PerfPacketsCheck returns the sum of all packet CRC8 values.
func PerfPacketsCheck(packets []byte) string {
	var sum byte
	i, n := 0, 0
	for i < len(packets) {
		l := int(binary.LittleEndian.Uint16(packets[i : i+2]))
		if i+2+l > len(packets) {
			break
		}
		n++
		sum += PerfCRC8(packets[i+2 : i+2+l])
		i += 2 + l
	}
	return fmt.Sprintf("CHECK: 0x%02X\nPACKETS: %d", sum, n)
}

// ---------------------------------------------------------------- 206

// PerfPopcount returns the number of set bits.
func PerfPopcount(v uint32) int {
	n := 0
	for v != 0 {
		v &= v - 1
		n++
	}
	return n
}

// PerfBitReverse32 mirrors the 32 bits of v.
func PerfBitReverse32(v uint32) uint32 {
	var r uint32
	for i := 0; i < 32; i++ {
		r = (r << 1) | (v & 1)
		v >>= 1
	}
	return r
}

// PerfValuesPopcount lists popcounts of u32 LE values.
func PerfValuesPopcount(values []byte) string {
	var sb strings.Builder
	for i := 0; i+4 <= len(values); i += 4 {
		v := binary.LittleEndian.Uint32(values[i : i+4])
		fmt.Fprintf(&sb, "0x%08X -> %d\n", v, PerfPopcount(v))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// PerfValuesBitReverse lists bit-reversed u32 LE values.
func PerfValuesBitReverse(values []byte) string {
	var sb strings.Builder
	for i := 0; i+4 <= len(values); i += 4 {
		v := binary.LittleEndian.Uint32(values[i : i+4])
		fmt.Fprintf(&sb, "0x%08X -> 0x%08X\n", v, PerfBitReverse32(v))
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ---------------------------------------------------------------- 207

// PerfRecordFields lists 16-byte records (id, magic, x, y, z).
func PerfRecordFields(records []byte) string {
	var sb strings.Builder
	for i := 0; i+16 <= len(records); i += 16 {
		id := binary.LittleEndian.Uint32(records[i:])
		x := binary.LittleEndian.Uint16(records[i+8:])
		y := binary.LittleEndian.Uint16(records[i+10:])
		z := binary.LittleEndian.Uint32(records[i+12:])
		fmt.Fprintf(&sb, "R%d: ID=0x%08X X=%d Y=%d Z=%d\n", i/16+1, id, x, y, z)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// PerfRecordZSum returns the sum of all z fields.
func PerfRecordZSum(records []byte) string {
	var sum uint64
	for i := 0; i+16 <= len(records); i += 16 {
		sum += uint64(binary.LittleEndian.Uint32(records[i+12:]))
	}
	return fmt.Sprintf("SUM: 0x%08X", sum)
}

// ---------------------------------------------------------------- 208

// PerfTopKeys returns the top-3 (key [16]byte + u64 LE value) by value desc.
func PerfTopKeys(index []byte) string {
	m := make(map[[16]byte]uint64)
	for i := 0; i+24 <= len(index); i += 24 {
		var k [16]byte
		copy(k[:], index[i:i+16])
		m[k] = binary.LittleEndian.Uint64(index[i+16 : i+24])
	}
	type kv struct {
		k [16]byte
		v uint64
	}
	all := make([]kv, 0, len(m))
	for k, v := range m {
		all = append(all, kv{k, v})
	}
	sort.Slice(all, func(a, b int) bool {
		if all[a].v != all[b].v {
			return all[a].v > all[b].v
		}
		return fmt.Sprintf("%X", all[a].k) < fmt.Sprintf("%X", all[b].k)
	})
	var sb strings.Builder
	for i := 0; i < 3 && i < len(all); i++ {
		fmt.Fprintf(&sb, "TOP%d: %X\n", i+1, all[i].k)
	}
	return strings.TrimRight(sb.String(), "\n")
}

// ---------------------------------------------------------------- 209

// PerfEventsSummary aggregates event counts and time bounds (u64 ts + u8 type).
func PerfEventsSummary(events []byte) string {
	counts := map[uint8]int{}
	var first, last uint64
	for i := 0; i+9 <= len(events); i += 9 {
		ts := binary.LittleEndian.Uint64(events[i:])
		t := events[i+8]
		counts[t]++
		if i == 0 {
			first = ts
		}
		last = ts
	}
	var sb strings.Builder
	for _, t := range []uint8{1, 2, 3} {
		fmt.Fprintf(&sb, "T%d: %d\n", t, counts[t])
	}
	fmt.Fprintf(&sb, "FIRST: 0x%016X\n", first)
	fmt.Fprintf(&sb, "LAST: 0x%016X", last)
	return sb.String()
}

// ---------------------------------------------------------------- 210

// PerfBlocksSummary aggregates variable-length blocks (u16 LE size + payload).
func PerfBlocksSummary(blocks []byte) string {
	var sum uint64
	count, maxLen := 0, 0
	for i := 0; i < len(blocks); {
		l := int(binary.LittleEndian.Uint16(blocks[i : i+2]))
		i += 2
		count++
		for j := 0; j < l; j++ {
			sum += uint64(blocks[i+j])
		}
		i += l
		if l > maxLen {
			maxLen = l
		}
	}
	return fmt.Sprintf("COUNT: %d\nSUM: 0x%08X\nMAXLEN: %d", count, sum&0xFFFFFFFF, maxLen)
}