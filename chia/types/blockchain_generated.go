// Generated, do not edit.
package types

import (
	"chiastat/chia/utils"
	"math/big"
)

// This class is not included or hashed into the blockchain, but it is kept in memory as a more
// efficient way to maintain data about the blockchain. This allows us to validate future blocks,
// difficulty adjustments, etc, without saving the whole header block in memory.
type BlockRecord struct {
	HeaderHash [32]byte
	// Header hash of the previous block
	PrevHash [32]byte
	Height   uint32
	// Total cumulative difficulty of all ancestor blocks since genesis
	Weight *big.Int
	// Total number of VDF iterations since genesis, including this block
	TotalIters        *big.Int
	SignagePointIndex uint8
	// This is the intermediary VDF output at ip_iters in challenge chain
	ChallengeVdfOutput ClassgroupElement
	// (optional) This is the intermediary VDF output at ip_iters in infused cc, iff deficit <= 3
	InfusedChallengeVdfOutput *ClassgroupElement
	// The reward chain infusion output, input to next VDF
	RewardInfusionNewChallenge [32]byte
	// Hash of challenge chain data, used to validate end of slots in the future
	ChallengeBlockInfoHash [32]byte
	// Current network sub_slot_iters parameter
	SubSlotIters uint64
	// Need to keep track of these because Coins are created in a future block
	PoolPuzzleHash   [32]byte
	FarmerPuzzleHash [32]byte
	// The number of iters required for this proof of space
	RequiredIters uint64
	// A deficit of 16 is an overflow block after an infusion. Deficit of 15 is a challenge block
	Deficit                    uint8
	Overflow                   bool
	PrevTransactionBlockHeight uint32
	// (optional)
	Timestamp uint64
	// (optional) Header hash of the previous transaction block
	PrevTransactionBlockHash *[32]byte
	// (optional)
	Fees uint64
	// (optional)
	RewardClaimsIncorporated []Coin
	// (optional)
	FinishedChallengeSlotHashes [][32]byte
	// (optional)
	FinishedInfusedChallengeSlotHashes [][32]byte
	// (optional)
	FinishedRewardSlotHashes [][32]byte
	// (optional)
	SubEpochSummaryIncluded *SubEpochSummary
}

func (obj *BlockRecord) FromBytes(buf *utils.ParseBuf) {
	obj.HeaderHash = buf.Bytes32()
	obj.PrevHash = buf.Bytes32()
	obj.Height = buf.Uint32()
	obj.Weight = buf.Uint128()
	obj.TotalIters = buf.Uint128()
	obj.SignagePointIndex = buf.Uint8()
	obj.ChallengeVdfOutput.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t ClassgroupElement
		t.FromBytes(buf)
		obj.InfusedChallengeVdfOutput = &t
	}
	obj.RewardInfusionNewChallenge = buf.Bytes32()
	obj.ChallengeBlockInfoHash = buf.Bytes32()
	obj.SubSlotIters = buf.Uint64()
	obj.PoolPuzzleHash = buf.Bytes32()
	obj.FarmerPuzzleHash = buf.Bytes32()
	obj.RequiredIters = buf.Uint64()
	obj.Deficit = buf.Uint8()
	obj.Overflow = buf.Bool()
	obj.PrevTransactionBlockHeight = buf.Uint32()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.Timestamp = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.PrevTransactionBlockHash = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.Fees = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		len_obj_RewardClaimsIncorporated := buf.Uint32()
		obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
		for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
			obj.RewardClaimsIncorporated[i].FromBytes(buf)
			if buf.Err() != nil {
				return
			}
		}
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		len_obj_FinishedChallengeSlotHashes := buf.Uint32()
		obj.FinishedChallengeSlotHashes = make([][32]byte, len_obj_FinishedChallengeSlotHashes)
		for i := uint32(0); i < len_obj_FinishedChallengeSlotHashes; i++ {
			obj.FinishedChallengeSlotHashes[i] = buf.Bytes32()
			if buf.Err() != nil {
				return
			}
		}
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		len_obj_FinishedInfusedChallengeSlotHashes := buf.Uint32()
		obj.FinishedInfusedChallengeSlotHashes = make([][32]byte, len_obj_FinishedInfusedChallengeSlotHashes)
		for i := uint32(0); i < len_obj_FinishedInfusedChallengeSlotHashes; i++ {
			obj.FinishedInfusedChallengeSlotHashes[i] = buf.Bytes32()
			if buf.Err() != nil {
				return
			}
		}
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		len_obj_FinishedRewardSlotHashes := buf.Uint32()
		obj.FinishedRewardSlotHashes = make([][32]byte, len_obj_FinishedRewardSlotHashes)
		for i := uint32(0); i < len_obj_FinishedRewardSlotHashes; i++ {
			obj.FinishedRewardSlotHashes[i] = buf.Bytes32()
			if buf.Err() != nil {
				return
			}
		}
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t SubEpochSummary
		t.FromBytes(buf)
		obj.SubEpochSummaryIncluded = &t
	}
}

