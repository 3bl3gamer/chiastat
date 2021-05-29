// Generated with `go genetare`. Do not edit.
package chia

type OperatorInfo struct {
	keyword string
	name    string
	atom    CLVMAtom
	f       func(CLVMObject) (int64, CLVMObject, error)
}

var OP_FROM_BYTE = [256]OperatorInfo{
	// core opcodes 0x01-x08
	0x01: {keyword: "q", name: "quote", f: nil},
	0x02: {keyword: "a", name: "apply", f: nil},
	0x03: {keyword: "i", name: "if", f: opIf},
	0x04: {keyword: "c", name: "cons", f: opCons},
	0x05: {keyword: "f", name: "first", f: opFirst},
	0x06: {keyword: "r", name: "rest", f: opRest},
	0x07: {keyword: "l", name: "listp", f: opListp},
	// opcodes on atoms as strings 0x09-0x0f
	0x09: {keyword: "=", name: "eq", f: opEq},
	0x0a: {keyword: ">s", name: "gr_bytes", f: opGrBytes},
	0x0b: {keyword: "sha256", name: "sha256", f: opSha256},
	0x0c: {keyword: "substr", name: "substr", f: opSubstr},
	0x0e: {keyword: "concat", name: "concat", f: opConcat},
	// opcodes on atoms as ints 0x10-0x17
	0x10: {keyword: "+", name: "add", f: opAdd},
}

var ATOM_QUOTE = CLVMAtom{[]byte{0x01}}
var ATOM_APPLY = CLVMAtom{[]byte{0x02}}

var ATOM_FROM_OP_KEYWORD = map[string]CLVMAtom{
	"q":              ATOM_QUOTE,
	"a":              ATOM_APPLY,
	"i":              {[]byte{0x03}},
	"c":              {[]byte{0x04}},
	"f":              {[]byte{0x05}},
	"r":              {[]byte{0x06}},
	"l":              {[]byte{0x07}},
	"x":              {[]byte{0x08}},
	"=":              {[]byte{0x09}},
	">s":             {[]byte{0x0a}},
	"sha256":         {[]byte{0x0b}},
	"substr":         {[]byte{0x0c}},
	"strlen":         {[]byte{0x0d}},
	"concat":         {[]byte{0x0e}},
	"+":              {[]byte{0x10}},
	"-":              {[]byte{0x11}},
	"*":              {[]byte{0x12}},
	"/":              {[]byte{0x13}},
	"divmod":         {[]byte{0x14}},
	">":              {[]byte{0x15}},
	"ash":            {[]byte{0x16}},
	"lsh":            {[]byte{0x17}},
	"logand":         {[]byte{0x18}},
	"logior":         {[]byte{0x19}},
	"logxor":         {[]byte{0x1a}},
	"lognot":         {[]byte{0x1b}},
	"point_add":      {[]byte{0x1d}},
	"pubkey_for_exp": {[]byte{0x1e}},
	"not":            {[]byte{0x20}},
	"any":            {[]byte{0x21}},
	"all":            {[]byte{0x22}},
	"softfork":       {[]byte{0x24}},
}
