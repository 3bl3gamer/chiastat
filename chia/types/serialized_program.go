package types

import (
	"chiastat/chia/clvm"
	"chiastat/chia/utils"
)

type SerializedProgram struct {
	Root  clvm.SExp
	Bytes []byte
}

// https://github.com/Chia-Network/clvm/blob/main/clvm/serialize.py
func (prog *SerializedProgram) FromBytes(buf *utils.ParseBuf) {
	startBufPos := buf.Pos()
	sexp := clvm.SExpFromBytes(buf)
	if buf.Err() != nil {
		return
	}
	prog.Root = sexp
	prog.Bytes = buf.Copy(startBufPos, buf.Pos())
}

func (p SerializedProgram) ToBytes(buf *[]byte) {
	panic("not implemented yet")
}