func (obj BlockRecord) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.HeaderHash)
	utils.Bytes32ToBytes(buf, obj.PrevHash)
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Uint128ToBytes(buf, obj.Weight)
	utils.Uint128ToBytes(buf, obj.TotalIters)
	utils.Uint8ToBytes(buf, obj.SignagePointIndex)
	obj.ChallengeVdfOutput.ToBytes(buf)
	obj_InfusedChallengeVdfOutput_isSet := !(obj.InfusedChallengeVdfOutput == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeVdfOutput_isSet)
	if obj_InfusedChallengeVdfOutput_isSet {
		obj.InfusedChallengeVdfOutput.ToBytes(buf)
	}
	utils.Bytes32ToBytes(buf, obj.RewardInfusionNewChallenge)
	utils.Bytes32ToBytes(buf, obj.ChallengeBlockInfoHash)
	utils.Uint64ToBytes(buf, obj.SubSlotIters)
	utils.Bytes32ToBytes(buf, obj.PoolPuzzleHash)
	utils.Bytes32ToBytes(buf, obj.FarmerPuzzleHash)
	utils.Uint64ToBytes(buf, obj.RequiredIters)
	utils.Uint8ToBytes(buf, obj.Deficit)
	utils.BoolToBytes(buf, obj.Overflow)
	utils.Uint32ToBytes(buf, obj.PrevTransactionBlockHeight)
	obj_Timestamp_isSet := !(obj.Timestamp == 0)
	utils.BoolToBytes(buf, obj_Timestamp_isSet)
	if obj_Timestamp_isSet {
		utils.Uint64ToBytes(buf, obj.Timestamp)
	}
	obj_PrevTransactionBlockHash_isSet := !(obj.PrevTransactionBlockHash == nil)
	utils.BoolToBytes(buf, obj_PrevTransactionBlockHash_isSet)
	if obj_PrevTransactionBlockHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.PrevTransactionBlockHash)
	}
	obj_Fees_isSet := !(obj.Fees == 0)
	utils.BoolToBytes(buf, obj_Fees_isSet)
	if obj_Fees_isSet {
		utils.Uint64ToBytes(buf, obj.Fees)
	}
	obj_RewardClaimsIncorporated_isSet := !(len(obj.RewardClaimsIncorporated) == 0)
	utils.BoolToBytes(buf, obj_RewardClaimsIncorporated_isSet)
	if obj_RewardClaimsIncorporated_isSet {
		utils.Uint32ToBytes(buf, uint32(len(obj.RewardClaimsIncorporated)))
		for _, item := range obj.RewardClaimsIncorporated {
			item.ToBytes(buf)
		}
	}
	obj_FinishedChallengeSlotHashes_isSet := !(len(obj.FinishedChallengeSlotHashes) == 0)
	utils.BoolToBytes(buf, obj_FinishedChallengeSlotHashes_isSet)
	if obj_FinishedChallengeSlotHashes_isSet {
		utils.Uint32ToBytes(buf, uint32(len(obj.FinishedChallengeSlotHashes)))
		for _, item := range obj.FinishedChallengeSlotHashes {
			utils.Bytes32ToBytes(buf, item)
		}
	}
	obj_FinishedInfusedChallengeSlotHashes_isSet := !(len(obj.FinishedInfusedChallengeSlotHashes) == 0)
	utils.BoolToBytes(buf, obj_FinishedInfusedChallengeSlotHashes_isSet)
	if obj_FinishedInfusedChallengeSlotHashes_isSet {
		utils.Uint32ToBytes(buf, uint32(len(obj.FinishedInfusedChallengeSlotHashes)))
		for _, item := range obj.FinishedInfusedChallengeSlotHashes {
			utils.Bytes32ToBytes(buf, item)
		}
	}
	obj_FinishedRewardSlotHashes_isSet := !(len(obj.FinishedRewardSlotHashes) == 0)
	utils.BoolToBytes(buf, obj_FinishedRewardSlotHashes_isSet)
	if obj_FinishedRewardSlotHashes_isSet {
		utils.Uint32ToBytes(buf, uint32(len(obj.FinishedRewardSlotHashes)))
		for _, item := range obj.FinishedRewardSlotHashes {
			utils.Bytes32ToBytes(buf, item)
		}
	}
	obj_SubEpochSummaryIncluded_isSet := !(obj.SubEpochSummaryIncluded == nil)
	utils.BoolToBytes(buf, obj_SubEpochSummaryIncluded_isSet)
	if obj_SubEpochSummaryIncluded_isSet {
		obj.SubEpochSummaryIncluded.ToBytes(buf)
	}
}

// This structure is used in the body for the reward and fees genesis coins.
type Coin struct {
	ParentCoinInfo [32]byte
	PuzzleHash     [32]byte
	Amount         uint64
}

func (obj *Coin) FromBytes(buf *utils.ParseBuf) {
	obj.ParentCoinInfo = buf.Bytes32()
	obj.PuzzleHash = buf.Bytes32()
	obj.Amount = buf.Uint64()
}

func (obj Coin) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.ParentCoinInfo)
	utils.Bytes32ToBytes(buf, obj.PuzzleHash)
	utils.Uint64ToBytes(buf, obj.Amount)
}

// Represents a classgroup element (a,b,c) where a, b, and c are 512 bit signed integers. However this is using
// a compressed representation. VDF outputs are a single classgroup element. VDF proofs can also be one classgroup
// element (or multiple).
type ClassgroupElement struct {
	Data [100]byte
}

func (obj *ClassgroupElement) FromBytes(buf *utils.ParseBuf) {
	obj.Data = buf.Bytes100()
}

func (obj ClassgroupElement) ToBytes(buf *[]byte) {
	utils.Bytes100ToBytes(buf, obj.Data)
}

type SubEpochSummary struct {
	PrevSubepochSummaryHash [32]byte
	// hash of reward chain at end of last segment
	RewardChainHash [32]byte
	// How many more blocks than 384*(N-1)
	NumBlocksOverflow uint8
	// (optional) Only once per epoch (diff adjustment)
	NewDifficulty uint64
	// (optional) Only once per epoch (diff adjustment)
	NewSubSlotIters uint64
}

func (obj *SubEpochSummary) FromBytes(buf *utils.ParseBuf) {
	obj.PrevSubepochSummaryHash = buf.Bytes32()
	obj.RewardChainHash = buf.Bytes32()
	obj.NumBlocksOverflow = buf.Uint8()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewDifficulty = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewSubSlotIters = buf.Uint64()
	}
}

func (obj SubEpochSummary) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.PrevSubepochSummaryHash)
	utils.Bytes32ToBytes(buf, obj.RewardChainHash)
	utils.Uint8ToBytes(buf, obj.NumBlocksOverflow)
	obj_NewDifficulty_isSet := !(obj.NewDifficulty == 0)
	utils.BoolToBytes(buf, obj_NewDifficulty_isSet)
	if obj_NewDifficulty_isSet {
		utils.Uint64ToBytes(buf, obj.NewDifficulty)
	}
	obj_NewSubSlotIters_isSet := !(obj.NewSubSlotIters == 0)
	utils.BoolToBytes(buf, obj_NewSubSlotIters_isSet)
	if obj_NewSubSlotIters_isSet {
		utils.Uint64ToBytes(buf, obj.NewSubSlotIters)
	}
}

type VDFProof struct {
	WitnessType          uint8
	Witness              []byte
	NormalizedToIdentity bool
}

func (obj *VDFProof) FromBytes(buf *utils.ParseBuf) {
	obj.WitnessType = buf.Uint8()
	obj.Witness = buf.Bytes()
	obj.NormalizedToIdentity = buf.Bool()
}

func (obj VDFProof) ToBytes(buf *[]byte) {
	utils.Uint8ToBytes(buf, obj.WitnessType)
	utils.BytesToBytes(buf, obj.Witness)
	utils.BoolToBytes(buf, obj.NormalizedToIdentity)
}

type VDFInfo struct {
	// Used to generate the discriminant (VDF group)
	Challenge          [32]byte
	NumberOfIterations uint64
	Output             ClassgroupElement
}

func (obj *VDFInfo) FromBytes(buf *utils.ParseBuf) {
	obj.Challenge = buf.Bytes32()
	obj.NumberOfIterations = buf.Uint64()
	obj.Output.FromBytes(buf)
}

func (obj VDFInfo) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.Challenge)
	utils.Uint64ToBytes(buf, obj.NumberOfIterations)
	obj.Output.ToBytes(buf)
}

type Foliage struct {
	PrevBlockHash             [32]byte
	RewardBlockHash           [32]byte
	FoliageBlockData          FoliageBlockData
	FoliageBlockDataSignature G2Element
	// (optional)
	FoliageTransactionBlockHash *[32]byte
	// (optional)
	FoliageTransactionBlockSignature *G2Element
}

func (obj *Foliage) FromBytes(buf *utils.ParseBuf) {
	obj.PrevBlockHash = buf.Bytes32()
	obj.RewardBlockHash = buf.Bytes32()
	obj.FoliageBlockData.FromBytes(buf)
	obj.FoliageBlockDataSignature.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.FoliageTransactionBlockHash = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t G2Element
		t.FromBytes(buf)
		obj.FoliageTransactionBlockSignature = &t
	}
}

