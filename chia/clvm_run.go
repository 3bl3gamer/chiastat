package chia

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/ansel1/merry"
)

//go:generate go run gen/gen_clvm_ops_map.go -fname clvm_ops_map_generated.go
//go:generate go fmt clvm_ops_map_generated.go

const RUN_DEBUG = false

func mallocCost(cost int64, atom CLVMAtom) (int64, CLVMAtom) {
	return cost + int64(len(atom.Bytes))*MALLOC_COST_PER_BYTE, atom
}

func opIf(args CLVMObject) (int64, CLVMObject, error) {
	if args.ListLen() != 3 {
		log.Fatalf("i takes exactly 3 arguments: %s", args)
	}
	r := args.(CLVMPair).Rest.(CLVMPair)
	if args.(CLVMPair).First.Nullp() {
		return IF_COST, r.Rest.(CLVMPair).First, nil
	}
	return IF_COST, r.First, nil
}
func opCons(args CLVMObject) (int64, CLVMObject, error) {
	if args.ListLen() != 2 {
		log.Fatalf("c takes exactly 2 arguments, got %d: %s", args.ListLen(), args)
	}
	return CONS_COST, CLVMPair{args.(CLVMPair).First, args.(CLVMPair).Rest.(CLVMPair).First}, nil
}
func opFirst(args CLVMObject) (int64, CLVMObject, error) {
	if args.ListLen() != 1 {
		log.Fatalf("f takes exactly 1 argument: %s", args)
	}
	return FIRST_COST, args.(CLVMPair).First.(CLVMPair).First, nil
}
func opRest(args CLVMObject) (int64, CLVMObject, error) {
	if args.ListLen() != 1 {
		log.Fatalf("r takes exactly 1 argument: %s", args)
	}
	return REST_COST, args.(CLVMPair).First.(CLVMPair).Rest, nil
}

// def op_rest(args):
//     if args.list_len() != 1:
//         raise EvalError("r takes exactly 1 argument", args)
//     return REST_COST, args.first().rest()

func opListp(args CLVMObject) (int64, CLVMObject, error) {
	if args.ListLen() != 1 {
		log.Fatalf("l takes exactly 1 argument: %s", args)
	}
	if args.(CLVMPair).First.Listp() {
		return LISTP_COST, ATOM_TRUE, nil
	} else {
		return LISTP_COST, ATOM_FALSE, nil
	}
}

// def op_raise(args):
//     raise EvalError("clvm raise", args)

func opEq(args CLVMObject) (int64, CLVMObject, error) {
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
	cost := int64(EQ_BASE_COST)
	cost += int64(len(a0.Bytes)+len(a1.Bytes)) * EQ_COST_PER_BYTE
	if a0.Equal(a1) {
		return cost, ATOM_TRUE, nil
	}
	return cost, ATOM_FALSE, nil
}

func opAdd(args CLVMObject) (int64, CLVMObject, error) {
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
	cost, res := mallocCost(cost, CLVMAtomFromInt(total))
	return cost, res, nil
}
func opSha256(args CLVMObject) (int64, CLVMObject, error) {
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
	cost, res := mallocCost(cost, CLVMAtom{h.Sum(nil)})
	return cost, res, nil
}
func opGrBytes(args CLVMObject) (int64, CLVMObject, error) {
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
	cost := int64(GRS_BASE_COST)
	cost += int64(len(a0.Bytes)+len(a1.Bytes)) * GRS_COST_PER_BYTE
	if bytes.Compare(a0.Bytes, a1.Bytes) > 0 {
		return cost, ATOM_TRUE, nil
	}
	return cost, ATOM_FALSE, nil
}
func opSubstr(args CLVMObject) (int64, CLVMObject, error) {
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
	cost := int64(1)
	return cost, CLVMAtom{s}, nil
}
func opConcat(args CLVMObject) (int64, CLVMObject, error) {
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
	cost, res := mallocCost(cost, CLVMAtom{s})
	return cost, res, nil
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

func traversePath(sexp CLVMAtom, env CLVMObject) (int64, CLVMObject, error) {
	cost := int64(PATH_LOOKUP_BASE_COST)
	cost += PATH_LOOKUP_COST_PER_LEG
	if sexp.Nullp() {
		return cost, ATOM_NULL, nil
	}

	b := sexp.Bytes

	endByteCursor := 0
	for endByteCursor < len(b) && b[endByteCursor] == 0 {
		endByteCursor += 1
	}

	cost += int64(endByteCursor) * PATH_LOOKUP_COST_PER_ZERO_BYTE
	if endByteCursor == len(b) {
		return cost, ATOM_NULL, nil
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
			cost += PATH_LOOKUP_COST_PER_LEG
			bitmask <<= 1
			if bitmask == 0x100 {
				byteCursor -= 1
				bitmask = 0x01
			}
		} else {
			return cost, nil, merry.Errorf("path into atom: %s", env)
		}
	}
	return cost, env, nil
}

