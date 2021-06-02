package clvm

import (
	"encoding/hex"
	"log"
	"math/big"
	"strconv"
	"testing"
)

func TestAtomFromInt(t *testing.T) {
	testOk := func(numStr string, bufHex string) {
		v, ok := new(big.Int).SetString(numStr, 10)
		if !ok {
			log.Fatalf("wrong int: %s", numStr)
		}
		resHex := hex.EncodeToString(AtomFromInt(v).Bytes)
		if resHex != bufHex {
			t.Errorf("AtomFromInt(%s) = %s, expected %s", numStr, resHex, bufHex)
		}
	}
	testOk("0", "")
	testOk("1", "01")
	testOk("2", "02")
	testOk("127", "7f")
	testOk("128", "0080")
	testOk("129", "0081")
	testOk("192", "00c0")
	testOk("1024", "0400")
	testOk("32767", "7fff")
	testOk("32512", "7f00")
	testOk("32768", "008000")
	testOk("-0", "")
	testOk("-1", "ff")
	testOk("-2", "fe")
	testOk("-127", "81")
	testOk("-128", "80")
	testOk("-129", "ff7f")
	testOk("-32768", "8000")
	testOk("-32769", "ff7fff")
}

func TestAtomAsInt(t *testing.T) {
	test := func(bufHex string, numStr string, bitLen int) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := Atom{buf}
		resStr := atom.AsInt().String()
		if resStr != numStr {
			t.Errorf("Atom{%s}.AsInt() = %s, expected %s", bufHex, resStr, numStr)
		}
		resBitLen := atom.AsInt().BitLen()
		if bitLen >= 0 && resBitLen != bitLen {
			t.Errorf("Atom{%s}.AsInt().BitLen() = %d, expected %d", bufHex, resBitLen, bitLen)
		}
	}
	test("", "0", 0)
	test("01", "1", 1)
	test("02", "2", 2)
	test("7f", "127", 7)
	test("0080", "128", 8)
	test("0081", "129", 8)
	test("00c0", "192", 8)
	test("0400", "1024", 11)
	test("7f00", "32512", 15)
	test("7fff", "32767", 15)
	test("008000", "32768", 16)
	test("000000008000", "32768", 16)
	test("ff", "-1", 1)
	test("fe", "-2", 2)
	test("81", "-127", 7)
	test("80", "-128", 8)
	test("ff7f", "-129", 8)
	test("8000", "-32768", 16)
	test("ff7fff", "-32769", 16)
}

func TestAtomAsInt32(t *testing.T) {
	test := func(bufHex string, numStr string) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := Atom{buf}
		resVal, resErr := atom.AsInt32()
		var resStr string
		if resErr == nil {
			resStr = strconv.FormatInt(int64(resVal), 10)
		} else {
			resStr = "FAIL: " + resErr.Error()
		}
		if resStr != numStr {
			t.Errorf("Atom{%s}.AsInt32() = %s, expected %s", bufHex, resStr, numStr)
		}
	}
	test("", "0")
	test("01", "1")
	test("02", "2")
	test("7f", "127")
	test("0080", "128")
	test("0081", "129")
	test("00c0", "192")
	test("0400", "1024")
	test("7f00", "32512")
	test("7fff", "32767")
	test("008000", "32768")
	test("00008000", "32768")
	test("ff", "-1")
	test("fe", "-2")
	test("81", "-127")
	test("80", "-128")
	test("ff7f", "-129")
	test("8000", "-32768")
	test("ff7fff", "-32769")
	test("ffffffff", "-1")
	test("7fffffff", "2147483647")
	test("80000000", "-2147483648")
	test("01ffeeddcc", "FAIL: int32 requires 4 bytes at most, got 5: 0x01ffeeddcc: atom=0x01ffeeddcc")
}

