package chia

import "log"

func SerializeAtomBytes(buf []byte) []byte {
	size := len(buf)
	if size == 0 {
		return []byte{0x80}
	}
	if size == 1 {
		if buf[0] <= MAX_SINGLE_BYTE {
			return []byte{buf[0]}
		}
	}
	var sizeBuf []byte
	if size < 0x40 {
		sizeBuf = []byte{0x80 | byte(size)}
	} else if size < 0x2000 {
		sizeBuf = []byte{0xC0 | byte(size>>8), byte(size>>0) & 0xFF}
	} else if size < 0x100000 {
		sizeBuf = []byte{0xE0 | byte(size>>16), byte(size>>8) & 0xFF, byte(size>>0) & 0xFF}
	} else if size < 0x8000000 {
		sizeBuf = []byte{
			0xF0 | byte(size>>24),
			byte(size>>16) & 0xFF,
			byte(size>>8) & 0xFF,
			byte(size>>0) & 0xFF,
		}
	} else if size < 0x400000000 {
		sizeBuf = []byte{
			0xF8 | byte(size>>32),
			byte(size>>24) & 0xFF,
			byte(size>>16) & 0xFF,
			byte(size>>8) & 0xFF,
			byte(size>>0) & 0xFF,
		}
	} else {
		log.Fatalf("atom buf too long: %d", len(buf))
	}
	res := make([]byte, len(sizeBuf)+size)
	copy(res, sizeBuf)
	copy(res[len(sizeBuf):], buf)
	return res
}
