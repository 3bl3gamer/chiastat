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

func _opReadSExp(opStack *[]interface{}, valStack *[]SExp, buf *utils.ParseBuf) {
	b := buf.Uint8()
	if b == CONS_BOX_MARKER {
		*opStack = append(*opStack, _opCons)
		*opStack = append(*opStack, _opReadSExp)
		*opStack = append(*opStack, _opReadSExp)
	} else {
		*valStack = append(*valStack, *_atomFromBytes(buf, b))
	}
}

func _opCons(opStack *[]interface{}, valStack *[]SExp, buf *utils.ParseBuf) {
	l := len(*valStack)
	right := (*valStack)[l-1]
	left := (*valStack)[l-2]
	*valStack = (*valStack)[:l-2]
	*valStack = append(*valStack, Pair{First: left, Rest: right})
}

func _atomFromBytes(buf *utils.ParseBuf, b byte) *Atom {
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
	opStack := []interface{}{_opReadSExp}
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

// sexp is nil or isListStart is true
type irStackItem struct {
	sexp          SExp
	isListStart   bool
	consWithNext  bool
	consStringPos int
}

func irReadSpaces(str string, pos int) int {
	for pos < len(str) && str[pos] == ' ' {
		pos += 1
	}
	return pos
}
func irReadToken(str string, pos int) (string, int) {
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
func irNonConsSExpOnTop(stack []irStackItem) bool {
	return len(stack) > 0 && stack[len(stack)-1].sexp != nil && !stack[len(stack)-1].consWithNext
}
func irSingleValueList(sexp SExp) bool {
	if pair, ok := sexp.(Pair); ok {
		return pair.Rest.Nullp()
	}
	return false
}
func irPopListFromStack(stack *[]irStackItem) (SExp, error) {
	var list SExp = NULL
	lastConsPos := 0
	for i := len(*stack) - 1; i >= 0; i -= 1 {
		item := (*stack)[i]
		if item.sexp != nil {
			if item.consWithNext {
				if lastConsPos == 0 {
					lastConsPos = item.consStringPos
				}
				if !irSingleValueList(list) {
					return nil, merry.Errorf("from ir: unexpected '.' at pos %d", lastConsPos+1)
				}
				list = Pair{First: item.sexp, Rest: list.(Pair).First}
			} else {
				list = Pair{First: item.sexp, Rest: list}
			}
		} else if item.isListStart {
			*stack = (*stack)[:i]
			break
		}
	}
	return list, nil
}
func irReadAtom(str string, pos int) (int, Atom, error) {
	tokenStartPos := pos
	var token string
	token, pos = irReadToken(str, pos)
	if vInt, ok := (&big.Int{}).SetString(token, 10); ok {
		// int
		return pos, AtomFromInt(vInt), nil
	}
	if strings.HasPrefix(token, "0x") || strings.HasPrefix(token, "0X") {
		// hex
		hexChars := []byte(token)[2:]
		if len(token)%2 == 1 {
			hexChars = []byte(token)[1:]
			hexChars[0] = '0'
		}
		buf := make([]byte, len(hexChars)/2)
		if _, err := hex.Decode(buf, hexChars); err != nil {
			return pos, Atom{}, merry.Errorf("from ir: invalid hex at pos %d: %s", tokenStartPos, token)
		}
		return pos, Atom{buf}, nil
	}
	if len(token) >= 2 && (token[0] == '\'' || token[0] == '"') {
		// quoted string
		if token[len(token)-1] != token[0] {
			return pos, Atom{}, merry.Errorf("from ir: unterminated string starting at pos %d: %s", tokenStartPos, token)
		}
		return pos, Atom{[]byte(token[1 : len(token)-1])}, nil
	}
	// symbol (as operator)
	if atom, ok := ATOM_FROM_OP_KEYWORD[token]; ok {
		return pos, atom, nil
	}
	// symbol (as string)
	return pos, Atom{[]byte(token)}, nil
}
func SExpNextFromIRString(str string, pos int) (int, SExp, error) {
	var err error
	var stack []irStackItem
	for {
		if len(stack) == 1 {
			if stack[0].sexp != nil {
				return pos, stack[0].sexp, nil
			}
		}

		pos = irReadSpaces(str, pos)
		if pos >= len(str) {
			break
		}

		c := str[pos]
		if c == '(' {
			stack = append(stack, irStackItem{isListStart: true})
			pos += 1
		} else if c == ')' {
			if len(stack) == 0 {
				return pos, nil, merry.Errorf("from ir: unexpected ')' at pos %d", pos+1)
			}
			list, err := irPopListFromStack(&stack)
			if err != nil {
				return pos, nil, err
			}
			stack = append(stack, irStackItem{sexp: list})
			pos += 1
		} else if c == '.' {
			if !irNonConsSExpOnTop(stack) {
				return pos, nil, merry.Errorf("from ir: unexpected '.' at pos %d", pos+1)
			}
			stack[len(stack)-1].consWithNext = true
			stack[len(stack)-1].consStringPos = pos
			pos += 1
		} else {
			var atom Atom
			pos, atom, err = irReadAtom(str, pos)
			if err != nil {
				return pos, nil, err
			}
			stack = append(stack, irStackItem{sexp: atom})
		}
	}
	return pos, nil, merry.Errorf("from ir: unexpected end of string")
}

func SExpFromIRString(str string) (SExp, error) {
	pos, obj, err := SExpNextFromIRString(str, 0)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	pos = irReadSpaces(str, pos)
	if pos != len(str) {
		return obj, merry.Errorf("from ir: extra characters in string with len %d after pos %d", len(str), pos+1)
	}
	return obj, nil
}

func SExpOneOrTwoFromIRString(str string) (SExp, SExp, error) {
	pos, objA, err := SExpNextFromIRString(str, 0)
	if err != nil {
		return nil, nil, merry.Wrap(err)
	}

	var objB SExp = NULL
	pos = irReadSpaces(str, pos)
	if pos < len(str) {
		pos, objB, err = SExpNextFromIRString(str, pos)
		if err != nil {
			return objA, nil, merry.Wrap(err)
		}
		pos = irReadSpaces(str, pos)
	}

	if pos != len(str) {
		return objA, objB, merry.Errorf("from ir: extra characters in string with len %d after pos %d", len(str), pos+1)
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
