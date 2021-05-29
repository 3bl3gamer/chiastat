package chia

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

//go:generate go run gen/gen_clvm_ops_map.go -fname clvm_ops_map_generated.go
//go:generate go fmt clvm_ops_map_generated.go

const RUN_DEBUG = false

func mallocCost(cost int64, atom CLVMAtom) (int64, CLVMAtom) {
	return cost + int64(len(atom.Bytes))*MALLOC_COST_PER_BYTE, atom
}

func opIf(args CLVMObject) CLVMObject {
	if args.ListLen() != 3 {
		log.Fatalf("i takes exactly 3 arguments: %s", args)
	}
	r := args.(CLVMPair).Rest.(CLVMPair)
	if args.(CLVMPair).First.Nullp() {
		return r.Rest.(CLVMPair).First //IF_COST
	}
	return r.First //IF_COST
}
func opCons(args CLVMObject) CLVMObject {
	if args.ListLen() != 2 {
		log.Fatalf("c takes exactly 2 arguments, got %d: %s", args.ListLen(), args)
	}
	return CLVMPair{args.(CLVMPair).First, args.(CLVMPair).Rest.(CLVMPair).First} //CONS_COST
}
func opFirst(args CLVMObject) CLVMObject {
	if args.ListLen() != 1 {
		log.Fatalf("f takes exactly 1 argument: %s", args)
	}
	return args.(CLVMPair).First.(CLVMPair).First //FIRST_COST
}
func opRest(args CLVMObject) CLVMObject {
	if args.ListLen() != 1 {
		log.Fatalf("r takes exactly 1 argument: %s", args)
	}
	return args.(CLVMPair).First.(CLVMPair).Rest //REST_COST
}

// def op_rest(args):
//     if args.list_len() != 1:
//         raise EvalError("r takes exactly 1 argument", args)
//     return REST_COST, args.first().rest()

func opListp(args CLVMObject) CLVMObject {
	if args.ListLen() != 1 {
		log.Fatalf("l takes exactly 1 argument: %s", args)
	}
	//LISTP_COST
	if args.(CLVMPair).First.Listp() {
		return ATOM_TRUE
	} else {
		return ATOM_FALSE
	}
}

// def op_raise(args):
//     raise EvalError("clvm raise", args)

func opEq(args CLVMObject) CLVMObject {
	if args.ListLen() != 2 {
		log.Fatalf("= takes exactly 2 arguments: %s", args)
	}
	a0, ok0 := args.(CLVMPair).First.(CLVMAtom)
	a1, ok1 := args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
	if !ok0 {
		log.Fatalf("= on list: %s", a0)
	}
	if !ok1 {
		log.Fatalf("= on list: %s", a1)
	}
	// cost = EQ_BASE_COST
	// cost += (len(b0) + len(b1)) * EQ_COST_PER_BYTE
	if a0.Equal(a1) {
		return ATOM_TRUE
	}
	return ATOM_FALSE
}