func (obj Foliage) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.PrevBlockHash)
	utils.Bytes32ToBytes(buf, obj.RewardBlockHash)
	obj.FoliageBlockData.ToBytes(buf)
	obj.FoliageBlockDataSignature.ToBytes(buf)
	obj_FoliageTransactionBlockHash_isSet := !(obj.FoliageTransactionBlockHash == nil)
	utils.BoolToBytes(buf, obj_FoliageTransactionBlockHash_isSet)
	if obj_FoliageTransactionBlockHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.FoliageTransactionBlockHash)
	}
	obj_FoliageTransactionBlockSignature_isSet := !(obj.FoliageTransactionBlockSignature == nil)
	utils.BoolToBytes(buf, obj_FoliageTransactionBlockSignature_isSet)
	if obj_FoliageTransactionBlockSignature_isSet {
		obj.FoliageTransactionBlockSignature.ToBytes(buf)
	}
}

type FoliageTransactionBlock struct {
	PrevTransactionBlockHash [32]byte
	Timestamp                uint64
	FilterHash               [32]byte
	AdditionsRoot            [32]byte
	RemovalsRoot             [32]byte
	TransactionsInfoHash     [32]byte
}

func (obj *FoliageTransactionBlock) FromBytes(buf *utils.ParseBuf) {
	obj.PrevTransactionBlockHash = buf.Bytes32()
	obj.Timestamp = buf.Uint64()
	obj.FilterHash = buf.Bytes32()
	obj.AdditionsRoot = buf.Bytes32()
	obj.RemovalsRoot = buf.Bytes32()
	obj.TransactionsInfoHash = buf.Bytes32()
}

func (obj FoliageTransactionBlock) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.PrevTransactionBlockHash)
	utils.Uint64ToBytes(buf, obj.Timestamp)
	utils.Bytes32ToBytes(buf, obj.FilterHash)
	utils.Bytes32ToBytes(buf, obj.AdditionsRoot)
	utils.Bytes32ToBytes(buf, obj.RemovalsRoot)
	utils.Bytes32ToBytes(buf, obj.TransactionsInfoHash)
}

type FoliageBlockData struct {
	UnfinishedRewardBlockHash [32]byte
	PoolTarget                PoolTarget
	// (optional) Iff ProofOfSpace has a pool pk
	PoolSignature          *G2Element
	FarmerRewardPuzzleHash [32]byte
	// Used for future updates. Can be any 32 byte value initially
	ExtensionData [32]byte
}

func (obj *FoliageBlockData) FromBytes(buf *utils.ParseBuf) {
	obj.UnfinishedRewardBlockHash = buf.Bytes32()
	obj.PoolTarget.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t G2Element
		t.FromBytes(buf)
		obj.PoolSignature = &t
	}
	obj.FarmerRewardPuzzleHash = buf.Bytes32()
	obj.ExtensionData = buf.Bytes32()
}

func (obj FoliageBlockData) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.UnfinishedRewardBlockHash)
	obj.PoolTarget.ToBytes(buf)
	obj_PoolSignature_isSet := !(obj.PoolSignature == nil)
	utils.BoolToBytes(buf, obj_PoolSignature_isSet)
	if obj_PoolSignature_isSet {
		obj.PoolSignature.ToBytes(buf)
	}
	utils.Bytes32ToBytes(buf, obj.FarmerRewardPuzzleHash)
	utils.Bytes32ToBytes(buf, obj.ExtensionData)
}

type TransactionsInfo struct {
	// sha256 of the block generator in this block
	GeneratorRoot [32]byte
	// sha256 of the concatenation of the generator ref list entries
	GeneratorRefsRoot   [32]byte
	AggregatedSignature G2Element
	// This only includes user fees, not block rewards
	Fees uint64
	// This is the total cost of running this block in the CLVM
	Cost uint64
	// These can be in any order
	RewardClaimsIncorporated []Coin
}

func (obj *TransactionsInfo) FromBytes(buf *utils.ParseBuf) {
	obj.GeneratorRoot = buf.Bytes32()
	obj.GeneratorRefsRoot = buf.Bytes32()
	obj.AggregatedSignature.FromBytes(buf)
	obj.Fees = buf.Uint64()
	obj.Cost = buf.Uint64()
	len_obj_RewardClaimsIncorporated := buf.Uint32()
	obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
	for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
		obj.RewardClaimsIncorporated[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
}

func (obj TransactionsInfo) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.GeneratorRoot)
	utils.Bytes32ToBytes(buf, obj.GeneratorRefsRoot)
	obj.AggregatedSignature.ToBytes(buf)
	utils.Uint64ToBytes(buf, obj.Fees)
	utils.Uint64ToBytes(buf, obj.Cost)
	utils.Uint32ToBytes(buf, uint32(len(obj.RewardClaimsIncorporated)))
	for _, item := range obj.RewardClaimsIncorporated {
		item.ToBytes(buf)
	}
}

type RewardChainBlock struct {
	Weight               *big.Int
	Height               uint32
	TotalIters           *big.Int
	SignagePointIndex    uint8
	PosSsCcChallengeHash [32]byte
	ProofOfSpace         ProofOfSpace
	// (optional) Not present for first sp in slot
	ChallengeChainSpVdf       *VDFInfo
	ChallengeChainSpSignature G2Element
	ChallengeChainIpVdf       VDFInfo
	// (optional) Not present for first sp in slot
	RewardChainSpVdf       *VDFInfo
	RewardChainSpSignature G2Element
	RewardChainIpVdf       VDFInfo
	// (optional) Iff deficit < 16
	InfusedChallengeChainIpVdf *VDFInfo
	IsTransactionBlock         bool
}

func (obj *RewardChainBlock) FromBytes(buf *utils.ParseBuf) {
	obj.Weight = buf.Uint128()
	obj.Height = buf.Uint32()
	obj.TotalIters = buf.Uint128()
	obj.SignagePointIndex = buf.Uint8()
	obj.PosSsCcChallengeHash = buf.Bytes32()
	obj.ProofOfSpace.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.ChallengeChainSpVdf = &t
	}
	obj.ChallengeChainSpSignature.FromBytes(buf)
	obj.ChallengeChainIpVdf.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.RewardChainSpVdf = &t
	}
	obj.RewardChainSpSignature.FromBytes(buf)
	obj.RewardChainIpVdf.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.InfusedChallengeChainIpVdf = &t
	}
	obj.IsTransactionBlock = buf.Bool()
}

