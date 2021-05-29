package chia

import (
	"bytes"
	"encoding/hex"
	"log"
	"math/big"
)

var ATOM_NULL = CLVMAtom{[]byte{}}
var ATOM_FALSE = ATOM_NULL
var ATOM_TRUE = CLVMAtom{[]byte{0x01}}

type CLVMObject interface {
	Nullp() bool
	Listp() bool
	ListLen() int
	String() string
	StringExt(keywords bool, hexValues bool, bracketNil bool, compactLists bool) string
}
type CLVMAtom struct {
	Bytes []byte
}

func (a CLVMAtom) Nullp() bool {
	return len(a.Bytes) == 0
}
func (a CLVMAtom) Listp() bool {
	return false
}
func (a CLVMAtom) ListLen() int {
	return 0
}
func (a CLVMAtom) AsInt() *big.Int {
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
func (a CLVMAtom) Equal(other CLVMAtom) bool {
	return bytes.Equal(a.Bytes, other.Bytes)
}
func CLVMAtomFromInt(v *big.Int) CLVMAtom {
	if v.Sign() == 0 {
		return ATOM_NULL
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
	return CLVMAtom{bytes}
}
func (a CLVMAtom) AsInt32() int32 {
	l := len(a.Bytes)
	if l > 4 {
		log.Fatalf("int32 requires 4 bytes at most, got %d: 0x%s", len(a.Bytes), hex.EncodeToString(a.Bytes))
	}
	var v uint32 = 0
	if l > 0 {
		v = uint32(a.Bytes[0])
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
	return int32(v)
}
func (a CLVMAtom) StringExt(keywords bool, hexValues bool, bracketNil bool, compactLists bool) string {
	if len(a.Bytes) == 0 {
		if bracketNil {
			return "()"
		} else {
			return "nil"
		}
	}
	if keywords && len(a.Bytes) == 1 {
		if op := OP_FROM_BYTE[a.Bytes[0]]; op.keyword != "" {
			return op.keyword
		}
	}
	if hexValues {
		return hex.EncodeToString(a.Bytes)
	} else {
		if len(a.Bytes) <= 2 {
			return (&big.Int{}).SetBytes(a.Bytes).String()
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
func (a CLVMAtom) String() string {
	return a.StringExt(false, true, false, false)
}

type CLVMPair struct {
	First CLVMObject
	Rest  CLVMObject
}

func (a CLVMPair) Nullp() bool {
	return false
}
func (a CLVMPair) Listp() bool {
	return true
}
func (a CLVMPair) ListLen() int {
	var item CLVMObject = a
	size := 0
	for {
		if pair, ok := item.(CLVMPair); ok {
			item = pair.Rest
		} else {
			break
		}
		size += 1
	}
	return size
}
func (a CLVMPair) StringExt(keywords bool, hexValues bool, bracketNil bool, compactLists bool) string {
	leftStr := a.First.StringExt(keywords, hexValues, bracketNil, compactLists)
	if compactLists {
		res := "(" + leftStr
		cur := a.Rest
		for !cur.Nullp() {
			if pair, ok := cur.(CLVMPair); ok {
				_, isAtom := pair.First.(CLVMAtom)
				res += " " + pair.First.StringExt(keywords && !isAtom, hexValues, bracketNil, compactLists)
				cur = pair.Rest
			} else {
				res += " . " + cur.StringExt(false, hexValues, bracketNil, compactLists)
				break
			}
		}
		return res + ")"
	}
	_, rightIsAtom := a.Rest.(CLVMAtom)
	rightStr := a.Rest.StringExt(keywords && !rightIsAtom, hexValues, bracketNil, compactLists)
	return "(" + leftStr + " . " + rightStr + ")"
}
func (a CLVMPair) String() string {
	return a.StringExt(true, true, false, false)
}
