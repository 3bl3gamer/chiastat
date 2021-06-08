// Generated, do not edit.
package types

import "chiastat/chia/utils"

// === Tuples ===

type TupleUint16Str struct {
	V0 uint16
	V1 string
}

func (obj *TupleUint16Str) FromBytes(buf *utils.ParseBuf) {
	obj.V0 = buf.Uint16()
	obj.V1 = buf.String()
}

func (obj TupleUint16Str) ToBytes(buf *[]byte) {
	utils.Uint16ToBytes(buf, obj.V0)
	utils.StringToBytes(buf, obj.V1)
}

// === Dummy ===

type G1Element struct{ Bytes []byte }

func (obj *G1Element) FromBytes(buf *utils.ParseBuf) {
	obj.Bytes = buf.BytesN(48)
}
func (obj G1Element) ToBytes(buf *[]byte) {
	utils.BytesWOSizeToBytes(buf, obj.Bytes)
}

type G2Element struct{ Bytes []byte }

func (obj *G2Element) FromBytes(buf *utils.ParseBuf) {
	obj.Bytes = buf.BytesN(96)
}
func (obj G2Element) ToBytes(buf *[]byte) {
	utils.BytesWOSizeToBytes(buf, obj.Bytes)
}