func (obj RewardChainBlock) ToBytes(buf *[]byte) {
	utils.Uint128ToBytes(buf, obj.Weight)
	utils.Uint32ToBytes(buf, obj.Height)
	utils.Uint128ToBytes(buf, obj.TotalIters)
	utils.Uint8ToBytes(buf, obj.SignagePointIndex)
	utils.Bytes32ToBytes(buf, obj.PosSsCcChallengeHash)
	obj.ProofOfSpace.ToBytes(buf)
	obj_ChallengeChainSpVdf_isSet := !(obj.ChallengeChainSpVdf == nil)
	utils.BoolToBytes(buf, obj_ChallengeChainSpVdf_isSet)
	if obj_ChallengeChainSpVdf_isSet {
		obj.ChallengeChainSpVdf.ToBytes(buf)
	}
	obj.ChallengeChainSpSignature.ToBytes(buf)
	obj.ChallengeChainIpVdf.ToBytes(buf)
	obj_RewardChainSpVdf_isSet := !(obj.RewardChainSpVdf == nil)
	utils.BoolToBytes(buf, obj_RewardChainSpVdf_isSet)
	if obj_RewardChainSpVdf_isSet {
		obj.RewardChainSpVdf.ToBytes(buf)
	}
	obj.RewardChainSpSignature.ToBytes(buf)
	obj.RewardChainIpVdf.ToBytes(buf)
	obj_InfusedChallengeChainIpVdf_isSet := !(obj.InfusedChallengeChainIpVdf == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainIpVdf_isSet)
	if obj_InfusedChallengeChainIpVdf_isSet {
		obj.InfusedChallengeChainIpVdf.ToBytes(buf)
	}
	utils.BoolToBytes(buf, obj.IsTransactionBlock)
}

type RewardChainBlockUnfinished struct {
	TotalIters           *big.Int
	SignagePointIndex    uint8
	PosSsCcChallengeHash [32]byte
	ProofOfSpace         ProofOfSpace
	// (optional) Not present for first sp in slot
	ChallengeChainSpVdf       *VDFInfo
	ChallengeChainSpSignature G2Element
	// (optional) Not present for first sp in slot
	RewardChainSpVdf       *VDFInfo
	RewardChainSpSignature G2Element
}

func (obj *RewardChainBlockUnfinished) FromBytes(buf *utils.ParseBuf) {
	obj.TotalIters = buf.Uint128()
	obj.SignagePointIndex = buf.Uint8()
	obj.PosSsCcChallengeHash = buf.Bytes32()
	obj.ProofOfSpace.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.ChallengeChainSpVdf = &t
	}
	obj.ChallengeChainSpSignature.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.RewardChainSpVdf = &t
	}
	obj.RewardChainSpSignature.FromBytes(buf)
}

func (obj RewardChainBlockUnfinished) ToBytes(buf *[]byte) {
	utils.Uint128ToBytes(buf, obj.TotalIters)
	utils.Uint8ToBytes(buf, obj.SignagePointIndex)
	utils.Bytes32ToBytes(buf, obj.PosSsCcChallengeHash)
	obj.ProofOfSpace.ToBytes(buf)
	obj_ChallengeChainSpVdf_isSet := !(obj.ChallengeChainSpVdf == nil)
	utils.BoolToBytes(buf, obj_ChallengeChainSpVdf_isSet)
	if obj_ChallengeChainSpVdf_isSet {
		obj.ChallengeChainSpVdf.ToBytes(buf)
	}
	obj.ChallengeChainSpSignature.ToBytes(buf)
	obj_RewardChainSpVdf_isSet := !(obj.RewardChainSpVdf == nil)
	utils.BoolToBytes(buf, obj_RewardChainSpVdf_isSet)
	if obj_RewardChainSpVdf_isSet {
		obj.RewardChainSpVdf.ToBytes(buf)
	}
	obj.RewardChainSpSignature.ToBytes(buf)
}

type ChallengeChainSubSlot struct {
	ChallengeChainEndOfSlotVdf VDFInfo
	// (optional) Only at the end of a slot
	InfusedChallengeChainSubSlotHash *[32]byte
	// (optional) Only once per sub-epoch, and one sub-epoch delayed
	SubepochSummaryHash *[32]byte
	// (optional) Only at the end of epoch, sub-epoch, and slot
	NewSubSlotIters uint64
	// (optional) Only at the end of epoch, sub-epoch, and slot
	NewDifficulty uint64
}

func (obj *ChallengeChainSubSlot) FromBytes(buf *utils.ParseBuf) {
	obj.ChallengeChainEndOfSlotVdf.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.InfusedChallengeChainSubSlotHash = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.SubepochSummaryHash = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewSubSlotIters = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewDifficulty = buf.Uint64()
	}
}

func (obj ChallengeChainSubSlot) ToBytes(buf *[]byte) {
	obj.ChallengeChainEndOfSlotVdf.ToBytes(buf)
	obj_InfusedChallengeChainSubSlotHash_isSet := !(obj.InfusedChallengeChainSubSlotHash == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainSubSlotHash_isSet)
	if obj_InfusedChallengeChainSubSlotHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.InfusedChallengeChainSubSlotHash)
	}
	obj_SubepochSummaryHash_isSet := !(obj.SubepochSummaryHash == nil)
	utils.BoolToBytes(buf, obj_SubepochSummaryHash_isSet)
	if obj_SubepochSummaryHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.SubepochSummaryHash)
	}
	obj_NewSubSlotIters_isSet := !(obj.NewSubSlotIters == 0)
	utils.BoolToBytes(buf, obj_NewSubSlotIters_isSet)
	if obj_NewSubSlotIters_isSet {
		utils.Uint64ToBytes(buf, obj.NewSubSlotIters)
	}
	obj_NewDifficulty_isSet := !(obj.NewDifficulty == 0)
	utils.BoolToBytes(buf, obj_NewDifficulty_isSet)
	if obj_NewDifficulty_isSet {
		utils.Uint64ToBytes(buf, obj.NewDifficulty)
	}
}

type InfusedChallengeChainSubSlot struct {
	InfusedChallengeChainEndOfSlotVdf VDFInfo
}

func (obj *InfusedChallengeChainSubSlot) FromBytes(buf *utils.ParseBuf) {
	obj.InfusedChallengeChainEndOfSlotVdf.FromBytes(buf)
}

func (obj InfusedChallengeChainSubSlot) ToBytes(buf *[]byte) {
	obj.InfusedChallengeChainEndOfSlotVdf.ToBytes(buf)
}

type RewardChainSubSlot struct {
	EndOfSlotVdf              VDFInfo
	ChallengeChainSubSlotHash [32]byte
	// (optional)
	InfusedChallengeChainSubSlotHash *[32]byte
	// 16 or less. usually zero
	Deficit uint8
}

func (obj *RewardChainSubSlot) FromBytes(buf *utils.ParseBuf) {
	obj.EndOfSlotVdf.FromBytes(buf)
	obj.ChallengeChainSubSlotHash = buf.Bytes32()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.InfusedChallengeChainSubSlotHash = &t
	}
	obj.Deficit = buf.Uint8()
}

func (obj RewardChainSubSlot) ToBytes(buf *[]byte) {
	obj.EndOfSlotVdf.ToBytes(buf)
	utils.Bytes32ToBytes(buf, obj.ChallengeChainSubSlotHash)
	obj_InfusedChallengeChainSubSlotHash_isSet := !(obj.InfusedChallengeChainSubSlotHash == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainSubSlotHash_isSet)
	if obj_InfusedChallengeChainSubSlotHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.InfusedChallengeChainSubSlotHash)
	}
	utils.Uint8ToBytes(buf, obj.Deficit)
}

