package chia

import (
	"strings"

	"github.com/ansel1/merry"
)

const CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

var CHARSET_MAP = map[rune]byte{}

func init() {
	for i, c := range CHARSET {
		CHARSET_MAP[c] = byte(i)
	}
}

// Internal function that computes the Bech32 checksum.
func bech32Polymod(values []byte) uint32 {
	generator := []uint32{0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3}
	chk := uint32(1)
	for _, value := range values {
		top := chk >> 25
		chk = (chk&0x1FFFFFF)<<5 ^ uint32(value)
		for i := 0; i < 5; i++ {
			if (top>>i)&1 == 1 {
				chk ^= generator[i]
			}
		}
	}
	return chk
}

// Expand the HRP into values for checksum computation.
func bech32HrpExpand(hrp string) []byte {
	hrpBuf := []byte(hrp)
	res := make([]byte, len(hrpBuf)*2+1)
	for i, c := range hrpBuf {
		res[i] = c >> 5
		res[i+1+len(hrpBuf)] = c & 31
	}
	return res
}

func byteConcat(a, b []byte) []byte {
	res := make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return res
}

const M uint32 = 0x2BC830A3

func bech32VerifyChecksum(hrp string, data []byte) bool {
	return bech32Polymod(byteConcat(bech32HrpExpand(hrp), data)) == M
}

func bech32CreateChecksum(hrp string, data []byte) []byte {
	values := byteConcat(bech32HrpExpand(hrp), data)
	polymod := bech32Polymod(byteConcat(values, []byte{0, 0, 0, 0, 0, 0})) ^ M
	checksum := make([]byte, 6)
	for i := uint32(0); i < 6; i++ {
		checksum[i] = byte((polymod >> (5 * (5 - i))) & 31)
	}
	return checksum
}

// Compute a Bech32 string given HRP and data values.
func Bech32Encode(hrp string, data []byte) string {
	combined := byteConcat(data, bech32CreateChecksum(hrp, data))
	for i, c := range combined {
		combined[i] = CHARSET[c]
	}
	return hrp + "1" + string(combined)
}

// Validate a Bech32 string, and determine HRP and data.
func Bech32Decode(bech string) (string, []byte, error) {
	for _, c := range bech {
		if c < 33 || c > 126 {
			return "", nil, merry.Errorf("bech decode: character '%c' not in charset", c)
		}
	}
	bech = strings.ToLower(bech)
	pos := strings.LastIndex(bech, "1")
	if pos < 1 || pos+7 > len(bech) || len(bech) > 90 {
		return "", nil, merry.Errorf("bech decode: invalid hrp end index %d", pos)
	}
	hrp := bech[:pos]
	dataStr := bech[pos+1:]
	data := make([]byte, len(dataStr))
	for i, c := range dataStr {
		var ok bool
		data[i], ok = CHARSET_MAP[c]
		if !ok {
			return "", nil, merry.Errorf("bech decode: character '%c' not in charset", c)
		}
	}
	if !bech32VerifyChecksum(hrp, data) {
		return "", nil, merry.Errorf("bech decode: checksum mismatch")
	}
	return hrp, data[:len(data)-6], nil
}

// General power-of-2 base conversion.
func convertbits(data []byte, frombits, tobits byte, pad bool) ([]byte, error) {
	acc := 0
	bits := 0
	ret := []byte{}
	maxv := (1 << tobits) - 1
	maxAcc := (1 << (frombits + tobits - 1)) - 1
	for _, value := range data {
		if value < 0 || (value>>frombits) > 0 {
			return nil, merry.Errorf("convertbits: invalid value %d", value)
		}
		acc = ((acc << frombits) | int(value)) & maxAcc
		bits += int(frombits)
		for bits >= int(tobits) {
			bits -= int(tobits)
			ret = append(ret, byte((acc>>bits)&maxv))
		}
	}
	if pad {
		if bits > 0 {
			ret = append(ret, byte((acc<<(int(tobits)-bits))&maxv))
		}
	} else if bits >= int(frombits) || (acc<<(int(tobits)-bits))&maxv > 0 {
		return nil, merry.Errorf("convertbits: invalid bits")
	}
	return ret, nil
}

func EncodePuzzleHash(puzzleHash [32]byte, prefix string) string {
	// should not fail (frombits = 8, padding enabled)
	data, _ := convertbits(puzzleHash[:], 8, 5, true)
	return Bech32Encode(prefix, data)
}

func DecodePuzzleHash(address string) (string, [32]byte, error) {
	hrp, data, err := Bech32Decode(address)
	if err != nil {
		return "", [32]byte{}, merry.Wrap(err)
	}
	decoded, err := convertbits(data, 5, 8, false)
	if err != nil {
		return "", [32]byte{}, merry.Wrap(err)
	}
	if len(decoded) != 32 {
		return "", [32]byte{}, merry.Errorf("decode puzzle hash: wrong puzzle hash length %d, expected 32", len(decoded))
	}
	var res [32]byte
	copy(res[:], decoded)
	return hrp, res, nil
}
