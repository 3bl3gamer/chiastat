package chia

import (
	"encoding/hex"
	"log"
	"math/big"
	"strconv"
	"testing"
)

func TestCLVMAtomFromInt(t *testing.T) {
	testOk := func(numStr string, bufHex string) {
		v, ok := new(big.Int).SetString(numStr, 10)
		if !ok {
			log.Fatalf("wrong int: %s", numStr)
		}
		resHex := hex.EncodeToString(CLVMAtomFromInt(v).Bytes)
		if resHex != bufHex {
			t.Errorf("CLVMAtomFromInt(%s) = %s, expected %s", numStr, resHex, bufHex)
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

func TestCLVMAtomAsInt(t *testing.T) {
	test := func(bufHex string, numStr string, bitLen int) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := CLVMAtom{buf}
		resStr := atom.AsInt().String()
		if resStr != numStr {
			t.Errorf("CLVMAtom{%s}.AsInt() = %s, expected %s", bufHex, resStr, numStr)
		}
		resBitLen := atom.AsInt().BitLen()
		if bitLen >= 0 && resBitLen != bitLen {
			t.Errorf("CLVMAtom{%s}.AsInt().BitLen() = %d, expected %d", bufHex, resBitLen, bitLen)
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

func TestCLVMAtomAsInt32(t *testing.T) {
	test := func(bufHex string, numStr string) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := CLVMAtom{buf}
		resVal, resErr := atom.AsInt32()
		var resStr string
		if resErr == nil {
			resStr = strconv.FormatInt(int64(resVal), 10)
		} else {
			resStr = "FAIL: " + resErr.Error()
		}
		if resStr != numStr {
			t.Errorf("CLVMAtom{%s}.AsInt32() = %s, expected %s", bufHex, resStr, numStr)
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

func TestCLVMAtomAsInt64(t *testing.T) {
	test := func(bufHex string, numStr string) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := CLVMAtom{buf}
		resVal, resErr := atom.AsInt64()
		var resStr string
		if resErr == nil {
			resStr = strconv.FormatInt(int64(resVal), 10)
		} else {
			resStr = "FAIL: " + resErr.Error()
		}
		if resStr != numStr {
			t.Errorf("CLVMAtom{%s}.AsInt64() = %s, expected %s", bufHex, resStr, numStr)
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

func TestCLVMFromIRString(t *testing.T) {
	testOk := func(ir string, dest string) {
		res, err := CLVMFromIRString(ir)
		if err != nil {
			t.Errorf("CLVMFromIRString '%s' failed: %s", ir, err)
		}
		resStr := res.StringExt(CLVMStringExtCfg{Keywords: true, OnlyHexValues: true, CompactLists: false, Nil: "nil"})
		if resStr != dest {
			t.Errorf("CLVMFromIRString '%s' result: %s != %s", ir, resStr, dest)
		}
	}
	testOk("0", "nil")
	testOk("1", "01")
	testOk("127", "7f")
	testOk("128", "0080")
	testOk("255", "00ff")
	testOk("1024", "0400")
	testOk("-1", "ff")
	testOk("-32769", "ff7fff")

	testOk("0x", "nil")
	testOk("0x0", "00")
	testOk("0x1", "01")
	testOk("0xff", "ff")
	testOk("0x4ff", "04ff")
	testOk("0x04ff", "04ff")

	testOk(`""`, `nil`)
	testOk(`"A"`, `41`)
	testOk(`"ABC"`, `414243`)
	testOk(`'ABC'`, `414243`)

	testOk("q", "01")
	testOk("a", "02")
	testOk("i", "03")
	testOk("if", "6966")

	testOk("()", "nil")
	testOk("(())", "(nil . nil)")
	testOk("((0))", "((nil . nil) . nil)")
	testOk("(1)", "(q . nil)")
	testOk("(1024)", "(0400 . nil)")
	testOk("(1 2)", "(q . (a . nil))")
	testOk("(1 2 3)", "(q . (a . (i . nil)))")
	testOk("(1 (2))", "(q . ((a . nil) . nil))")
	testOk("(i () 1 2)", "(i . (nil . (q . (a . nil))))")

	testOk("(1 . 2)", "(q . 02)")
	testOk("(1 . (2))", "(q . (a . nil))")
}
