package chia

import (
	"chiastat/chia/clvm"
	"chiastat/chia/utils"
)

type SerializedProgram struct {
	Root  clvm.SExp
	Bytes []byte
}

// https://github.com/Chia-Network/clvm/blob/main/clvm/serialize.py
func SerializedProgramFromBytes(buf *utils.ParseBuf) (obj SerializedProgram) {
	startBufPos := buf.Pos()
	sexp := clvm.SExpFromBytes(buf)
	if buf.Err() != nil {
		return
	}
	obj.Root = sexp
	obj.Bytes = buf.Copy(startBufPos, buf.Pos())
	return
}