type SubSlotProofs struct {
	ChallengeChainSlotProof VDFProof
	// (optional)
	InfusedChallengeChainSlotProof *VDFProof
	RewardChainSlotProof           VDFProof
}

func (obj *SubSlotProofs) FromBytes(buf *utils.ParseBuf) {
	obj.ChallengeChainSlotProof.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.InfusedChallengeChainSlotProof = &t
	}
	obj.RewardChainSlotProof.FromBytes(buf)
}

func (obj SubSlotProofs) ToBytes(buf *[]byte) {
	obj.ChallengeChainSlotProof.ToBytes(buf)
	obj_InfusedChallengeChainSlotProof_isSet := !(obj.InfusedChallengeChainSlotProof == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainSlotProof_isSet)
	if obj_InfusedChallengeChainSlotProof_isSet {
		obj.InfusedChallengeChainSlotProof.ToBytes(buf)
	}
	obj.RewardChainSlotProof.ToBytes(buf)
}

type PoolTarget struct {
	PuzzleHash [32]byte
	// A max height of 0 means it is valid forever
	MaxHeight uint32
}

func (obj *PoolTarget) FromBytes(buf *utils.ParseBuf) {
	obj.PuzzleHash = buf.Bytes32()
	obj.MaxHeight = buf.Uint32()
}

func (obj PoolTarget) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.PuzzleHash)
	utils.Uint32ToBytes(buf, obj.MaxHeight)
}

type ProofOfSpace struct {
	Challenge [32]byte
	// (optional) Only one of these two should be present
	PoolPublicKey *G1Element
	// (optional)
	PoolContractPuzzleHash *[32]byte
	PlotPublicKey          G1Element
	Size                   uint8
	Proof                  []byte
}

func (obj *ProofOfSpace) FromBytes(buf *utils.ParseBuf) {
	obj.Challenge = buf.Bytes32()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t G1Element
		t.FromBytes(buf)
		obj.PoolPublicKey = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t [32]byte
		t = buf.Bytes32()
		obj.PoolContractPuzzleHash = &t
	}
	obj.PlotPublicKey.FromBytes(buf)
	obj.Size = buf.Uint8()
	obj.Proof = buf.Bytes()
}

func (obj ProofOfSpace) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.Challenge)
	obj_PoolPublicKey_isSet := !(obj.PoolPublicKey == nil)
	utils.BoolToBytes(buf, obj_PoolPublicKey_isSet)
	if obj_PoolPublicKey_isSet {
		obj.PoolPublicKey.ToBytes(buf)
	}
	obj_PoolContractPuzzleHash_isSet := !(obj.PoolContractPuzzleHash == nil)
	utils.BoolToBytes(buf, obj_PoolContractPuzzleHash_isSet)
	if obj_PoolContractPuzzleHash_isSet {
		utils.Bytes32ToBytes(buf, *obj.PoolContractPuzzleHash)
	}
	obj.PlotPublicKey.ToBytes(buf)
	utils.Uint8ToBytes(buf, obj.Size)
	utils.BytesToBytes(buf, obj.Proof)
}

type FullBlock struct {
	// If first sb
	FinishedSubSlots []EndOfSubSlotBundle
	// Reward chain trunk data
	RewardChainBlock RewardChainBlock
	// (optional) If not first sp in sub-slot
	ChallengeChainSpProof *VDFProof
	ChallengeChainIpProof VDFProof
	// (optional) If not first sp in sub-slot
	RewardChainSpProof *VDFProof
	RewardChainIpProof VDFProof
	// (optional) Iff deficit < 4
	InfusedChallengeChainIpProof *VDFProof
	// Reward chain foliage data
	Foliage Foliage
	// (optional) Reward chain foliage data (tx block)
	FoliageTransactionBlock *FoliageTransactionBlock
	// (optional) Reward chain foliage data (tx block additional)
	TransactionsInfo *TransactionsInfo
	// (optional) Program that generates transactions
	TransactionsGenerator *SerializedProgram
	// List of block heights of previous generators referenced in this block
	TransactionsGeneratorRefList []uint32
}

func (obj *FullBlock) FromBytes(buf *utils.ParseBuf) {
	len_obj_FinishedSubSlots := buf.Uint32()
	obj.FinishedSubSlots = make([]EndOfSubSlotBundle, len_obj_FinishedSubSlots)
	for i := uint32(0); i < len_obj_FinishedSubSlots; i++ {
		obj.FinishedSubSlots[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	obj.RewardChainBlock.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.ChallengeChainSpProof = &t
	}
	obj.ChallengeChainIpProof.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.RewardChainSpProof = &t
	}
	obj.RewardChainIpProof.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.InfusedChallengeChainIpProof = &t
	}
	obj.Foliage.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t FoliageTransactionBlock
		t.FromBytes(buf)
		obj.FoliageTransactionBlock = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t TransactionsInfo
		t.FromBytes(buf)
		obj.TransactionsInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t SerializedProgram
		t.FromBytes(buf)
		obj.TransactionsGenerator = &t
	}
	len_obj_TransactionsGeneratorRefList := buf.Uint32()
	obj.TransactionsGeneratorRefList = make([]uint32, len_obj_TransactionsGeneratorRefList)
	for i := uint32(0); i < len_obj_TransactionsGeneratorRefList; i++ {
		obj.TransactionsGeneratorRefList[i] = buf.Uint32()
		if buf.Err() != nil {
			return
		}
	}
}

func (obj FullBlock) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.FinishedSubSlots)))
	for _, item := range obj.FinishedSubSlots {
		item.ToBytes(buf)
	}
	obj.RewardChainBlock.ToBytes(buf)
	obj_ChallengeChainSpProof_isSet := !(obj.ChallengeChainSpProof == nil)
	utils.BoolToBytes(buf, obj_ChallengeChainSpProof_isSet)
	if obj_ChallengeChainSpProof_isSet {
		obj.ChallengeChainSpProof.ToBytes(buf)
	}
	obj.ChallengeChainIpProof.ToBytes(buf)
	obj_RewardChainSpProof_isSet := !(obj.RewardChainSpProof == nil)
	utils.BoolToBytes(buf, obj_RewardChainSpProof_isSet)
	if obj_RewardChainSpProof_isSet {
		obj.RewardChainSpProof.ToBytes(buf)
	}
	obj.RewardChainIpProof.ToBytes(buf)
	obj_InfusedChallengeChainIpProof_isSet := !(obj.InfusedChallengeChainIpProof == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainIpProof_isSet)
	if obj_InfusedChallengeChainIpProof_isSet {
		obj.InfusedChallengeChainIpProof.ToBytes(buf)
	}
	obj.Foliage.ToBytes(buf)
	obj_FoliageTransactionBlock_isSet := !(obj.FoliageTransactionBlock == nil)
	utils.BoolToBytes(buf, obj_FoliageTransactionBlock_isSet)
	if obj_FoliageTransactionBlock_isSet {
		obj.FoliageTransactionBlock.ToBytes(buf)
	}
	obj_TransactionsInfo_isSet := !(obj.TransactionsInfo == nil)
	utils.BoolToBytes(buf, obj_TransactionsInfo_isSet)
	if obj_TransactionsInfo_isSet {
		obj.TransactionsInfo.ToBytes(buf)
	}
	obj_TransactionsGenerator_isSet := !(obj.TransactionsGenerator == nil)
	utils.BoolToBytes(buf, obj_TransactionsGenerator_isSet)
	if obj_TransactionsGenerator_isSet {
		obj.TransactionsGenerator.ToBytes(buf)
	}
	utils.Uint32ToBytes(buf, uint32(len(obj.TransactionsGeneratorRefList)))
	for _, item := range obj.TransactionsGeneratorRefList {
		utils.Uint32ToBytes(buf, item)
	}
}

