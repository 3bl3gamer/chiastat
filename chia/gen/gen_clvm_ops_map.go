package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func opFuncName(s string) string {
	s = strings.Replace(s, "_", " ", -1)
	s = "op" + strings.Title(s)
	s = strings.Replace(s, " ", "", -1)
	return s
}

func atomConstName(s string) string {
	return "ATOM_" + strings.ToUpper(s)
}

type Keyword struct {
	groupIndex   int
	code         byte
	kw           string
	name         string
	groupComment string
}

func main() {
	// https://github.com/Chia-Network/clvm/blob/main/clvm/operators.py
	keywordStrGroups := []struct{ keywords, comment string }{
		{". q a i c f r l x", "core opcodes 0x01-x08"},
		{"= >s sha256 substr strlen concat .", "opcodes on atoms as strings 0x09-0x0f"},
		{"+ - * / divmod > ash lsh", "opcodes on atoms as ints 0x10-0x17"},
		{"logand logior logxor lognot .", "opcodes on atoms as vectors of bools 0x18-0x1c"},
		{"point_add pubkey_for_exp .", "opcodes for bls 1381 0x1d-0x1f"},
		{"not any all .", "bool opcodes 0x20-0x23"},
		{"softfork", "misc 0x24"},
	}
	var opRename = map[string]string{
		"q":  "quote",
		"a":  "apply",
		"+":  "add",
		"-":  "subtract",
		"*":  "multiply",
		"/":  "div",
		"i":  "if",
		"c":  "cons",
		"f":  "first",
		"r":  "rest",
		"l":  "listp",
		"x":  "raise",
		"=":  "eq",
		">":  "gr",
		">s": "gr_bytes",
	}
	var kwNilListFunc = map[string]bool{
		"q": true,
		"a": true,
	}
	var kwOmitFromList = map[string]bool{
		//TODO: implement
		"x":              true,
		"strlen":         true,
		"-":              true,
		"*":              true,
		"/":              true,
		"divmod":         true,
		">":              true,
		"ash":            true,
		"lsh":            true,
		"logand":         true,
		"logior":         true,
		"logxor":         true,
		"lognot":         true,
		"point_add":      true,
		"pubkey_for_exp": true,
		"not":            true,
		"any":            true,
		"all":            true,
		"softfork":       true,
	}
	var kwAsConstAtom = map[string]bool{
		"q": true,
		"a": true,
	}

	keywords := []Keyword{}
	code := 0
	for groupIndex, groupStr := range keywordStrGroups {
		for _, kw := range strings.Split(strings.TrimSpace(groupStr.keywords), " ") {
			name := kw
			if rn, ok := opRename[name]; ok {
				name = rn
			}
			keywords = append(keywords, Keyword{
				groupIndex:   groupIndex,
				code:         byte(code),
				kw:           kw,
				name:         name,
				groupComment: groupStr.comment,
			})
			code += 1
		}
	}

	outFName := flag.String("fname", "", "output file name")
	flag.Parse()
	if *outFName == "" {
		log.Fatal("-fname flag is required")
	}

	outFile, err := os.Create(*outFName)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	write := func(format string, a ...interface{}) {
		if _, err := fmt.Fprintf(outFile, format, a...); err != nil {
			log.Fatal(err)
		}
	}

	write("// Generated with `go genetare`. Do not edit.\n")
	write("package chia\n\n")

	write(`type OperatorInfo struct {
	keyword string
	name string
	atom CLVMAtom
	f func(CLVMObject) (int64, CLVMObject, error)
}
`)

	write("var OP_FROM_BYTE = [256]OperatorInfo {\n")
	prevGroupI := -1
	for _, keyword := range keywords {
		if _, ok := kwOmitFromList[keyword.kw]; !ok && keyword.kw != "." {
			if keyword.groupIndex != prevGroupI {
				write("// %s\n", keyword.groupComment)
				prevGroupI = keyword.groupIndex
			}
			funcName := "nil"
			if _, ok := kwNilListFunc[keyword.kw]; !ok {
				funcName = opFuncName(keyword.name)
			}
			write("0x%02x: {keyword: \"%s\", name: \"%s\", f: %s},\n",
				keyword.code, keyword.kw, keyword.name, funcName)
		}
	}
	write("}\n")
	write("\n")

	for _, keyword := range keywords {
		if _, ok := kwAsConstAtom[keyword.kw]; ok {
			write("var %s = CLVMAtom{[]byte{0x%02x}}\n", atomConstName(keyword.name), keyword.code)
		}
	}
	write("\n")

	write("var ATOM_FROM_OP_KEYWORD = map[string]CLVMAtom {\n")
	for _, keyword := range keywords {
		if keyword.kw != "." {
			if _, ok := kwAsConstAtom[keyword.kw]; ok {
				write("\"%s\": %s,\n", keyword.kw, atomConstName(keyword.name))
			} else {
				write("\"%s\": {[]byte{0x%02x}},\n", keyword.kw, keyword.code)
			}
		}
	}
	write("}\n")
	write("\n")

	if err := outFile.Close(); err != nil {
		log.Fatal(err)
	}
}
