// Generated with `go genetare`. Do not edit.
package chia

type OperatorInfo struct {
	keyword string
	name    string
	atom    CLVMAtom
	f       func(CLVMObject) CLVMObject
}

var OP_FROM_BYTE = [256]OperatorInfo{
	// core opcodes 0x01-x08
	0x03: OperatorInfo{keyword: "i", name: "if", f: opIf},
	0x04: OperatorInfo{keyword: "c", name: "cons", f: opCons},
	0x05: OperatorInfo{keyword: "f", name: "first", f: opFirst},
	0x06: OperatorInfo{keyword: "r", name: "rest", f: opRest},
	0x07: OperatorInfo{keyword: "l", name: "listp", f: opListp},
	// opcodes on atoms as strings 0x09-0x0f
	0x09: OperatorInfo{keyword: "=", name: "eq", f: opEq},
	0x0a: OperatorInfo{keyword: ">s", name: "gr_bytes", f: opGrBytes},
	0x0b: OperatorInfo{keyword: "sha256", name: "sha256", f: opSha256},
	0x0c: OperatorInfo{keyword: "substr", name: "substr", f: opSubstr},
	0x0e: OperatorInfo{keyword: "concat", name: "concat", f: opConcat},
	// opcodes on atoms as ints 0x10-0x17
	0x10: OperatorInfo{keyword: "+", name: "add", f: opAdd},
}

var ATOM_QUOTE = CLVMAtom{[]byte{0x01}}
var ATOM_APPLY = CLVMAtom{[]byte{0x02}}

var ATOM_FROM_OP_KEYWORD = map[string]CLVMAtom{
	"q":              ATOM_QUOTE,
	"a":              ATOM_APPLY,
	"i":              CLVMAtom{[]byte{0x03}},
	"c":              CLVMAtom{[]byte{0x04}},
	"f":              CLVMAtom{[]byte{0x05}},
	"r":              CLVMAtom{[]byte{0x06}},
	"l":              CLVMAtom{[]byte{0x07}},
	"x":              CLVMAtom{[]byte{0x08}},
	"=":              CLVMAtom{[]byte{0x09}},
	">s":             CLVMAtom{[]byte{0x0a}},
	"sha256":         CLVMAtom{[]byte{0x0b}},
	"substr":         CLVMAtom{[]byte{0x0c}},
	"strlen":         CLVMAtom{[]byte{0x0d}},
	"concat":         CLVMAtom{[]byte{0x0e}},
	"+":              CLVMAtom{[]byte{0x10}},
	"-":              CLVMAtom{[]byte{0x11}},
	"*":              CLVMAtom{[]byte{0x12}},
	"/":              CLVMAtom{[]byte{0x13}},
	"divmod":         CLVMAtom{[]byte{0x14}},
	">":              CLVMAtom{[]byte{0x15}},
	"ash":            CLVMAtom{[]byte{0x16}},
	"lsh":            CLVMAtom{[]byte{0x17}},
	"logand":         CLVMAtom{[]byte{0x18}},
	"logior":         CLVMAtom{[]byte{0x19}},
	"logxor":         CLVMAtom{[]byte{0x1a}},
	"lognot":         CLVMAtom{[]byte{0x1b}},
	"point_add":      CLVMAtom{[]byte{0x1d}},
	"pubkey_for_exp": CLVMAtom{[]byte{0x1e}},
	"not":            CLVMAtom{[]byte{0x20}},
	"any":            CLVMAtom{[]byte{0x21}},
	"all":            CLVMAtom{[]byte{0x22}},
	"softfork":       CLVMAtom{[]byte{0x24}},
}
