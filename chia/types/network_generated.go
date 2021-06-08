// Generated, do not edit.
package types

import (
	"chiastat/chia/utils"
	"math/big"
)

type Message struct {
	// one of ProtocolMessageTypes
	Type uint8
	// (optional)
	ID   uint16
	Data []byte
}

func (obj *Message) FromBytes(buf *utils.ParseBuf) {
	obj.Type = buf.Uint8()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.ID = buf.Uint16()
	}
	obj.Data = buf.Bytes()
}

func (obj Message) ToBytes(buf *[]byte) {
	utils.Uint8ToBytes(buf, obj.Type)
	obj_ID_isSet := !(obj.ID == 0)
	utils.BoolToBytes(buf, obj_ID_isSet)
	if obj_ID_isSet {
		utils.Uint16ToBytes(buf, obj.ID)
	}
	utils.BytesToBytes(buf, obj.Data)
}

type Handshake struct {
	NetworkID       string
	ProtocolVersion string
	SoftwareVersion string
	ServerPort      uint16
	NodeType        uint8
	Capabilities    []TupleUint16Str
}

func (obj *Handshake) FromBytes(buf *utils.ParseBuf) {
	obj.NetworkID = buf.String()
	obj.ProtocolVersion = buf.String()
	obj.SoftwareVersion = buf.String()
	obj.ServerPort = buf.Uint16()
	obj.NodeType = buf.Uint8()
	len_obj_Capabilities := buf.Uint32()
	obj.Capabilities = make([]TupleUint16Str, len_obj_Capabilities)
	for i := uint32(0); i < len_obj_Capabilities; i++ {
		obj.Capabilities[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
}

func (obj Handshake) ToBytes(buf *[]byte) {
	utils.StringToBytes(buf, obj.NetworkID)
	utils.StringToBytes(buf, obj.ProtocolVersion)
	utils.StringToBytes(buf, obj.SoftwareVersion)
	utils.Uint16ToBytes(buf, obj.ServerPort)
	utils.Uint8ToBytes(buf, obj.NodeType)
	utils.Uint32ToBytes(buf, uint32(len(obj.Capabilities)))
	for _, item := range obj.Capabilities {
		item.ToBytes(buf)
	}
}

type NewPeak struct {
	HeaderHash                [32]byte
	Height                    uint32
	Weight                    *big.Int
	ForkPointWithPreviousPeak uint32
	UnfinishedRewardBlockHash [32]byte
}

func (obj *NewPeak) FromBytes(buf *utils.ParseBuf) {
	obj.HeaderHash = buf.Bytes32()
	obj.Height = buf.Uint32()
	obj.Weight = buf.Uint128()
	obj.ForkPointWithPreviousPeak = buf.Uint32()
	obj.UnfinishedRewardBlockHash = buf.Bytes32()
}

func (obj NewPeak) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.HeaderHash)
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Uint128ToBytes(buf, obj.Weight)
	utils.Uint32ToBytes(buf, obj.ForkPointWithPreviousPeak)
	utils.Bytes32ToBytes(buf, obj.UnfinishedRewardBlockHash)
}

type NewTransaction struct {
	TransactionID [32]byte
	Cost          uint64
	Fees          uint64
}

func (obj *NewTransaction) FromBytes(buf *utils.ParseBuf) {
	obj.TransactionID = buf.Bytes32()
	obj.Cost = buf.Uint64()
	obj.Fees = buf.Uint64()
}

func (obj NewTransaction) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.TransactionID)
	utils.Uint64ToBytes(buf, obj.Cost)
	utils.Uint64ToBytes(buf, obj.Fees)
}

type RequestTransaction struct {
	TransactionID [32]byte
}

func (obj *RequestTransaction) FromBytes(buf *utils.ParseBuf) {
	obj.TransactionID = buf.Bytes32()
}

func (obj RequestTransaction) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.TransactionID)
}

