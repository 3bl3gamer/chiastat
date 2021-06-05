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
