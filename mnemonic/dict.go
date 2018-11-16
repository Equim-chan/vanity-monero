package mnemonic

import (
	"encoding/binary"
	"hash/crc32"
)

const DictSize = 1626

var (
	ChineseSimplified = NewDict(chineseSimplifiedTable, 1)
	Dutch             = NewDict(dutchTable, 4)
	English           = NewDict(englishTable, 3)
	Esperanto         = NewDict(esperantoTable, 4)
	Spanish           = NewDict(spanishTable, 4)
	French            = NewDict(frenchTable, 4)
	German            = NewDict(germanTable, 4)
	Italian           = NewDict(italianTable, 4)
	Japanese          = NewDict(japaneseTable, 3)
	Lojban            = NewDict(lojbanTable, 4)
	Portuguese        = NewDict(portugueseTable, 4)
	Russian           = NewDict(russianTable, 4)
)

// Dict is a dictionary for mnemonic seed.
type Dict struct {
	UniquePrefixLength int
	Table              *[DictSize]string
	ReversedTable      map[string]int
}

func NewDict(table *[DictSize]string, prefixLen int) *Dict {
	d := &Dict{prefixLen, table, make(map[string]int)}
	for i, v := range table {
		d.ReversedTable[v] = i
	}

	return d
}

// Encode encodes a key to mnemonic seeds.
func (d *Dict) Encode(key *[32]byte) *[25]string {
	w := new([25]string)
	for i := 0; i < 32; i += 4 {
		x := binary.LittleEndian.Uint32(key[i : i+4])
		w1 := x % DictSize
		w2 := (x/DictSize + w1) % DictSize
		w3 := (x/DictSize/DictSize + w2) % DictSize
		w[i/4*3] = d.Table[w1]
		w[i/4*3+1] = d.Table[w2]
		w[i/4*3+2] = d.Table[w3]
	}
	w[24] = d.getChecksumWord(w)

	return w
}

func (d *Dict) getChecksumWord(w *[25]string) string {
	h := crc32.NewIEEE()
	for _, v := range w[:24] {
		r := string([]rune(v)[:d.UniquePrefixLength])
		h.Write([]byte(r))
	}
	sum := h.Sum32()
	idx := sum % 24

	return w[idx]
}