type RespondTransaction struct {
	Transaction SpendBundle
}

func (obj *RespondTransaction) FromBytes(buf *utils.ParseBuf) {
	obj.Transaction.FromBytes(buf)
}

func (obj RespondTransaction) ToBytes(buf *[]byte) {
	obj.Transaction.ToBytes(buf)
}

type RequestProofOfWeight struct {
	TotalNumberOfBlocks uint32
	Tip                 [32]byte
}

func (obj *RequestProofOfWeight) FromBytes(buf *utils.ParseBuf) {
	obj.TotalNumberOfBlocks = buf.Uint32()
	obj.Tip = buf.Bytes32()
}

func (obj RequestProofOfWeight) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.TotalNumberOfBlocks)
	utils.Bytes32ToBytes(buf, obj.Tip)
}

type RespondProofOfWeight struct {
	Wp  WeightProof
	Tip [32]byte
}

func (obj *RespondProofOfWeight) FromBytes(buf *utils.ParseBuf) {
	obj.Wp.FromBytes(buf)
	obj.Tip = buf.Bytes32()
}

func (obj RespondProofOfWeight) ToBytes(buf *[]byte) {
	obj.Wp.ToBytes(buf)
	utils.Bytes32ToBytes(buf, obj.Tip)
}

type RequestBlock struct {
	Height                  uint32
	IncludeTransactionBlock bool
}

func (obj *RequestBlock) FromBytes(buf *utils.ParseBuf) {
	obj.Height = buf.Uint32()
	obj.IncludeTransactionBlock = buf.Bool()
}

func (obj RequestBlock) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.Height)
	utils.BoolToBytes(buf, obj.IncludeTransactionBlock)
}

type RejectBlock struct {
	Height uint32
}

func (obj *RejectBlock) FromBytes(buf *utils.ParseBuf) {
	obj.Height = buf.Uint32()
}

func (obj RejectBlock) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.Height)
}

type RequestBlocks struct {
	StartHeight             uint32
	EndHeight               uint32
	IncludeTransactionBlock bool
}

func (obj *RequestBlocks) FromBytes(buf *utils.ParseBuf) {
	obj.StartHeight = buf.Uint32()
	obj.EndHeight = buf.Uint32()
	obj.IncludeTransactionBlock = buf.Bool()
}

func (obj RequestBlocks) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.StartHeight)
	utils.Uint32ToBytes(buf, obj.EndHeight)
	utils.BoolToBytes(buf, obj.IncludeTransactionBlock)
}

type RespondBlocks struct {
	StartHeight uint32
	EndHeight   uint32
	Blocks      []FullBlock
}

func (obj *RespondBlocks) FromBytes(buf *utils.ParseBuf) {
	obj.StartHeight = buf.Uint32()
	obj.EndHeight = buf.Uint32()
	len_obj_Blocks := buf.Uint32()
	obj.Blocks = make([]FullBlock, len_obj_Blocks)
	for i := uint32(0); i < len_obj_Blocks; i++ {
		obj.Blocks[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
}

func (obj RespondBlocks) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.StartHeight)
	utils.Uint32ToBytes(buf, obj.EndHeight)
	utils.Uint32ToBytes(buf, uint32(len(obj.Blocks)))
	for _, item := range obj.Blocks {
		item.ToBytes(buf)
	}
}

type RejectBlocks struct {
	StartHeight uint32
	EndHeight   uint32
}

func (obj *RejectBlocks) FromBytes(buf *utils.ParseBuf) {
	obj.StartHeight = buf.Uint32()
	obj.EndHeight = buf.Uint32()
}

func (obj RejectBlocks) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.StartHeight)
	utils.Uint32ToBytes(buf, obj.EndHeight)
}

type RespondBlock struct {
	Block FullBlock
}

func (obj *RespondBlock) FromBytes(buf *utils.ParseBuf) {
	obj.Block.FromBytes(buf)
}