type EndOfSubSlotBundle struct {
	ChallengeChain ChallengeChainSubSlot
	// (optional)
	InfusedChallengeChain *InfusedChallengeChainSubSlot
	RewardChain           RewardChainSubSlot
	Proofs                SubSlotProofs
}

func (obj *EndOfSubSlotBundle) FromBytes(buf *utils.ParseBuf) {
	obj.ChallengeChain.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t InfusedChallengeChainSubSlot
		t.FromBytes(buf)
		obj.InfusedChallengeChain = &t
	}
	obj.RewardChain.FromBytes(buf)
	obj.Proofs.FromBytes(buf)
}

func (obj EndOfSubSlotBundle) ToBytes(buf *[]byte) {
	obj.ChallengeChain.ToBytes(buf)
	obj_InfusedChallengeChain_isSet := !(obj.InfusedChallengeChain == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChain_isSet)
	if obj_InfusedChallengeChain_isSet {
		obj.InfusedChallengeChain.ToBytes(buf)
	}
	obj.RewardChain.ToBytes(buf)
	obj.Proofs.ToBytes(buf)
}

type HeaderBlock struct {
	// If first sb
	FinishedSubSlots []EndOfSubSlotBundle
	// Reward chain trunk data
	RewardChainBlock RewardChainBlock
	// (optional) If not first sp in sub-slot
	ChallengeChainSpProof *VDFProof
	ChallengeChainIpProof VDFProof
	// (optional) If not first sp in sub-slot
	RewardChainSpProof *VDFProof
	RewardChainIpProof VDFProof
	// (optional) Iff deficit < 4
	InfusedChallengeChainIpProof *VDFProof
	// Reward chain foliage data
	Foliage Foliage
	// (optional) Reward chain foliage data (tx block)
	FoliageTransactionBlock *FoliageTransactionBlock
	// Filter for block transactions
	TransactionsFilter []byte
	// (optional) Reward chain foliage data (tx block additional)
	TransactionsInfo *TransactionsInfo
}

func (obj *HeaderBlock) FromBytes(buf *utils.ParseBuf) {
	len_obj_FinishedSubSlots := buf.Uint32()
	obj.FinishedSubSlots = make([]EndOfSubSlotBundle, len_obj_FinishedSubSlots)
	for i := uint32(0); i < len_obj_FinishedSubSlots; i++ {
		obj.FinishedSubSlots[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	obj.RewardChainBlock.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.ChallengeChainSpProof = &t
	}
	obj.ChallengeChainIpProof.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.RewardChainSpProof = &t
	}
	obj.RewardChainIpProof.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.InfusedChallengeChainIpProof = &t
	}
	obj.Foliage.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t FoliageTransactionBlock
		t.FromBytes(buf)
		obj.FoliageTransactionBlock = &t
	}
	obj.TransactionsFilter = buf.Bytes()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t TransactionsInfo
		t.FromBytes(buf)
		obj.TransactionsInfo = &t
	}
}

func (obj HeaderBlock) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.FinishedSubSlots)))
	for _, item := range obj.FinishedSubSlots {
		item.ToBytes(buf)
	}
	obj.RewardChainBlock.ToBytes(buf)
	obj_ChallengeChainSpProof_isSet := !(obj.ChallengeChainSpProof == nil)
	utils.BoolToBytes(buf, obj_ChallengeChainSpProof_isSet)
	if obj_ChallengeChainSpProof_isSet {
		obj.ChallengeChainSpProof.ToBytes(buf)
	}
	obj.ChallengeChainIpProof.ToBytes(buf)
	obj_RewardChainSpProof_isSet := !(obj.RewardChainSpProof == nil)
	utils.BoolToBytes(buf, obj_RewardChainSpProof_isSet)
	if obj_RewardChainSpProof_isSet {
		obj.RewardChainSpProof.ToBytes(buf)
	}
	obj.RewardChainIpProof.ToBytes(buf)
	obj_InfusedChallengeChainIpProof_isSet := !(obj.InfusedChallengeChainIpProof == nil)
	utils.BoolToBytes(buf, obj_InfusedChallengeChainIpProof_isSet)
	if obj_InfusedChallengeChainIpProof_isSet {
		obj.InfusedChallengeChainIpProof.ToBytes(buf)
	}
	obj.Foliage.ToBytes(buf)
	obj_FoliageTransactionBlock_isSet := !(obj.FoliageTransactionBlock == nil)
	utils.BoolToBytes(buf, obj_FoliageTransactionBlock_isSet)
	if obj_FoliageTransactionBlock_isSet {
		obj.FoliageTransactionBlock.ToBytes(buf)
	}
	utils.BytesToBytes(buf, obj.TransactionsFilter)
	obj_TransactionsInfo_isSet := !(obj.TransactionsInfo == nil)
	utils.BoolToBytes(buf, obj_TransactionsInfo_isSet)
	if obj_TransactionsInfo_isSet {
		obj.TransactionsInfo.ToBytes(buf)
	}
}

type WeightProof struct {
	SubEpochs []SubEpochData
	// sampled sub epoch
	SubEpochSegments []SubEpochChallengeSegment
	RecentChainData  []HeaderBlock
}

