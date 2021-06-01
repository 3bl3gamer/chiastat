package utils

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

func (b ParseBuf) Pos() int {
	return b.pos
}

func (b *ParseBuf) SeekSet(offset int) int {
	prev := b.pos
	b.pos = offset
	return prev
}

func (b ParseBuf) Slice(start, end int) []byte {
	return b.buf[start:end]
}

func (b ParseBuf) Copy(start, end int) []byte {
	res := make([]byte, end-start)
	copy(res, b.buf[start:end])
	return res
}

func (b ParseBuf) Err() error {
	return b.err
}

func (b *ParseBuf) SetErr(err error) {
	b.err = err
}

func (b *ParseBuf) PrependErr(msg string) {
	b.err = merry.Prepend(b.err, msg)
}

func (b *ParseBuf) EnsureBytes(n int) bool {
	if b.pos+n > len(b.buf) {
		b.err = merry.Errorf("buffer too short: size=%d, pos=%d, left=%d, need=%d",
			len(b.buf), b.pos, len(b.buf)-b.pos, n)
		return false
	}
	return true
}

func (b *ParseBuf) EnsureEmpty() bool {
	if b.pos != len(b.buf) {
		b.err = merry.Errorf("buffer is not empty: size=%d, pos=%d, left=%d",
			len(b.buf), b.pos, len(b.buf)-b.pos)
		return false
	}
	return true
}

func (b *ParseBuf) Bool() bool {
	if b.err != nil || !b.EnsureBytes(1) {
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

func (b *ParseBuf) Uint8() uint8 {
	if b.err != nil || !b.EnsureBytes(1) {
		return 0
	}
	v := b.buf[b.pos]
	b.pos += 1
	return v
}

func (b *ParseBuf) Uint32() uint32 {
	if b.err != nil || !b.EnsureBytes(4) {
		return 0
	}
	v := binary.BigEndian.Uint32(b.buf[b.pos:])
	b.pos += 4
	return v
}

func (b *ParseBuf) Uint64() uint64 {
	if b.err != nil || !b.EnsureBytes(8) {
		return 0
	}
	v := binary.BigEndian.Uint64(b.buf[b.pos:])
	b.pos += 8
	return v
}

func (b *ParseBuf) Uint128() *big.Int {
	if b.err != nil || !b.EnsureBytes(16) {
		return &big.Int{}
	}
	v := &big.Int{}
	v.SetBytes(b.buf[b.pos : b.pos+16])
	b.pos += 16
	return v
}

func (b *ParseBuf) Bytes32() [32]byte {
	if b.err != nil || !b.EnsureBytes(32) {
		return [32]byte{}
	}
	var v [32]byte
	copy(v[:], b.buf[b.pos:b.pos+32])
	b.pos += 32
	return v
}

func (b *ParseBuf) Bytes100() [100]byte {
	if b.err != nil || !b.EnsureBytes(100) {
		return [100]byte{}
	}
	var v [100]byte
	copy(v[:], b.buf[b.pos:b.pos+100])
	b.pos += 100
	return v
}

func (b *ParseBuf) BytesN(n int) []byte {
	v := make([]byte, n)
	copy(v, b.buf[b.pos:b.pos+n])
	b.pos += n
	return v
}

func (b *ParseBuf) Bytes() []byte {
	l := int(b.Uint32())
	if b.err != nil || !b.EnsureBytes(l) {
		return []byte{}
	}
	return b.BytesN(l)
}
