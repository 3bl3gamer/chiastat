package chia

import (
	"encoding/hex"
	"log"
	"math/big"
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
	testOk := func(bufHex string, numStr string) {
		buf, err := hex.DecodeString(bufHex)
		if err != nil {
			log.Fatalf("wrong hex: %s", err)
		}
		atom := CLVMAtom{buf}
		resStr := atom.AsInt().String()
		if resStr != numStr {
			t.Errorf("CLVMAtom{%s}.AsInt() = %s, expected %s", bufHex, resStr, numStr)
		}
	}
	testOk("", "0")
	testOk("01", "1")
	testOk("02", "2")
	testOk("7f", "127")
	testOk("0080", "128")
	testOk("0081", "129")
	testOk("00c0", "192")
	testOk("0400", "1024")
	testOk("7f00", "32512")
	testOk("7fff", "32767")
	testOk("008000", "32768")
	testOk("ff", "-1")
	testOk("fe", "-2")
	testOk("81", "-127")
	testOk("80", "-128")
	testOk("ff7f", "-129")
	testOk("8000", "-32768")
	testOk("ff7fff", "-32769")
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