func runSwap(opStack *[]interface{}, valueStack *[]CLVMObject) (int64, error) {
	v2 := popValue(valueStack)
	v1 := popValue(valueStack)
	*valueStack = append(*valueStack, v2, v1)
	return 0, nil
}

func runCons(opStack *[]interface{}, valueStack *[]CLVMObject) (int64, error) {
	v1 := popValue(valueStack)
	v2 := popValue(valueStack)
	*valueStack = append(*valueStack, CLVMPair{v1, v2})
	return 0, nil
}

func runEval(opStack *[]interface{}, valueStack *[]CLVMObject) (int64, error) {
	// pre_eval_op?

	pair := popValue(valueStack).(CLVMPair)
	sexp := pair.First
	args := pair.Rest

	// put a bunch of ops on op_stack

	switch sexp := sexp.(type) {
	case CLVMAtom:
		cost, r, err := traversePath(sexp, args)
		if err != nil {
			return cost, merry.Wrap(err)
		}
		*valueStack = append(*valueStack, r)
		return cost, nil
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
			*opStack = append(*opStack, runApply)
			return APPLY_COST, nil
		case CLVMAtom:
			operandList := sexp.Rest
			if operator.Equal(ATOM_QUOTE) {
				*valueStack = append(*valueStack, operandList)
				return QUOTE_COST, nil
			}
			*opStack = append(*opStack, runApply)
			*valueStack = append(*valueStack, operator)
			for !operandList.Nullp() {
				first := operandList.(CLVMPair).First
				*valueStack = append(*valueStack, CLVMPair{first, args}) //first.cons(args)
				*opStack = append(*opStack, runCons)
				*opStack = append(*opStack, runEval)
				*opStack = append(*opStack, runSwap)
				operandList = operandList.(CLVMPair).Rest
			}
			*valueStack = append(*valueStack, ATOM_NULL)
			return 1, nil
		default:
			log.Fatalf("unexpected operator: %T", operator)
			return 0, nil
		}
	default:
		log.Fatalf("unexpected sexp: %T", sexp)
		return 0, nil
	}
}

func runApply(opStack *[]interface{}, valueStack *[]CLVMObject) (int64, error) {
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
		*opStack = append(*opStack, runEval)
		return APPLY_COST, nil
	}

	var opFunc func(CLVMObject) (int64, CLVMObject, error) = nil
	if len(op.Bytes) == 1 {
		opFunc = OP_FROM_BYTE[op.Bytes[0]].f //may still be nil
		if RUN_DEBUG {
			fmt.Println("op", OP_FROM_BYTE[op.Bytes[0]].name)
		}
	}
	if opFunc != nil { //TODO: more bytes/zero bytes
		cost, r, err := opFunc(operandList)
		if err != nil {
			return 0, merry.Wrap(err)
		}
		*valueStack = append(*valueStack, r)
		return cost, nil
	} else {
		return 0, merry.Errorf("unknown op %s with args %s", hex.EncodeToString(op.Bytes), operandList)
	}
}

func RunProgram(program CLVMObject, args CLVMObject) (int64, CLVMObject, error) {
	opStack := []interface{}{runEval}
	valueStack := []CLVMObject{CLVMPair{program, args}}
	cost := int64(0)

	for len(opStack) > 0 {
		f := opStack[len(opStack)-1].(func(*[]interface{}, *[]CLVMObject) (int64, error))
		opStack = opStack[:len(opStack)-1]
		if RUN_DEBUG {
			fmt.Println("pop", len(opStack))
		}
		fCost, err := f(&opStack, &valueStack)
		if err != nil {
			return cost, nil, merry.Wrap(err)
		}
		cost += fCost
	}
	return cost, valueStack[0], nil
}