func (obj *WeightProof) FromBytes(buf *utils.ParseBuf) {
	len_obj_SubEpochs := buf.Uint32()
	obj.SubEpochs = make([]SubEpochData, len_obj_SubEpochs)
	for i := uint32(0); i < len_obj_SubEpochs; i++ {
		obj.SubEpochs[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	len_obj_SubEpochSegments := buf.Uint32()
	obj.SubEpochSegments = make([]SubEpochChallengeSegment, len_obj_SubEpochSegments)
	for i := uint32(0); i < len_obj_SubEpochSegments; i++ {
		obj.SubEpochSegments[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	len_obj_RecentChainData := buf.Uint32()
	obj.RecentChainData = make([]HeaderBlock, len_obj_RecentChainData)
	for i := uint32(0); i < len_obj_RecentChainData; i++ {
		obj.RecentChainData[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
}

func (obj WeightProof) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.SubEpochs)))
	for _, item := range obj.SubEpochs {
		item.ToBytes(buf)
	}
	utils.Uint32ToBytes(buf, uint32(len(obj.SubEpochSegments)))
	for _, item := range obj.SubEpochSegments {
		item.ToBytes(buf)
	}
	utils.Uint32ToBytes(buf, uint32(len(obj.RecentChainData)))
	for _, item := range obj.RecentChainData {
		item.ToBytes(buf)
	}
}

type SubEpochData struct {
	RewardChainHash   [32]byte
	NumBlocksOverflow uint8
	// (optional)
	NewSubSlotIters uint64
	// (optional)
	NewDifficulty uint64
}

func (obj *SubEpochData) FromBytes(buf *utils.ParseBuf) {
	obj.RewardChainHash = buf.Bytes32()
	obj.NumBlocksOverflow = buf.Uint8()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewSubSlotIters = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewDifficulty = buf.Uint64()
	}
}

func (obj SubEpochData) ToBytes(buf *[]byte) {
	utils.Bytes32ToBytes(buf, obj.RewardChainHash)
	utils.Uint8ToBytes(buf, obj.NumBlocksOverflow)
	obj_NewSubSlotIters_isSet := !(obj.NewSubSlotIters == 0)
	utils.BoolToBytes(buf, obj_NewSubSlotIters_isSet)
	if obj_NewSubSlotIters_isSet {
		utils.Uint64ToBytes(buf, obj.NewSubSlotIters)
	}
	obj_NewDifficulty_isSet := !(obj.NewDifficulty == 0)
	utils.BoolToBytes(buf, obj_NewDifficulty_isSet)
	if obj_NewDifficulty_isSet {
		utils.Uint64ToBytes(buf, obj.NewDifficulty)
	}
}

type SubEpochChallengeSegment struct {
	SubEpochN uint32
	SubSlots  []SubSlotData
	// (optional) in first segment of each sub_epoch
	RcSlotEndInfo *VDFInfo
}

func (obj *SubEpochChallengeSegment) FromBytes(buf *utils.ParseBuf) {
	obj.SubEpochN = buf.Uint32()
	len_obj_SubSlots := buf.Uint32()
	obj.SubSlots = make([]SubSlotData, len_obj_SubSlots)
	for i := uint32(0); i < len_obj_SubSlots; i++ {
		obj.SubSlots[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.RcSlotEndInfo = &t
	}
}

func (obj SubEpochChallengeSegment) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, obj.SubEpochN)
	utils.Uint32ToBytes(buf, uint32(len(obj.SubSlots)))
	for _, item := range obj.SubSlots {
		item.ToBytes(buf)
	}
	obj_RcSlotEndInfo_isSet := !(obj.RcSlotEndInfo == nil)
	utils.BoolToBytes(buf, obj_RcSlotEndInfo_isSet)
	if obj_RcSlotEndInfo_isSet {
		obj.RcSlotEndInfo.ToBytes(buf)
	}
}

type SubSlotData struct {
	// (optional)
	ProofOfSpace *ProofOfSpace
	// (optional)
	CcSignagePoint *VDFProof
	// (optional)
	CcInfusionPoint *VDFProof
	// (optional)
	IccInfusionPoint *VDFProof
	// (optional)
	CcSpVdfInfo *VDFInfo
	// (optional)
	SignagePointIndex uint8
	// (optional)
	CcSlotEnd *VDFProof
	// (optional)
	IccSlotEnd *VDFProof
	// (optional)
	CcSlotEndInfo *VDFInfo
	// (optional)
	IccSlotEndInfo *VDFInfo
	// (optional)
	CcIpVdfInfo *VDFInfo
	// (optional)
	IccIpVdfInfo *VDFInfo
	// (optional)
	TotalIters *big.Int
}

func (obj *SubSlotData) FromBytes(buf *utils.ParseBuf) {
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t ProofOfSpace
		t.FromBytes(buf)
		obj.ProofOfSpace = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.CcSignagePoint = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.CcInfusionPoint = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.IccInfusionPoint = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.CcSpVdfInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.SignagePointIndex = buf.Uint8()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.CcSlotEnd = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.IccSlotEnd = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.CcSlotEndInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.IccSlotEndInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.CcIpVdfInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFInfo
		t.FromBytes(buf)
		obj.IccIpVdfInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.TotalIters = buf.Uint128()
	}
}

func (obj SubSlotData) ToBytes(buf *[]byte) {
	obj_ProofOfSpace_isSet := !(obj.ProofOfSpace == nil)
	utils.BoolToBytes(buf, obj_ProofOfSpace_isSet)
	if obj_ProofOfSpace_isSet {
		obj.ProofOfSpace.ToBytes(buf)
	}
	obj_CcSignagePoint_isSet := !(obj.CcSignagePoint == nil)
	utils.BoolToBytes(buf, obj_CcSignagePoint_isSet)
	if obj_CcSignagePoint_isSet {
		obj.CcSignagePoint.ToBytes(buf)
	}
	obj_CcInfusionPoint_isSet := !(obj.CcInfusionPoint == nil)
	utils.BoolToBytes(buf, obj_CcInfusionPoint_isSet)
	if obj_CcInfusionPoint_isSet {
		obj.CcInfusionPoint.ToBytes(buf)
	}
	obj_IccInfusionPoint_isSet := !(obj.IccInfusionPoint == nil)
	utils.BoolToBytes(buf, obj_IccInfusionPoint_isSet)
	if obj_IccInfusionPoint_isSet {
		obj.IccInfusionPoint.ToBytes(buf)
	}
	obj_CcSpVdfInfo_isSet := !(obj.CcSpVdfInfo == nil)
	utils.BoolToBytes(buf, obj_CcSpVdfInfo_isSet)
	if obj_CcSpVdfInfo_isSet {
		obj.CcSpVdfInfo.ToBytes(buf)
	}
	obj_SignagePointIndex_isSet := !(obj.SignagePointIndex == 0)
	utils.BoolToBytes(buf, obj_SignagePointIndex_isSet)
	if obj_SignagePointIndex_isSet {
		utils.Uint8ToBytes(buf, obj.SignagePointIndex)
	}
	obj_CcSlotEnd_isSet := !(obj.CcSlotEnd == nil)
	utils.BoolToBytes(buf, obj_CcSlotEnd_isSet)
	if obj_CcSlotEnd_isSet {
		obj.CcSlotEnd.ToBytes(buf)
	}
	obj_IccSlotEnd_isSet := !(obj.IccSlotEnd == nil)
	utils.BoolToBytes(buf, obj_IccSlotEnd_isSet)
	if obj_IccSlotEnd_isSet {
		obj.IccSlotEnd.ToBytes(buf)
	}
	obj_CcSlotEndInfo_isSet := !(obj.CcSlotEndInfo == nil)
	utils.BoolToBytes(buf, obj_CcSlotEndInfo_isSet)
	if obj_CcSlotEndInfo_isSet {
		obj.CcSlotEndInfo.ToBytes(buf)
	}
	obj_IccSlotEndInfo_isSet := !(obj.IccSlotEndInfo == nil)
	utils.BoolToBytes(buf, obj_IccSlotEndInfo_isSet)
	if obj_IccSlotEndInfo_isSet {
		obj.IccSlotEndInfo.ToBytes(buf)
	}
	obj_CcIpVdfInfo_isSet := !(obj.CcIpVdfInfo == nil)
	utils.BoolToBytes(buf, obj_CcIpVdfInfo_isSet)
	if obj_CcIpVdfInfo_isSet {
		obj.CcIpVdfInfo.ToBytes(buf)
	}
	obj_IccIpVdfInfo_isSet := !(obj.IccIpVdfInfo == nil)
	utils.BoolToBytes(buf, obj_IccIpVdfInfo_isSet)
	if obj_IccIpVdfInfo_isSet {
		obj.IccIpVdfInfo.ToBytes(buf)
	}
	obj_TotalIters_isSet := !(obj.TotalIters == nil)
	utils.BoolToBytes(buf, obj_TotalIters_isSet)
	if obj_TotalIters_isSet {
		utils.Uint128ToBytes(buf, obj.TotalIters)
	}
}

