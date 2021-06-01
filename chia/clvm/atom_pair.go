package clvm

import (
	"bytes"
	"encoding/hex"
	"math/big"
)

var NULL = Atom{[]byte{}}
var FALSE = NULL
var TRUE = Atom{[]byte{0x01}}

type SExp interface {
	Nullp() bool
	Listp() bool
	ListLen() int
	String() string
	StringExt(StringExtCfg) string
	DumpTo(*[]byte)
	Dump() []byte
}

type StringExtCfg struct {
	Keywords      bool
	OnlyHexValues bool
	CompactLists  bool
	Nil           string
	MaxDepth      int
	isNotRoot     bool
}

func (cfg StringExtCfg) KeywordsAnd(val bool) StringExtCfg {
	res := cfg
	res.Keywords = res.Keywords && val
	return res
}

var STRING_EXT_CFG_DEFAULT = StringExtCfg{Keywords: true, OnlyHexValues: false, CompactLists: true, Nil: "nil"}
var STRING_EXT_CFG_ERRORS = StringExtCfg{Keywords: true, OnlyHexValues: false, CompactLists: true, Nil: "nil", MaxDepth: 4}

type Atom struct {
	Bytes []byte
}

func (a Atom) Nullp() bool {
	return len(a.Bytes) == 0
}

func (a Atom) Listp() bool {
	return false
}

func (a Atom) ListLen() int {
	return 0
}

func (a Atom) AsInt() *big.Int {
	if len(a.Bytes) > 0 && a.Bytes[0]&0x80 != 0 {
		buf := make([]byte, len(a.Bytes))
		for i, b := range a.Bytes {
			buf[i] = ^b
		}
		out := new(big.Int).SetBytes(buf)
		out.Not(out)
		return out
	} else {
		return new(big.Int).SetBytes(a.Bytes)
	}
}

func (a Atom) Equal(other Atom) bool {
	return bytes.Equal(a.Bytes, other.Bytes)
}

func (a Atom) AsInt32() (int32, *EvalError) {
	l := len(a.Bytes)
	if l > 4 {
		return 0, NewEvalError("int32 requires 4 bytes at most, got %d: 0x%s",
			l, hex.EncodeToString(a.Bytes)).With("atom", a)
	}
	var v uint32 = 0
	if len(a.Bytes) > 0 && a.Bytes[0]&0x80 != 0 {
		v = ^uint32(0)
	}
	if l > 0 {
		v = (v << 8) + uint32(a.Bytes[0])
	}
	if l > 1 {
		v = (v << 8) + uint32(a.Bytes[1])
	}
	if l > 2 {
		v = (v << 8) + uint32(a.Bytes[2])
	}
	if l > 3 {
		v = (v << 8) + uint32(a.Bytes[3])
	}
	return int32(v), nil
}

func (a Atom) AsInt64() (int64, *EvalError) {
	l := len(a.Bytes)
	if l > 8 {
		return 0, NewEvalError("int64 requires 8 bytes at most, got %d: 0x%s",
			l, hex.EncodeToString(a.Bytes)).With("atom", a)
	}
	var v uint64 = 0
	if len(a.Bytes) > 0 && a.Bytes[0]&0x80 != 0 {
		v = ^uint64(0)
	}
	for _, b := range a.Bytes {
		v = (v << 8) + uint64(b)
	}
	return int64(v), nil
}

func (a Atom) AsBytes32() ([32]byte, *EvalError) {
	if len(a.Bytes) != 32 {
		return [32]byte{}, NewEvalError("expected 32 bytes, got %d: 0x%s",
			len(a.Bytes), hex.EncodeToString(a.Bytes)).With("atom", a)
	}
	var res [32]byte
	copy(res[:], a.Bytes)
	return res, nil
}

func (a Atom) StringExt(cfg StringExtCfg) string {
	if len(a.Bytes) == 0 {
		return cfg.Nil
	}
	if cfg.Keywords && cfg.isNotRoot && len(a.Bytes) == 1 {
		if op := OP_FROM_BYTE[a.Bytes[0]]; op.keyword != "" {
			return op.keyword
		}
	}
	if cfg.OnlyHexValues {
		return hex.EncodeToString(a.Bytes)
	} else {
		if len(a.Bytes) == 1 && a.Bytes[0] == 0 {
			return "0x00"
		}
		if len(a.Bytes) <= 2 {
			return a.AsInt().String()
		}
		allPrintable := true
		for _, c := range a.Bytes {
			if c < ' ' || c > '~' {
				allPrintable = false
				break
			}
		}
		if allPrintable {
			return `"` + string(a.Bytes) + `"`
		}
		return "0x" + hex.EncodeToString(a.Bytes)
	}
}

