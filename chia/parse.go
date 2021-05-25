package chia

import (
	"encoding/binary"
	"math/big"

	"github.com/ansel1/merry"
)

type ParseBuf struct {
	buf []byte
	pos int
	err error
}

func NewParseBuf(buf []byte) *ParseBuf {
	return &ParseBuf{buf: buf, pos: 0, err: nil}
}

func (b *ParseBuf) ensureBytes(n int) bool {
	if b.pos+n > len(b.buf) {
		b.err = merry.Errorf("buffer too short: size=%d, pos=%d, left=%d, need=%d",
			len(b.buf), b.pos, len(b.buf)-b.pos, n)
		return false
	}
	return true
}

func (b *ParseBuf) ensureEmpty() bool {
	if b.pos != len(b.buf) {
		b.err = merry.Errorf("buffer is not empty: size=%d, pos=%d, left=%d",
			len(b.buf), b.pos, len(b.buf)-b.pos)
		return false
	}
	return true
}

func BoolFromBytes(b *ParseBuf) bool {
	if b.err != nil || !b.ensureBytes(1) {
		return false
	}
	v := b.buf[b.pos]
	b.pos += 1
	if v == 0 {
		return false
	}
	if v == 1 {
		return true
	}
	b.err = merry.Errorf("wrong bool: expected 0 or 1, got %d", v)
	return false
}

func Uint8FromBytes(b *ParseBuf) uint8 {
	if b.err != nil || !b.ensureBytes(1) {
		return 0
	}
	v := b.buf[b.pos]
	b.pos += 1
	return v
}

func Uint32FromBytes(b *ParseBuf) uint32 {
	if b.err != nil || !b.ensureBytes(4) {
		return 0
	}
	v := binary.BigEndian.Uint32(b.buf[b.pos:])
	b.pos += 4
	return v
}

func Uint64FromBytes(b *ParseBuf) uint64 {
	if b.err != nil || !b.ensureBytes(8) {
		return 0
	}
	v := binary.BigEndian.Uint64(b.buf[b.pos:])
	b.pos += 8
	return v
}

func Uint128FromBytes(b *ParseBuf) *big.Int {
	if b.err != nil || !b.ensureBytes(16) {
		return &big.Int{}
	}
	v := &big.Int{}
	v.SetBytes(b.buf[b.pos : b.pos+16])
	b.pos += 16
	return v
}

func Bytes32FromBytes(b *ParseBuf) [32]byte {
	if b.err != nil || !b.ensureBytes(32) {
		return [32]byte{}
	}
	var v [32]byte
	copy(v[:], b.buf[b.pos:b.pos+32])
	b.pos += 32
	return v
}

func Bytes100FromBytes(b *ParseBuf) [100]byte {
	if b.err != nil || !b.ensureBytes(100) {
		return [100]byte{}
	}
	var v [100]byte
	copy(v[:], b.buf[b.pos:b.pos+100])
	b.pos += 100
	return v
}

func BytesNFromBytes(b *ParseBuf, n int) []byte {
	v := make([]byte, n)
	copy(v, b.buf[b.pos:b.pos+n])
	b.pos += n
	return v
}

func BytesFromBytes(b *ParseBuf) []byte {
	l := int(Uint32FromBytes(b))
	if b.err != nil || !b.ensureBytes(l) {
		return []byte{}
	}
	return BytesNFromBytes(b, l)
}

const MAX_SINGLE_BYTE = 0x7F
const CONS_BOX_MARKER = 0xFF

type CLVMObject interface {
	IsAtom() bool
}
type CLVMAtom struct {
	Bytes []byte
}

func (a CLVMAtom) IsAtom() bool {
	return true
}

type CLVMPair struct {
	Left  CLVMObject
	Right CLVMObject
}

func (a CLVMPair) IsAtom() bool {
	return false
}

func _op_read_sexp(opStack *[]interface{}, valStack *[]CLVMObject, buf *ParseBuf) {
	b := Uint8FromBytes(buf)
	if b == CONS_BOX_MARKER {
		*opStack = append(*opStack, _op_cons)
		*opStack = append(*opStack, _op_read_sexp)
		*opStack = append(*opStack, _op_read_sexp)
	} else {
		*valStack = append(*valStack, *_atom_from_stream(buf, b))
	}
}

func _op_cons(opStack *[]interface{}, valStack *[]CLVMObject, buf *ParseBuf) {
	l := len(*valStack)
	right := (*valStack)[l-1]
	left := (*valStack)[l-2]
	*valStack = (*valStack)[:l-2]
	*valStack = append(*valStack, CLVMPair{Left: left, Right: right})
}

func _atom_from_stream(buf *ParseBuf, b byte) *CLVMAtom {
	if b == 0x80 {
		return &CLVMAtom{[]byte{}}
	}
	if b <= MAX_SINGLE_BYTE {
		return &CLVMAtom{[]byte{b}}
	}
	bitCount := 0
	bitMask := byte(0x80)
	for b&bitMask > 0 {
		bitCount += 1
		b &= 0xFF ^ bitMask
		bitMask >>= 1
	}
	sizeBlob := []byte{b}
	if bitCount > 1 {
		chunk := BytesNFromBytes(buf, bitCount-1)
		if buf.err != nil {
			buf.err = merry.Prepend(buf.err, "atom from stream: bad encoding")
			return nil
		}
		sizeBlob = append(sizeBlob, chunk...)
	}
	size := uint64(0)
	for _, v := range sizeBlob {
		size = size<<8 + uint64(v)
		if size >= 0x400000000 {
			buf.err = merry.New("atom from stream: blob too large")
			return nil
		}
	}
	blob := BytesNFromBytes(buf, int(size))
	if buf.err != nil {
		buf.err = merry.Prepend(buf.err, "atom from stream: bad encoding")
		return nil
	}
	return &CLVMAtom{blob}
}

type SerializedProgram struct {
	Root CLVMObject
}

// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/util/streamable.py
func SerializedProgramFromBytes(buf *ParseBuf) (obj SerializedProgram) {
	opStack := []interface{}{_op_read_sexp}
	valStack := make([]CLVMObject, 0, 0)

	for len(opStack) > 0 {
		f := opStack[len(opStack)-1].(func(*[]interface{}, *[]CLVMObject, *ParseBuf))
		opStack = opStack[:len(opStack)-1]
		f(&opStack, &valStack, buf)
		if buf.err != nil {
			return
		}
	}
	obj.Root = valStack[0]
	return
}