// This is a list of coins being spent along with their solution programs, and a single
// aggregated signature. This is the object that most closely corresponds to a bitcoin
// transaction (although because of non-interactive signature aggregation, the boundaries
// between transactions are more flexible than in bitcoin).
type SpendBundle struct {
	CoinSolutions       []CoinSolution
	AggregatedSignature G2Element
}

func (obj *SpendBundle) FromBytes(buf *utils.ParseBuf) {
	len_obj_CoinSolutions := buf.Uint32()
	obj.CoinSolutions = make([]CoinSolution, len_obj_CoinSolutions)
	for i := uint32(0); i < len_obj_CoinSolutions; i++ {
		obj.CoinSolutions[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	obj.AggregatedSignature.FromBytes(buf)
}

func (obj SpendBundle) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.CoinSolutions)))
	for _, item := range obj.CoinSolutions {
		item.ToBytes(buf)
	}
	obj.AggregatedSignature.ToBytes(buf)
}

// This is a rather disparate data structure that validates coin transfers. It's generally populated
// with data from different sources, since burned coins are identified by name, so it is built up
// more often that it is streamed.
type CoinSolution struct {
	Coin         Coin
	PuzzleReveal SerializedProgram
	Solution     SerializedProgram
}

func (obj *CoinSolution) FromBytes(buf *utils.ParseBuf) {
	obj.Coin.FromBytes(buf)
	obj.PuzzleReveal.FromBytes(buf)
	obj.Solution.FromBytes(buf)
}

func (obj CoinSolution) ToBytes(buf *[]byte) {
	obj.Coin.ToBytes(buf)
	obj.PuzzleReveal.ToBytes(buf)
	obj.Solution.ToBytes(buf)
}

type UnfinishedBlock struct {
	// If first sb
	FinishedSubSlots []EndOfSubSlotBundle
	// Reward chain trunk data
	RewardChainBlock RewardChainBlockUnfinished
	// (optional) If not first sp in sub-slot
	ChallengeChainSpProof *VDFProof
	// (optional) If not first sp in sub-slot
	RewardChainSpProof *VDFProof
	// Reward chain foliage data
	Foliage Foliage
	// (optional) Reward chain foliage data (tx block)
	FoliageTransactionBlock *FoliageTransactionBlock
	// (optional) Reward chain foliage data (tx block additional)
	TransactionsInfo *TransactionsInfo
	// (optional) Program that generates transactions
	TransactionsGenerator *SerializedProgram
	// List of block heights of previous generators referenced in this block
	TransactionsGeneratorRefList []uint32
}

func (obj *UnfinishedBlock) FromBytes(buf *utils.ParseBuf) {
	len_obj_FinishedSubSlots := buf.Uint32()
	obj.FinishedSubSlots = make([]EndOfSubSlotBundle, len_obj_FinishedSubSlots)
	for i := uint32(0); i < len_obj_FinishedSubSlots; i++ {
		obj.FinishedSubSlots[i].FromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	obj.RewardChainBlock.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.ChallengeChainSpProof = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t VDFProof
		t.FromBytes(buf)
		obj.RewardChainSpProof = &t
	}
	obj.Foliage.FromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t FoliageTransactionBlock
		t.FromBytes(buf)
		obj.FoliageTransactionBlock = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t TransactionsInfo
		t.FromBytes(buf)
		obj.TransactionsInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t SerializedProgram
		t.FromBytes(buf)
		obj.TransactionsGenerator = &t
	}
	len_obj_TransactionsGeneratorRefList := buf.Uint32()
	obj.TransactionsGeneratorRefList = make([]uint32, len_obj_TransactionsGeneratorRefList)
	for i := uint32(0); i < len_obj_TransactionsGeneratorRefList; i++ {
		obj.TransactionsGeneratorRefList[i] = buf.Uint32()
		if buf.Err() != nil {
			return
		}
	}
}

func (obj UnfinishedBlock) ToBytes(buf *[]byte) {
	utils.Uint32ToBytes(buf, uint32(len(obj.FinishedSubSlots)))
	for _, item := range obj.FinishedSubSlots {
		item.ToBytes(buf)
	}
	obj.RewardChainBlock.ToBytes(buf)
	obj_ChallengeChainSpProof_isSet := !(obj.ChallengeChainSpProof == nil)
	utils.BoolToBytes(buf, obj_ChallengeChainSpProof_isSet)
	if obj_ChallengeChainSpProof_isSet {
		obj.ChallengeChainSpProof.ToBytes(buf)
	}
	obj_RewardChainSpProof_isSet := !(obj.RewardChainSpProof == nil)
	utils.BoolToBytes(buf, obj_RewardChainSpProof_isSet)
	if obj_RewardChainSpProof_isSet {
		obj.RewardChainSpProof.ToBytes(buf)
	}
	obj.Foliage.ToBytes(buf)
	obj_FoliageTransactionBlock_isSet := !(obj.FoliageTransactionBlock == nil)
	utils.BoolToBytes(buf, obj_FoliageTransactionBlock_isSet)
	if obj_FoliageTransactionBlock_isSet {
		obj.FoliageTransactionBlock.ToBytes(buf)
	}
	obj_TransactionsInfo_isSet := !(obj.TransactionsInfo == nil)
	utils.BoolToBytes(buf, obj_TransactionsInfo_isSet)
	if obj_TransactionsInfo_isSet {
		obj.TransactionsInfo.ToBytes(buf)
	}
	obj_TransactionsGenerator_isSet := !(obj.TransactionsGenerator == nil)
	utils.BoolToBytes(buf, obj_TransactionsGenerator_isSet)
	if obj_TransactionsGenerator_isSet {
		obj.TransactionsGenerator.ToBytes(buf)
	}
	utils.Uint32ToBytes(buf, uint32(len(obj.TransactionsGeneratorRefList)))
	for _, item := range obj.TransactionsGeneratorRefList {
		utils.Uint32ToBytes(buf, item)
	}
}

type TimestampedPeerInfo struct {
	Host      string
	Port      uint16
	Timestamp uint64
}

func (obj *TimestampedPeerInfo) FromBytes(buf *utils.ParseBuf) {
	obj.Host = buf.String()
	obj.Port = buf.Uint16()
	obj.Timestamp = buf.Uint64()
}

func (obj TimestampedPeerInfo) ToBytes(buf *[]byte) {
	utils.StringToBytes(buf, obj.Host)
	utils.Uint16ToBytes(buf, obj.Port)
	utils.Uint64ToBytes(buf, obj.Timestamp)
}