func (a Atom) String() string {
	return a.StringExt(STRING_EXT_CFG_DEFAULT)
}

func (a Atom) DumpTo(buf *[]byte) {
	SerializeAtomBytes(buf, a.Bytes)
}

func (a Atom) Dump() []byte {
	var buf []byte
	a.DumpTo(&buf)
	return buf
}

func AtomFromInt(v *big.Int) Atom {
	if v.Sign() == 0 {
		return NULL
	}
	var bytes []byte
	if v.Sign() < 0 {
		v = new(big.Int).Not(v)
		bytes = make([]byte, v.BitLen()/8+1)
		v.FillBytes(bytes)
		for i, b := range bytes {
			bytes[i] = ^b
		}
	} else {
		bytes = make([]byte, v.BitLen()/8+1)
		v.FillBytes(bytes)
	}
	return Atom{bytes}
}

type Pair struct {
	First SExp
	Rest  SExp
}

func (a Pair) Nullp() bool {
	return false
}
func (a Pair) Listp() bool {
	return true
}
func (a Pair) ListLen() int {
	var item SExp = a
	size := 0
	for {
		if pair, ok := item.(Pair); ok {
			item = pair.Rest
		} else {
			break
		}
		size += 1
	}
	return size
}
func (a Pair) StringExt(cfg StringExtCfg) string {
	if cfg.MaxDepth < 0 {
		return "..."
	}
	if cfg.MaxDepth > 0 {
		cfg.MaxDepth -= 1
		if cfg.MaxDepth == 0 {
			cfg.MaxDepth = -1
		}
	}
	cfg.isNotRoot = true
	leftStr := a.First.StringExt(cfg)
	if cfg.CompactLists {
		res := "(" + leftStr
		cur := a.Rest
		for !cur.Nullp() {
			if pair, ok := cur.(Pair); ok {
				_, isAtom := pair.First.(Atom)
				res += " " + pair.First.StringExt(cfg.KeywordsAnd(!isAtom))
				cur = pair.Rest
			} else {
				res += " . " + cur.StringExt(cfg.KeywordsAnd(false))
				break
			}
		}
		return res + ")"
	}
	_, rightIsAtom := a.Rest.(Atom)
	rightStr := a.Rest.StringExt(cfg.KeywordsAnd(!rightIsAtom))
	return "(" + leftStr + " . " + rightStr + ")"
}
func (a Pair) String() string {
	return a.StringExt(STRING_EXT_CFG_DEFAULT)
}
func (a Pair) DumpTo(buf *[]byte) {
	*buf = append(*buf, CONS_BOX_MARKER)
	a.First.DumpTo(buf)
	a.Rest.DumpTo(buf)
}
func (a Pair) Dump() []byte {
	var buf []byte
	a.DumpTo(&buf)
	return buf
}

type Iter struct {
	root  SExp
	cur   *Pair
	index int
	err   *EvalError
}

func NewIter(obj SExp) *Iter {
	return &Iter{root: obj, index: -1}
}

func (iter *Iter) Next() bool {
	if iter.err != nil {
		return false
	}

	var next SExp
	var logCur SExp
	if iter.cur == nil {
		next = iter.root
		logCur = iter.root
	} else {
		next = iter.cur.Rest
		logCur = iter.cur
	}

	if next.Nullp() {
		return false
	}

	pair, ok := next.(Pair)
	if !ok {
		iter.err = NewEvalError("wrong list: item.rest is atom on index %d", iter.index).
			With("item", logCur).With("list", iter.root)
		return false
	}
	iter.cur = &pair
	iter.index += 1
	return true
}

func (iter Iter) Get() SExp {
	return iter.cur.First
}

func (iter Iter) Err() *EvalError {
	return iter.err
}

type AtomIter struct {
	Iter
}

func NewAtomIter(obj SExp) *AtomIter {
	return &AtomIter{*NewIter(obj)}
}

func (iter *AtomIter) Next() bool {
	if !iter.Iter.Next() {
		return false
	}
	if _, ok := iter.cur.First.(Atom); !ok {
		iter.err = NewEvalError("wrong atom list: item.first is not atom on index %d", iter.index).
			With("item", iter.cur).With("list", iter.root)
		return false
	}
	return true
}

func (iter AtomIter) Get() Atom {
	return iter.cur.First.(Atom)
}

func AtomSliceFromList(obj SExp) ([]Atom, error) {
	var res []Atom
	iter := NewAtomIter(obj)
	for iter.Next() {
		res = append(res, iter.Get())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
