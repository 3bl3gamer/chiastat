package clvm

import (
	"chiastat/chia/utils"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ansel1/merry"
)

const MAX_SINGLE_BYTE = 0x7F
const CONS_BOX_MARKER = 0xFF

func _op_read_sexp(opStack *[]interface{}, valStack *[]SExp, buf *utils.ParseBuf) {
	b := buf.Uint8()
	if b == CONS_BOX_MARKER {
		*opStack = append(*opStack, _op_cons)
		*opStack = append(*opStack, _op_read_sexp)
		*opStack = append(*opStack, _op_read_sexp)
	} else {
		*valStack = append(*valStack, *_atom_from_stream(buf, b))
	}
}

func _op_cons(opStack *[]interface{}, valStack *[]SExp, buf *utils.ParseBuf) {
	l := len(*valStack)
	right := (*valStack)[l-1]
	left := (*valStack)[l-2]
	*valStack = (*valStack)[:l-2]
	*valStack = append(*valStack, Pair{First: left, Rest: right})
}

func _atom_from_stream(buf *utils.ParseBuf, b byte) *Atom {
	if b == 0x80 {
		return &NULL
	}
	if b <= MAX_SINGLE_BYTE {
		return &Atom{[]byte{b}}
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
		chunk := buf.BytesN(bitCount - 1)
		if buf.Err() != nil {
			buf.PrependErr("atom from stream: bad encoding")
			return nil
		}
		sizeBlob = append(sizeBlob, chunk...)
	}
	size := uint64(0)
	for _, v := range sizeBlob {
		size = size<<8 + uint64(v)
		if size >= 0x400000000 {
			buf.SetErr(merry.New("atom from stream: blob too large"))
			return nil
		}
	}
	blob := buf.BytesN(int(size))
	if buf.Err() != nil {
		buf.PrependErr("atom from stream: bad encoding")
		return nil
	}
	return &Atom{blob}
}

// https://github.com/Chia-Network/clvm/blob/main/clvm/serialize.py
func SExpFromBytes(buf *utils.ParseBuf) SExp {
	opStack := []interface{}{_op_read_sexp}
	valStack := make([]SExp, 0, 0)

	for len(opStack) > 0 {
		f := opStack[len(opStack)-1].(func(*[]interface{}, *[]SExp, *utils.ParseBuf))
		opStack = opStack[:len(opStack)-1]
		f(&opStack, &valStack, buf)
		if buf.Err() != nil {
			return nil
		}
	}
	return valStack[0]
}

func readIRSpaces(str string, pos int) int {
	for pos < len(str) && str[pos] == ' ' {
		pos += 1
	}
	return pos
}
func readIRToken(str string, pos int) (string, int) {
	startPos := pos
	for pos < len(str) {
		c := str[pos]
		if c == '(' || c == ')' || c == ' ' {
			break
		}
		pos += 1
	}
	return str[startPos:pos], pos
}
func SExpNextFromIRString(str string, pos int) (int, SExp, error) {
	var stack []interface{}
	for {
		if len(stack) == 1 {
			if obj, ok := stack[0].(SExp); ok {
				return pos, obj, nil
			}
		}

		pos = readIRSpaces(str, pos)
		if pos >= len(str) {
			break
		}

		c := str[pos]
		if c == '(' {
			stack = append(stack, byte('('))
		} else if c == ')' {
			i := len(stack) - 1
			for ; i >= 0; i -= 1 {
				if sc, ok := stack[i].(byte); ok && sc == '(' {
					break
				}
			}
			if i < 0 {
				return pos, nil, merry.Errorf("from ir: %d", i)
			}
			items := stack[i+1:]
			stack = stack[:i]
			var list SExp = NULL
			shouldCons := false
			for i := len(items) - 1; i >= 0; i -= 1 {
				item := items[i]
				if obj, ok := item.(SExp); ok {
					if shouldCons {
						list = Pair{First: obj, Rest: list.(Pair).First}
						shouldCons = false
					} else {
						list = Pair{First: obj, Rest: list}
					}
				} else if ic, ok := item.(byte); ok && ic == '.' {
					shouldCons = true
				} else {
					return pos, nil, merry.Errorf("from ir: unexpected stack item: %#v", item)
				}
			}
			stack = append(stack, list)
		} else if c == '.' {
			stack = append(stack, byte('.'))
		} else {
			tokenStartPos := pos
			var token string
			token, pos = readIRToken(str, pos)
			pos -= 1
			if vInt, ok := (&big.Int{}).SetString(token, 10); ok {
				// int
				stack = append(stack, AtomFromInt(vInt))
			} else if strings.HasPrefix(token, "0x") || strings.HasPrefix(token, "0X") {
				// hex
				hexChars := []byte(token)[2:]
				if len(token)%2 == 1 {
					hexChars = []byte(token)[1:]
					hexChars[0] = '0'
				}
				buf := make([]byte, len(hexChars)/2)
				if _, err := hex.Decode(buf, hexChars); err != nil {
					return pos, nil, merry.Errorf("from ir: invalid hex at %d: %s", tokenStartPos, token)
				}
				stack = append(stack, Atom{buf})
			} else if len(token) >= 2 && (token[0] == '\'' || token[0] == '"') {
				// quoted string
				if token[len(token)-1] != token[0] {
					return pos, nil, merry.Errorf("from ir: unterminated string starting at %d: %s", tokenStartPos, token)
				}
				stack = append(stack, Atom{[]byte(token[1 : len(token)-1])})
			} else {
				// symbol
				if atom, ok := ATOM_FROM_OP_KEYWORD[token]; ok {
					stack = append(stack, atom)
				} else {
					stack = append(stack, Atom{[]byte(token)})
				}
			}
		}
		pos += 1
	}
	return pos, nil, merry.Errorf("from ir: unexpected string end")
}

func SExpFromIRString(str string) (SExp, error) {
	pos, obj, err := SExpNextFromIRString(str, 0)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	pos = readIRSpaces(str, pos)
	if pos != len(str) {
		return obj, merry.Errorf("from ir: extra characters in string with len %d after pos %d", len(str), pos)
	}
	return obj, nil
}

func SExpOneOrTwoFromIRString(str string) (SExp, SExp, error) {
	pos, objA, err := SExpNextFromIRString(str, 0)
	if err != nil {
		return nil, nil, merry.Wrap(err)
	}

	var objB SExp = NULL
	pos = readIRSpaces(str, pos)
	if pos < len(str) {
		pos, objB, err = SExpNextFromIRString(str, pos)
		if err != nil {
			return objA, nil, merry.Wrap(err)
		}
		pos = readIRSpaces(str, pos)
	}

	if pos != len(str) {
		return objA, objB, merry.Errorf("from ir: extra characters in string with len %d after pos %d", len(str), pos)
	}
	return objA, objB, nil
}

func MustSExpFromHex(hexStr string) SExp {
	byteBuf, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	buf := utils.NewParseBuf(byteBuf)
	prog := SExpFromBytes(buf)
	buf.EnsureEmpty()
	if buf.Err() != nil {
		panic(buf.Err())
	}
	return prog
}