func (obj RespondBlock) ToBytes(buf *[]byte) {
	obj.Block.ToBytes(buf)
}

type NewUnfinishedBlock struct {
	UnfinishedRewardHash [32]byte
}

func (obj *NewUnfinishedBlock) FromBytes(buf *utils.ParseBuf) {
	obj.UnfinishedRewardHash = buf.Bytes32()
}

func (obj NewUnfinishedBlock) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.UnfinishedRewardHash)
}

type RequestUnfinishedBlock struct {
	UnfinishedRewardHash [32]byte
}

func (obj *RequestUnfinishedBlock) FromBytes(buf *utils.ParseBuf) {
	obj.UnfinishedRewardHash = buf.Bytes32()
}

func (obj RequestUnfinishedBlock) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.UnfinishedRewardHash)
}

type RespondUnfinishedBlock struct {
	UnfinishedBlock UnfinishedBlock
}

func (obj *RespondUnfinishedBlock) FromBytes(buf *utils.ParseBuf) {
	obj.UnfinishedBlock.FromBytes(buf)
}

func (obj RespondUnfinishedBlock) ToBytes(buf *[]byte) {
	obj.UnfinishedBlock.ToBytes(buf)
}

type NewSignagePointOrEndOfSubSlot struct {
	// (optional)
	PrevChallengeHash  *[32]byte
	ChallengeHash      [32]byte
	IndexFromChallenge uint8
	LastRcInfusion     [32]byte
}

func (obj *NewSignagePointOrEndOfSubSlot) FromBytes(buf *utils.ParseBuf) {
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.PrevChallengeHash = &t
	}
	obj.ChallengeHash = buf.Bytes32()
	obj.IndexFromChallenge = buf.Uint8()
	obj.LastRcInfusion = buf.Bytes32()
}

func (obj NewSignagePointOrEndOfSubSlot) ToBytes(buf *[]byte) {
	obj_PrevChallengeHash_isSet := !(obj.PrevChallengeHash == nil)
	utils.BoolToBytes(buf, obj_PrevChallengeHash_isSet)
	if obj_PrevChallengeHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.PrevChallengeHash)
	}
	utils.Bytes32ToBytes(buf, obj.ChallengeHash)
	utils.Uint8ToBytes(buf, obj.IndexFromChallenge)
	utils.Bytes32ToBytes(buf, obj.LastRcInfusion)
}

type RequestSignagePointOrEndOfSubSlot struct {
	ChallengeHash      [32]byte
	IndexFromChallenge uint8
	LastRcInfusion     [32]byte
}

func (obj *RequestSignagePointOrEndOfSubSlot) FromBytes(buf *utils.ParseBuf) {
	obj.ChallengeHash = buf.Bytes32()
	obj.IndexFromChallenge = buf.Uint8()
	obj.LastRcInfusion = buf.Bytes32()
}

func (obj RequestSignagePointOrEndOfSubSlot) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.ChallengeHash)
	utils.Uint8ToBytes(buf, obj.IndexFromChallenge)
	utils.Bytes32ToBytes(buf, obj.LastRcInfusion)
}

type RespondSignagePoint struct {
	IndexFromChallenge  uint8
	ChallengeChainVdf   VDFInfo
	ChallengeChainProof VDFProof
	RewardChainVdf      VDFInfo
	RewardChainProof    VDFProof
}

func (obj *RespondSignagePoint) FromBytes(buf *utils.ParseBuf) {
	obj.IndexFromChallenge = buf.Uint8()
	obj.ChallengeChainVdf.FromBytes(buf)
	obj.ChallengeChainProof.FromBytes(buf)
	obj.RewardChainVdf.FromBytes(buf)
	obj.RewardChainProof.FromBytes(buf)
}

func (obj RespondSignagePoint) ToBytes(buf *[]byte) {
	utils.Uint8ToBytes(buf, obj.IndexFromChallenge)
	obj.ChallengeChainVdf.ToBytes(buf)
	obj.ChallengeChainProof.ToBytes(buf)
	obj.RewardChainVdf.ToBytes(buf)
	obj.RewardChainProof.ToBytes(buf)
}

