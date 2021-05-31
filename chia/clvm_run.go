package chia

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"math/big"
)

//go:generate go run gen/gen_clvm_ops_map.go -fname clvm_ops_map_generated.go
//go:generate go fmt clvm_ops_map_generated.go

const RUN_DEBUG = false

type EvalError struct {
	Msg    string
	Values map[string]CLVMObject
}

func (e *EvalError) Error() string {
	res := e.Msg
	if len(e.Values) > 0 {
		isFirst := true
		for k, v := range e.Values {
			delim := ", "
			if isFirst {
				delim = ": "
				isFirst = false
			}
			res += delim + k + "=" + v.StringExt(CLVM_STRING_EXT_CFG_ERRORS)
		}
	}
	return res
}
func (e *EvalError) With(name string, value CLVMObject) *EvalError {
	if e.Values == nil {
		e.Values = make(map[string]CLVMObject)
	}
	e.Values[name] = value
	return e
}
func NewEvalError(format string, a ...interface{}) *EvalError {
	return &EvalError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func mallocCost(cost int64, atom CLVMAtom) (int64, CLVMAtom) {
	return cost + int64(len(atom.Bytes))*MALLOC_COST_PER_BYTE, atom
}

func ensureArgsLen(funcName string, args CLVMObject, length int) *EvalError {
	s := ""
	if length > 1 {
		s = "s"
	}
	if args.ListLen() != length {
		return NewEvalError("%s takes exactly %d argument%s, got %d", funcName, length, s, args.ListLen()).
			With("args", args)
	}
	return nil
}

func opIf(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen("i", args, 3); err != nil {
		return 0, nil, err
	}
	r := args.(CLVMPair).Rest.(CLVMPair)
	if args.(CLVMPair).First.Nullp() {
		return IF_COST, r.Rest.(CLVMPair).First, nil
	}
	return IF_COST, r.First, nil
}
func opCons(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen("c", args, 2); err != nil {
		return 0, nil, err
	}
	return CONS_COST, CLVMPair{args.(CLVMPair).First, args.(CLVMPair).Rest.(CLVMPair).First}, nil
}
func opFirst(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen("f", args, 1); err != nil {
		return 0, nil, err
	}
	a0, ok := args.(CLVMPair).First.(CLVMPair)
	if !ok {
		return 0, nil, NewEvalError("first of non-cons").With("arg", args.(CLVMPair).First)
	}
	return FIRST_COST, a0.First, nil
}
func opRest(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen("r", args, 1); err != nil {
		return 0, nil, err
	}
	a0, ok := args.(CLVMPair).First.(CLVMPair)
	if !ok {
		return 0, nil, NewEvalError("rest of non-cons").With("arg", args.(CLVMPair).First)
	}
	return REST_COST, a0.Rest, nil
}

// def op_rest(args):
//     if args.list_len() != 1:
//         raise EvalError("r takes exactly 1 argument", args)
//     return REST_COST, args.first().rest()

func opListp(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen("l", args, 1); err != nil {
		return 0, nil, err
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
	if err := ensureArgsLen("=", args, 2); err != nil {
		return 0, nil, err
	}
	a0, ok0 := args.(CLVMPair).First.(CLVMAtom)
	a1, ok1 := args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
	if !ok0 {
		return 0, nil, NewEvalError("= on list").With("arg0", args.(CLVMPair).First)
	}
	if !ok1 {
		return 0, nil, NewEvalError("= on list").With("arg1", args.(CLVMPair).Rest.(CLVMPair).First)
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
	argIter := NewCLVMIter(args)
	for argIter.Next() {
		item := argIter.Get()
		if atom, ok := item.(CLVMAtom); ok {
			total.Add(total, atom.AsInt())
			argSize += int64(len(atom.Bytes))
			cost += ARITH_COST_PER_ARG
		} else {
			return cost, nil, NewEvalError("add on list").With("arg", item)
		}
	}
	if err := argIter.Err(); err != nil {
		return cost, nil, err
	}
	cost += argSize * ARITH_COST_PER_BYTE
	cost, res := mallocCost(cost, CLVMAtomFromInt(total))
	return cost, res, nil
}

func opMultiply(args CLVMObject) (int64, CLVMObject, error) {
	cost := int64(MUL_BASE_COST)

	argIter := NewCLVMIter(args)
	if !argIter.Next() {
		cost, res := mallocCost(cost, CLVMAtom{[]byte{1}})
		return cost, res, nil
	}

	item := argIter.Get()
	atom, ok := item.(CLVMAtom)
	if !ok {
		return cost, nil, NewEvalError("multiply on list").With("arg0", item)
	}
	v := atom.AsInt()
	vs := len(atom.Bytes)

	for argIter.Next() {
		item := argIter.Get()
		atom, ok := item.(CLVMAtom)
		if !ok {
			return cost, nil, NewEvalError("multiply on list").With("arg", item)
		}
		r := atom.AsInt()
		rs := len(atom.Bytes)
		cost += MUL_COST_PER_OP
		cost += int64(rs+vs) * MUL_LINEAR_COST_PER_BYTE
		cost += int64(rs*vs) / MUL_SQUARE_COST_PER_BYTE_DIVIDER
		v.Mul(v, r)
		vs = (v.BitLen() + 7) >> 3
	}
	if err := argIter.Err(); err != nil {
		return cost, nil, err
	}

	cost, res := mallocCost(cost, CLVMAtomFromInt(v))
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
				return cost, nil, NewEvalError("sha256 on list").With("arg", pair.First)
			}
			arg = pair.Rest
		} else {
			return cost, nil, NewEvalError("sha256: arg.rest is atom").With("arg.rest", arg)
		}
	}
	cost += argLen * SHA256_COST_PER_BYTE
	cost, res := mallocCost(cost, CLVMAtom{h.Sum(nil)})
	return cost, res, nil
}
func opGrBytes(args CLVMObject) (int64, CLVMObject, error) {
	if err := ensureArgsLen(">s", args, 2); err != nil {
		return 0, nil, err
	}
	a0, ok0 := args.(CLVMPair).First.(CLVMAtom)
	a1, ok1 := args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
	if !ok0 {
		return 0, nil, NewEvalError(">s on list").With("arg0", args.(CLVMPair).First)
	}
	if !ok1 {
		return 0, nil, NewEvalError(">s on list").With("arg1", args.(CLVMPair).Rest.(CLVMPair).First)
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
		return 0, nil, NewEvalError("substr takes 2 or 3 arguments, got %d", argCount).With("args", args)
	}
	s0, ok := args.(CLVMPair).First.(CLVMAtom)
	if !ok {
		return 0, nil, NewEvalError("substr on list").With("arg0", args.(CLVMPair).First)
	}
	a0, ok := args.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
	if !ok {
		return 0, nil, NewEvalError("substr on list").With("arg1", args.(CLVMPair).Rest.(CLVMPair).First)
	}

	i1, err := a0.AsInt32()
	if err != nil {
		return 0, nil, err
	}

	var i2 int32
	if argCount == 2 {
		i2 = int32(len(s0.Bytes))
	} else {
		a2, ok := args.(CLVMPair).Rest.(CLVMPair).Rest.(CLVMPair).First.(CLVMAtom)
		if !ok {
			return 0, nil, NewEvalError("substr on list").With("arg2", args.(CLVMPair).Rest.(CLVMPair).Rest.(CLVMPair).First)
		}
		i2, err = a2.AsInt32()
		if err != nil {
			return 0, nil, err
		}
	}

	if i2 > int32(len(s0.Bytes)) || i2 < i1 || i2 < 0 || i1 < 0 {
		return 0, nil, NewEvalError("invalid indices for substr: i1=%d, i2=%d, len=%d", i1, i2, len(s0.Bytes)).With("args", args)
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
				return cost, nil, NewEvalError("concat on list").With("arg", pair.First)
			}
			arg = pair.Rest
		} else {
			return cost, nil, NewEvalError("concat: arg.rest is atom").With("arg.rest", arg)
		}
	}
	cost += int64(len(s)) * CONCAT_COST_PER_BYTE
	cost, res := mallocCost(cost, CLVMAtom{s})
	return cost, res, nil
}

