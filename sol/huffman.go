package sol

import (
	"container/heap"
	"sort"
)

// HuffmanCodes builds Huffman codes for the given frequencies. Ties are
// broken by the smaller symbol first. Returns symbol -> binary code string.
// Mission 033, part 1.
func HuffmanCodes(freqs map[byte]int) map[byte]string {
	var h huffmanHeap
	for sym, f := range freqs {
		h = append(h, &huffNode{sym: sym, freq: f})
	}
	heap.Init(&h)
	for h.Len() > 1 {
		a := heap.Pop(&h).(*huffNode)
		b := heap.Pop(&h).(*huffNode)
		n := &huffNode{freq: a.freq + b.freq, left: a, right: b}
		heap.Push(&h, n)
	}
	root := h[0]
	out := make(map[byte]string)
	var walk func(n *huffNode, code string)
	walk = func(n *huffNode, code string) {
		if n.left == nil && n.right == nil {
			out[n.sym] = code
			return
		}
		walk(n.left, code+"0")
		walk(n.right, code+"1")
	}
	walk(root, "")
	return out
}

type huffNode struct {
	sym          byte
	freq         int
	left, right  *huffNode
}

type huffmanHeap []*huffNode

func (h huffmanHeap) Len() int { return len(h) }
func (h huffmanHeap) Less(i, j int) bool {
	if h[i].freq != h[j].freq {
		return h[i].freq < h[j].freq
	}
	return h[i].sym < h[j].sym
}
func (h huffmanHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *huffmanHeap) Push(x any)   { *h = append(*h, x.(*huffNode)) }
func (h *huffmanHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// CanonicalHuffmanCodes builds canonical codes from (length, symbol) pairs
// (mission 033, part 2): sorted by length then symbol, sequential codes.
func CanonicalHuffmanCodes(lens map[byte]int) map[byte]string {
	type pair struct {
		sym byte
		len int
	}
	var ps []pair
	for sym, l := range lens {
		ps = append(ps, pair{sym, l})
	}
	sort.Slice(ps, func(i, j int) bool {
		if ps[i].len != ps[j].len {
			return ps[i].len < ps[j].len
		}
		return ps[i].sym < ps[j].sym
	})
	code := 0
	out := make(map[byte]string, len(ps))
	prevLen := 0
	for _, p := range ps {
		code <<= (p.len - prevLen)
		b := make([]byte, p.len)
		for i := 0; i < p.len; i++ {
			if code&(1<<uint(p.len-1-i)) != 0 {
				b[i] = '1'
			} else {
				b[i] = '0'
			}
		}
		out[p.sym] = string(b)
		code++
		prevLen = p.len
	}
	return out
}