type RespondEndOfSubSlot struct {
	EndOfSlotBundle EndOfSubSlotBundle
}

func (obj *RespondEndOfSubSlot) FromBytes(buf *utils.ParseBuf) {
	obj.EndOfSlotBundle.FromBytes(buf)
}

func (obj RespondEndOfSubSlot) ToBytes(buf *[]byte) {
	obj.EndOfSlotBundle.ToBytes(buf)
}

type RequestMempoolTransactions struct {
	Filter []byte
}

func (obj *RequestMempoolTransactions) FromBytes(buf *utils.ParseBuf) {
	obj.Filter = buf.Bytes()
}

func (obj RequestMempoolTransactions) ToBytes(buf *[]byte) {
	utils.BytesToBytes(buf, obj.Filter)
}

type NewCompactVDF struct {
	Height     uint32
	HeaderHash [32]byte
	FieldVdf   uint8
	VdfInfo    VDFInfo
}

func (obj *NewCompactVDF) FromBytes(buf *utils.ParseBuf) {
	obj.Height = buf.Uint32()
	obj.HeaderHash = buf.Bytes32()
	obj.FieldVdf = buf.Uint8()
	obj.VdfInfo.FromBytes(buf)
}

func (obj NewCompactVDF) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Bytes32ToBytes(buf, obj.HeaderHash)
	utils.Uint8ToBytes(buf, obj.FieldVdf)
	obj.VdfInfo.ToBytes(buf)
}

type RequestCompactVDF struct {
	Height     uint32
	HeaderHash [32]byte
	FieldVdf   uint8
	VdfInfo    VDFInfo
}

func (obj *RequestCompactVDF) FromBytes(buf *utils.ParseBuf) {
	obj.Height = buf.Uint32()
	obj.HeaderHash = buf.Bytes32()
	obj.FieldVdf = buf.Uint8()
	obj.VdfInfo.FromBytes(buf)
}

func (obj RequestCompactVDF) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Bytes32ToBytes(buf, obj.HeaderHash)
	utils.Uint8ToBytes(buf, obj.FieldVdf)
	obj.VdfInfo.ToBytes(buf)
}

type RespondCompactVDF struct {
	Height     uint32
	HeaderHash [32]byte
	FieldVdf   uint8
	VdfInfo    VDFInfo
	VdfProof   VDFProof
}

func (obj *RespondCompactVDF) FromBytes(buf *utils.ParseBuf) {
	obj.Height = buf.Uint32()
	obj.HeaderHash = buf.Bytes32()
	obj.FieldVdf = buf.Uint8()
	obj.VdfInfo.FromBytes(buf)
	obj.VdfProof.FromBytes(buf)
}

func (obj RespondCompactVDF) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Bytes32ToBytes(buf, obj.HeaderHash)
	utils.Uint8ToBytes(buf, obj.FieldVdf)
	obj.VdfInfo.ToBytes(buf)
	obj.VdfProof.ToBytes(buf)
}

// Return full list of peers
type RequestPeers struct {
}

func (obj *RequestPeers) FromBytes(buf *utils.ParseBuf) {
}

func (obj RequestPeers) ToBytes(buf *[]byte) {
}

type RespondPeers struct {
	PeerList []TimestampedPeerInfo
}

func (obj *RespondPeers) FromBytes(buf *utils.ParseBuf) {
	len_obj_PeerList := buf.Uint32()
	obj.PeerList = make([]TimestampedPeerInfo, len_obj_PeerList)
	for i := uint32(0); i < len_obj_PeerList; i++ {
		obj.PeerList[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
}

func (obj RespondPeers) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.PeerList)))
	for _, item := range obj.PeerList {
		item.ToBytes(buf)
	}
}