func binopReduction(opName string, initialValue *big.Int, args CLVMObject, opFunc func(a, b *big.Int) *big.Int) (int64, CLVMObject, error) {
	total := initialValue
	argSize := 0
	cost := int64(LOG_BASE_COST)

	argIter := NewCLVMIter(args)
	for argIter.Next() {
		item := argIter.Get()
		if atom, ok := item.(CLVMAtom); ok {
			total = opFunc(total, atom.AsInt())
			argSize += len(atom.Bytes)
			cost += LOG_COST_PER_ARG
		} else {
			return cost, nil, NewEvalError("%s on list", opName).With("arg", item)
		}
	}
	if err := argIter.Err(); err != nil {
		return cost, nil, err
	}

	cost += int64(argSize) * LOG_COST_PER_BYTE
	cost, res := mallocCost(cost, CLVMAtomFromInt(total))
	return cost, res, nil
}

func opLogand(args CLVMObject) (int64, CLVMObject, error) {
	binop := func(a, b *big.Int) *big.Int {
		return a.And(a, b)
	}
	return binopReduction("logand", new(big.Int).SetInt64(-1), args, binop)
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
			return cost, nil, NewEvalError("path into atom").With("env", env)
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
			return cost, err
		}
		*valueStack = append(*valueStack, r)
		return cost, nil
	case CLVMPair:
		operator := sexp.First
		switch operator := operator.(type) {
		case CLVMPair:
			newOperator, mustBeNil := operator.First, operator.Rest
			if newOperator.Listp() {
				return 0, NewEvalError("in ((X)...) syntax X must be lone atom").With("sexp", sexp)
			}
			if atom, ok := mustBeNil.(CLVMAtom); !ok || len(atom.Bytes) != 0 {
				return 0, NewEvalError("in ((X)...) syntax X must be lone atom").With("sexp", sexp)
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
			return 0, NewEvalError("unexpected operator").With("operator", operator)
		}
	default:
		return 0, NewEvalError("unexpected sexp").With("sexp", sexp)
	}
}

func runApply(opStack *[]interface{}, valueStack *[]CLVMObject) (int64, error) {
	operandList := popValue(valueStack)
	operator := popValue(valueStack)

	op, ok := operator.(CLVMAtom)
	if !ok {
		return 0, NewEvalError("internal error").With("operator", operator)
	}

	if op.Equal(ATOM_APPLY) {
		operandListPair, ok := operandList.(CLVMPair)
		if !ok || operandListPair.ListLen() != 2 {
			return 0, NewEvalError("apply requires exactly 2 parameters, got %d", operandListPair.ListLen()).With("args", operandList)
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
			return 0, err
		}
		*valueStack = append(*valueStack, r)
		return cost, nil
	} else {
		return 0, NewEvalError("unknown op 0x%s", hex.EncodeToString(op.Bytes)).With("args", operandList)
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
			return cost, nil, err
		}
		cost += fCost
	}
	return cost, valueStack[0], nil
}