func TestAtomAsInt64(t *testing.T) {
	test := func(bufHex string, numStr string) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := Atom{buf}
		resVal, resErr := atom.AsInt64()
		var resStr string
		if resErr == nil {
			resStr = strconv.FormatInt(int64(resVal), 10)
		} else {
			resStr = "FAIL: " + resErr.Error()
		}
		if resStr != numStr {
			t.Errorf("Atom{%s}.AsInt64() = %s, expected %s", bufHex, resStr, numStr)
		}
	}
	test("", "0")
	test("01", "1")
	test("02", "2")
	test("7f", "127")
	test("0080", "128")
	test("0081", "129")
	test("00c0", "192")
	test("0400", "1024")
	test("7f00", "32512")
	test("7fff", "32767")
	test("008000", "32768")
	test("000000008000", "32768")
	test("ff", "-1")
	test("fe", "-2")
	test("81", "-127")
	test("80", "-128")
	test("ff7f", "-129")
	test("8000", "-32768")
	test("ff7fff", "-32769")
	test("ffffffff", "-1")
	test("7fffffff", "2147483647")
	test("80000000", "-2147483648")
	test("01ffeeddcc", "8588811724")
	test("ffffffffffffffff", "-1")
	test("7fffffffffffffff", "9223372036854775807")
	test("8000000000000000", "-9223372036854775808")
	test("01ffeeddccbbaa9988", "FAIL: int64 requires 8 bytes at most, got 9: 0x01ffeeddccbbaa9988: atom=0x01ffeeddccbbaa9988")
}

func TestFromIRString(t *testing.T) {
	test := func(ir string, dest string) {
		res, err := SExpFromIRString(ir)
		var resStr string
		if err == nil {
			resStr = res.StringExt(StringExtCfg{Keywords: true, OnlyHexValues: true, CompactLists: false, Nil: "nil"})
		} else {
			resStr = "FAIL: " + err.Error()
		}
		if resStr != dest {
			t.Errorf("FromIRString '%s' result: %s != %s", ir, resStr, dest)
		}
	}
	test("0", "nil")
	test("1", "01")
	test("127", "7f")
	test("128", "0080")
	test("255", "00ff")
	test("1024", "0400")
	test("-1", "ff")
	test("-32769", "ff7fff")

	test("0x", "nil")
	test("0x0", "00")
	test("0x1", "01")
	test("0xff", "ff")
	test("0x4ff", "04ff")
	test("0x04ff", "04ff")

	test(`""`, `nil`)
	test(`"A"`, `41`)
	test(`"ABC"`, `414243`)
	test(`'ABC'`, `414243`)

	test("q", "01")
	test("a", "02")
	test("i", "03")
	test("if", "6966")

	test("()", "nil")
	test("(())", "(nil . nil)")
	test("((0))", "((nil . nil) . nil)")
	test("(1)", "(q . nil)")
	test("(1024)", "(0400 . nil)")
	test("(1 2)", "(q . (a . nil))")
	test("(1 2 3)", "(q . (a . (i . nil)))")
	test("(1 (2))", "(q . ((a . nil) . nil))")
	test("(i () 1 2)", "(i . (nil . (q . (a . nil))))")

	test("(1 . 2)", "(q . 02)")
	test("(1 . (2))", "(q . (a . nil))")
	test("(1 2 . 3)", "(q . (a . 03))")

	test("   (   1   .   2   )   ", "(q . 02)")

	test("", "FAIL: from ir: unexpected end of string")
	test("(", "FAIL: from ir: unexpected end of string")
	test("(1", "FAIL: from ir: unexpected end of string")
	test("(1 .", "FAIL: from ir: unexpected end of string")
	test("(1 . 2", "FAIL: from ir: unexpected end of string")
	test(")", "FAIL: from ir: unexpected ')' at pos 1")
	test("1)", "FAIL: from ir: extra characters in string with len 2 after pos 2")
	test(". 2)", "FAIL: from ir: unexpected '.' at pos 1")
	test("1 . 2)", "FAIL: from ir: extra characters in string with len 6 after pos 3")
	test("(. 1 . 2)", "FAIL: from ir: unexpected '.' at pos 2")
	test("(1 . . 2)", "FAIL: from ir: unexpected '.' at pos 6")
	test("(1 . 2 . )", "FAIL: from ir: unexpected '.' at pos 8")
	test("(1 . 2 . 3)", "FAIL: from ir: unexpected '.' at pos 8")
	test("(1 . 2 . 3)", "FAIL: from ir: unexpected '.' at pos 8")
	test("(1 . 2 3)", "FAIL: from ir: unexpected '.' at pos 4")
}
