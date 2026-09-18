package bt

import (
	"bytes"
	"encoding/binary"
)

func ReadCString(buf []byte) string {
	nulIdx := bytes.IndexByte(buf, 0)

	if nulIdx == -1 {
		return string(buf)
	}

	return string(buf[:nulIdx])
}

type Cursor struct {
	b   []byte
	off int
}

func NewCursor(b []byte) Cursor {
	return Cursor{b: b, off: 0}
}

func (c *Cursor) BytesLeft() int { return len(c.b[c.off:]) }

func (c *Cursor) U8() byte { v := c.b[c.off]; c.off += 1; return v }

func (c *Cursor) U16LE() uint16 { v := binary.LittleEndian.Uint16(c.b[c.off:]); c.off += 2; return v }

func (c *Cursor) U16BE() uint16 { v := binary.BigEndian.Uint16(c.b[c.off:]); c.off += 2; return v }

func (c *Cursor) U32LE() uint32 { v := binary.LittleEndian.Uint32(c.b[c.off:]); c.off += 4; return v }

func (c *Cursor) U32BE() uint32 { v := binary.BigEndian.Uint32(c.b[c.off:]); c.off += 4; return v }

func (c *Cursor) U64LE() uint64 { v := binary.LittleEndian.Uint64(c.b[c.off:]); c.off += 8; return v }

func (c *Cursor) U64BE() uint64 { v := binary.BigEndian.Uint64(c.b[c.off:]); c.off += 8; return v }

func (c *Cursor) Str(sz int) string { v := ReadCString(c.b[c.off : c.off+sz]); c.off += sz; return v }