func opAdd(args CLVMObject) CLVMObject {
	total := big.NewInt(0)
	cost := int64(ARITH_BASE_COST)
	argSize := int64(0)
	arg := args
	for !arg.Nullp() {
		if pair, ok := arg.(CLVMPair); ok {
			if atom, ok := pair.First.(CLVMAtom); ok {
				total.Add(total, atom.AsInt())
				argSize += int64(len(atom.Bytes))
				cost += ARITH_COST_PER_ARG
			} else {
				log.Fatalf("add: arg.left not an atom: %s", pair.First)
			}
			arg = pair.Rest
		} else {
			log.Fatalf("add: arg not a pair: %s", arg)
		}
	}
	cost += argSize * ARITH_COST_PER_BYTE
	_, res := mallocCost(cost, CLVMAtomFromInt(total))
	return res
}
func opSha256(args CLVMObject) CLVMObject {
	cost := int64(SHA256_BASE_COST)
	argLen := int64(0)
	h := sha256.New()
	arg := args
	for !arg.Nullp() {
		if pair, ok := arg.(CLVMPair); ok {
			if atom, ok := pair.First.(CLVMAtom); ok {
				argLen += int64(len(atom.Bytes))
				cost += SHA256_COST_PER_ARG
				h.Write(atom.Bytes)
			} else {
				log.Fatalf("sha256: arg.left not an atom: %s", pair.First)
			}
			arg = pair.Rest
		} else {
			log.Fatalf("sha256: arg not a pair: %s", arg)
		}
	}
	cost += argLen * SHA256_COST_PER_BYTE
	_, res := mallocCost(cost, CLVMAtom{h.Sum(nil)})
	return res
}
func opGrBytes(args CLVMObject) CLVMObject {
	// argList := list(args.as_iter())
	argCount := args.ListLen()
	if argCount != 2 {
		log.Fatalf(">s takes exactly 2 arguments: %s", args)
	}
	a0, ok0 := args.(CLVMPair).First.(CLVMAtom)
	a1, ok1 := args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
	if !ok0 {
		log.Fatalf(">s on list: %s", a0)
	}
	if !ok1 {
		log.Fatalf(">s on list: %s", a1)
	}
	// cost = GRS_BASE_COST
	// cost += (len(b0) + len(b1)) * GRS_COST_PER_BYTE
	if bytes.Compare(a0.Bytes, a1.Bytes) > 0 {
		return ATOM_TRUE
	}
	return ATOM_FALSE
}
func opSubstr(args CLVMObject) CLVMObject {
	argCount := args.ListLen()
	if argCount != 2 && argCount != 3 {
		log.Fatalf("substr takes exactly 2 or 3 arguments: %s", args)
	}
	s0, ok := args.(CLVMPair).First.(CLVMAtom)
	if !ok {
		log.Fatalf("substr on list: %s", args.(CLVMPair).First)
	}

	var i1, i2 int32
	if argCount == 2 {
		i1 = args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom).AsInt32()
		i2 = int32(len(s0.Bytes))
	} else {
		i1 = args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom).AsInt32()
		i2 = args.(CLVMPair).Rest.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom).AsInt32()
	}

	if i2 > int32(len(s0.Bytes)) || i2 < i1 || i2 < 0 || i1 < 0 {
		log.Fatalf("invalid indices for substr: %s", args)
	}
	s := s0.Bytes[i1:i2]
	// cost := 1
	return CLVMAtom{s}
}
func opConcat(args CLVMObject) CLVMObject {
	cost := int64(CONCAT_BASE_COST)
	s := []byte{}
	arg := args
	for !arg.Nullp() {
		if pair, ok := arg.(CLVMPair); ok {
			if atom, ok := pair.First.(CLVMAtom); ok {
				s = append(s, atom.Bytes...)
				cost += CONCAT_COST_PER_ARG
			} else {
				log.Fatalf("concat: arg.left not an atom: %s", pair.First)
			}
			arg = pair.Rest
		} else {
			log.Fatalf("concat: arg not a pair: %s", arg)
		}
	}
	cost += int64(len(s)) * CONCAT_COST_PER_BYTE
	_, res := mallocCost(cost, CLVMAtom{s})
	return res
}

// https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/wallet/puzzles/rom_bootstrap_generator.clvm
// https://raw.githubusercontent.com/Chia-Network/chia-blockchain/latest/chia/wallet/puzzles/rom_bootstrap_generator.clvm.hex
//go:embed rom_bootstrap_generator.clvm.hex
var ROM_BOOTSTRAP_GENERATOR_HEX string
var ROM_BOOTSTRAP_GENERATOR = MustParseProgramFromHex(ROM_BOOTSTRAP_GENERATOR_HEX)

func msbMask(b byte) byte {
	b |= (b >> 1)
	b |= (b >> 2)
	b |= (b >> 4)
	if b == 255 {
		return 128
	}
	return (b + 1) >> 1
}

func popValue(valueStack *[]CLVMObject) CLVMObject {
	res := (*valueStack)[len(*valueStack)-1]
	*valueStack = (*valueStack)[:len(*valueStack)-1]
	return res
}

func traversePath(sexp CLVMAtom, env CLVMObject) CLVMObject {
	if sexp.Nullp() {
		return ATOM_NULL
	}

	b := sexp.Bytes

	endByteCursor := 0
	for endByteCursor < len(b) && b[endByteCursor] == 0 {
		endByteCursor += 1
	}

	if endByteCursor == len(b) {
		return ATOM_NULL
	}

	// create a bitmask for the most significant *set* bit in the last non-zero byte
	endBitmask := msbMask(b[endByteCursor])

	byteCursor := len(b) - 1
	bitmask := 0x01
	for byteCursor > endByteCursor || bitmask < int(endBitmask) {
		if envPair, ok := env.(CLVMPair); ok {
			if b[byteCursor]&byte(bitmask) > 0 {
				env = envPair.Rest
			} else {
				env = envPair.First
			}
			bitmask <<= 1
			if bitmask == 0x100 {
				byteCursor -= 1
				bitmask = 0x01
			}
		} else {
			log.Fatalf("path into atom: %s", env)
		}
	}
	return env
}

