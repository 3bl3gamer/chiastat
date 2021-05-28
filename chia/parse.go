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