func swapOp(opStack *[]interface{}, valueStack *[]CLVMObject) {
	v2 := popValue(valueStack)
	v1 := popValue(valueStack)
	*valueStack = append(*valueStack, v2, v1)
}

func consOp(opStack *[]interface{}, valueStack *[]CLVMObject) {
	v1 := popValue(valueStack)
	v2 := popValue(valueStack)
	*valueStack = append(*valueStack, CLVMPair{v1, v2})
}

func evalOp(opStack *[]interface{}, valueStack *[]CLVMObject) {
	// pre_eval_op?

	pair := popValue(valueStack).(CLVMPair)
	sexp := pair.First
	args := pair.Rest

	// put a bunch of ops on op_stack

	switch sexp := sexp.(type) {
	case CLVMAtom:
		r := traversePath(sexp, args)
		*valueStack = append(*valueStack, r)
		return
	case CLVMPair:
		operator := sexp.First
		switch operator := operator.(type) {
		case CLVMPair:
			newOperator, mustBeNil := operator.First, operator.Rest
			if newOperator.Listp() {
				log.Fatalf("in ((X)...) syntax X must be lone atom: %#v", sexp)
			}
			if atom, ok := mustBeNil.(CLVMAtom); !ok || len(atom.Bytes) != 0 {
				log.Fatalf("in ((X)...) syntax X must be lone atom: %#v", sexp)
			}
			newOperandList := sexp.Rest
			*valueStack = append(*valueStack, newOperator)
			*valueStack = append(*valueStack, newOperandList)
			*opStack = append(*opStack, applyOp)
			return
		case CLVMAtom:
			operandList := sexp.Rest
			if operator.Equal(ATOM_QUOTE) {
				*valueStack = append(*valueStack, operandList)
				return
			}
			*opStack = append(*opStack, applyOp)
			*valueStack = append(*valueStack, operator)
			for !operandList.Nullp() {
				first := operandList.(CLVMPair).First
				*valueStack = append(*valueStack, CLVMPair{first, args}) //first.cons(args)
				*opStack = append(*opStack, consOp)
				*opStack = append(*opStack, evalOp)
				*opStack = append(*opStack, swapOp)
				operandList = operandList.(CLVMPair).Rest
			}
			*valueStack = append(*valueStack, ATOM_NULL)
			return
		default:
			log.Fatalf("unexpected operator: %T", operator)
		}
	default:
		log.Fatalf("unexpected sexp: %T", sexp)
	}
}

func applyOp(opStack *[]interface{}, valueStack *[]CLVMObject) {
	operandList := popValue(valueStack)
	operator := popValue(valueStack)

	op, ok := operator.(CLVMAtom)
	if !ok {
		log.Fatalf("internal error: %#v", operator)
	}

	if op.Equal(ATOM_APPLY) {
		operandListPair, ok := operandList.(CLVMPair)
		if !ok || operandListPair.ListLen() != 2 {
			log.Fatalf("apply requires exactly 2 parameters, got %d: %#v",
				operandListPair.ListLen(), operandList)
		}
		newProgram := operandListPair.First
		newArgs := operandListPair.Rest.(CLVMPair).First
		*valueStack = append(*valueStack, CLVMPair{newProgram, newArgs})
		*opStack = append(*opStack, evalOp)
		return
	}

	var r CLVMObject = ATOM_NULL
	var opFunc func(CLVMObject) CLVMObject = nil
	if len(op.Bytes) == 1 {
		opFunc = OP_FROM_BYTE[op.Bytes[0]].f //may still be nil
		if RUN_DEBUG {
			fmt.Println("op", OP_FROM_BYTE[op.Bytes[0]].name)
		}
	}
	if opFunc != nil { //TODO: more bytes/zero bytes
		r = opFunc(operandList)
	} else {
		log.Fatalf("WARN: unknown op %s with args %s", hex.EncodeToString(op.Bytes), operandList)
	}
	*valueStack = append(*valueStack, r)
}

func RunProgram(program CLVMObject, args CLVMObject) CLVMObject {
	opStack := []interface{}{evalOp}
	valueStack := []CLVMObject{CLVMPair{program, args}}

	if RUN_DEBUG {
		fmt.Println("run")
	}
	for len(opStack) > 0 {
		f := opStack[len(opStack)-1].(func(*[]interface{}, *[]CLVMObject))
		opStack = opStack[:len(opStack)-1]
		if RUN_DEBUG {
			fmt.Println("pop", len(opStack))
		}
		f(&opStack, &valueStack)
	}
	if RUN_DEBUG {
		fmt.Println("end", len(valueStack))
	}
	return valueStack[0]
}
