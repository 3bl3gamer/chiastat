package chia

import (
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
	InfusedChallengeVdfOutput ClassgroupElement
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
	PrevTransactionBlockHash [32]byte
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
	SubEpochSummaryIncluded SubEpochSummary
}

func BlockRecordFromBytes(buf *ParseBuf) (obj BlockRecord) {
	obj.HeaderHash = Bytes32FromBytes(buf)
	obj.PrevHash = Bytes32FromBytes(buf)
	obj.Height = Uint32FromBytes(buf)
	obj.Weight = Uint128FromBytes(buf)
	obj.TotalIters = Uint128FromBytes(buf)
	obj.SignagePointIndex = Uint8FromBytes(buf)
	obj.ChallengeVdfOutput = ClassgroupElementFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeVdfOutput = ClassgroupElementFromBytes(buf)
	}
	obj.RewardInfusionNewChallenge = Bytes32FromBytes(buf)
	obj.ChallengeBlockInfoHash = Bytes32FromBytes(buf)
	obj.SubSlotIters = Uint64FromBytes(buf)
	obj.PoolPuzzleHash = Bytes32FromBytes(buf)
	obj.FarmerPuzzleHash = Bytes32FromBytes(buf)
	obj.RequiredIters = Uint64FromBytes(buf)
	obj.Deficit = Uint8FromBytes(buf)
	obj.Overflow = BoolFromBytes(buf)
	obj.PrevTransactionBlockHeight = Uint32FromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.Timestamp = Uint64FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.PrevTransactionBlockHash = Bytes32FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.Fees = Uint64FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		len_obj_RewardClaimsIncorporated := Uint32FromBytes(buf)
		obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
		for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
			obj.RewardClaimsIncorporated[i] = CoinFromBytes(buf)
			if buf.err != nil {
				return
			}
		}
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		len_obj_FinishedChallengeSlotHashes := Uint32FromBytes(buf)
		obj.FinishedChallengeSlotHashes = make([][32]byte, len_obj_FinishedChallengeSlotHashes)
		for i := uint32(0); i < len_obj_FinishedChallengeSlotHashes; i++ {
			obj.FinishedChallengeSlotHashes[i] = Bytes32FromBytes(buf)
			if buf.err != nil {
				return
			}
		}
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		len_obj_FinishedInfusedChallengeSlotHashes := Uint32FromBytes(buf)
		obj.FinishedInfusedChallengeSlotHashes = make([][32]byte, len_obj_FinishedInfusedChallengeSlotHashes)
		for i := uint32(0); i < len_obj_FinishedInfusedChallengeSlotHashes; i++ {
			obj.FinishedInfusedChallengeSlotHashes[i] = Bytes32FromBytes(buf)
			if buf.err != nil {
				return
			}
		}
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		len_obj_FinishedRewardSlotHashes := Uint32FromBytes(buf)
		obj.FinishedRewardSlotHashes = make([][32]byte, len_obj_FinishedRewardSlotHashes)
		for i := uint32(0); i < len_obj_FinishedRewardSlotHashes; i++ {
			obj.FinishedRewardSlotHashes[i] = Bytes32FromBytes(buf)
			if buf.err != nil {
				return
			}
		}
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.SubEpochSummaryIncluded = SubEpochSummaryFromBytes(buf)
	}
	return
}

// This structure is used in the body for the reward and fees genesis coins.
type Coin struct {
	ParentCoinInfo [32]byte
	PuzzleHash     [32]byte
	Amount         uint64
}

func CoinFromBytes(buf *ParseBuf) (obj Coin) {
	obj.ParentCoinInfo = Bytes32FromBytes(buf)
	obj.PuzzleHash = Bytes32FromBytes(buf)
	obj.Amount = Uint64FromBytes(buf)
	return
}

// Represents a classgroup element (a,b,c) where a, b, and c are 512 bit signed integers. However this is using
// a compressed representation. VDF outputs are a single classgroup element. VDF proofs can also be one classgroup
// element (or multiple).
type ClassgroupElement struct {
	Data [100]byte
}

func ClassgroupElementFromBytes(buf *ParseBuf) (obj ClassgroupElement) {
	obj.Data = Bytes100FromBytes(buf)
	return
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

func SubEpochSummaryFromBytes(buf *ParseBuf) (obj SubEpochSummary) {
	obj.PrevSubepochSummaryHash = Bytes32FromBytes(buf)
	obj.RewardChainHash = Bytes32FromBytes(buf)
	obj.NumBlocksOverflow = Uint8FromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.NewDifficulty = Uint64FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.NewSubSlotIters = Uint64FromBytes(buf)
	}
	return
}

type FullBlock struct {
	// If first sb
	FinishedSubSlots []EndOfSubSlotBundle
	// Reward chain trunk data
	RewardChainBlock RewardChainBlock
	// (optional) If not first sp in sub-slot
	ChallengeChainSpProof VDFProof
	ChallengeChainIpProof VDFProof
	// (optional) If not first sp in sub-slot
	RewardChainSpProof VDFProof
	RewardChainIpProof VDFProof
	// (optional) Iff deficit < 4
	InfusedChallengeChainIpProof VDFProof
	// Reward chain foliage data
	Foliage Foliage
	// (optional) Reward chain foliage data (tx block)
	FoliageTransactionBlock FoliageTransactionBlock
	// (optional) Reward chain foliage data (tx block additional)
	TransactionsInfo TransactionsInfo
	// (optional) Program that generates transactions
	TransactionsGenerator SerializedProgram
	// List of block heights of previous generators referenced in this block
	TransactionsGeneratorRefList []uint32
}

func FullBlockFromBytes(buf *ParseBuf) (obj FullBlock) {
	len_obj_FinishedSubSlots := Uint32FromBytes(buf)
	obj.FinishedSubSlots = make([]EndOfSubSlotBundle, len_obj_FinishedSubSlots)
	for i := uint32(0); i < len_obj_FinishedSubSlots; i++ {
		obj.FinishedSubSlots[i] = EndOfSubSlotBundleFromBytes(buf)
		if buf.err != nil {
			return
		}
	}
	obj.RewardChainBlock = RewardChainBlockFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.ChallengeChainSpProof = VDFProofFromBytes(buf)
	}
	obj.ChallengeChainIpProof = VDFProofFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.RewardChainSpProof = VDFProofFromBytes(buf)
	}
	obj.RewardChainIpProof = VDFProofFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChainIpProof = VDFProofFromBytes(buf)
	}
	obj.Foliage = FoliageFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.FoliageTransactionBlock = FoliageTransactionBlockFromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.TransactionsInfo = TransactionsInfoFromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.TransactionsGenerator = SerializedProgramFromBytes(buf)
	}
	len_obj_TransactionsGeneratorRefList := Uint32FromBytes(buf)
	obj.TransactionsGeneratorRefList = make([]uint32, len_obj_TransactionsGeneratorRefList)
	for i := uint32(0); i < len_obj_TransactionsGeneratorRefList; i++ {
		obj.TransactionsGeneratorRefList[i] = Uint32FromBytes(buf)
		if buf.err != nil {
			return
		}
	}
	return
}

type EndOfSubSlotBundle struct {
	ChallengeChain ChallengeChainSubSlot
	// (optional)
	InfusedChallengeChain InfusedChallengeChainSubSlot
	RewardChain           RewardChainSubSlot
	Proofs                SubSlotProofs
}

func EndOfSubSlotBundleFromBytes(buf *ParseBuf) (obj EndOfSubSlotBundle) {
	obj.ChallengeChain = ChallengeChainSubSlotFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChain = InfusedChallengeChainSubSlotFromBytes(buf)
	}
	obj.RewardChain = RewardChainSubSlotFromBytes(buf)
	obj.Proofs = SubSlotProofsFromBytes(buf)
	return
}

type VDFProof struct {
	WitnessType          uint8
	Witness              []byte
	NormalizedToIdentity bool
}

func VDFProofFromBytes(buf *ParseBuf) (obj VDFProof) {
	obj.WitnessType = Uint8FromBytes(buf)
	obj.Witness = BytesFromBytes(buf)
	obj.NormalizedToIdentity = BoolFromBytes(buf)
	return
}

type VDFInfo struct {
	// Used to generate the discriminant (VDF group)
	Challenge          [32]byte
	NumberOfIterations uint64
	Output             ClassgroupElement
}

func VDFInfoFromBytes(buf *ParseBuf) (obj VDFInfo) {
	obj.Challenge = Bytes32FromBytes(buf)
	obj.NumberOfIterations = Uint64FromBytes(buf)
	obj.Output = ClassgroupElementFromBytes(buf)
	return
}

type Foliage struct {
	PrevBlockHash             [32]byte
	RewardBlockHash           [32]byte
	FoliageBlockData          FoliageBlockData
	FoliageBlockDataSignature G2Element
	// (optional)
	FoliageTransactionBlockHash [32]byte
	// (optional)
	FoliageTransactionBlockSignature G2Element
}

func FoliageFromBytes(buf *ParseBuf) (obj Foliage) {
	obj.PrevBlockHash = Bytes32FromBytes(buf)
	obj.RewardBlockHash = Bytes32FromBytes(buf)
	obj.FoliageBlockData = FoliageBlockDataFromBytes(buf)
	obj.FoliageBlockDataSignature = G2ElementFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.FoliageTransactionBlockHash = Bytes32FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.FoliageTransactionBlockSignature = G2ElementFromBytes(buf)
	}
	return
}

type FoliageTransactionBlock struct {
	PrevTransactionBlockHash [32]byte
	Timestamp                uint64
	FilterHash               [32]byte
	AdditionsRoot            [32]byte
	RemovalsRoot             [32]byte
	TransactionsInfoHash     [32]byte
}

func FoliageTransactionBlockFromBytes(buf *ParseBuf) (obj FoliageTransactionBlock) {
	obj.PrevTransactionBlockHash = Bytes32FromBytes(buf)
	obj.Timestamp = Uint64FromBytes(buf)
	obj.FilterHash = Bytes32FromBytes(buf)
	obj.AdditionsRoot = Bytes32FromBytes(buf)
	obj.RemovalsRoot = Bytes32FromBytes(buf)
	obj.TransactionsInfoHash = Bytes32FromBytes(buf)
	return
}

type FoliageBlockData struct {
	UnfinishedRewardBlockHash [32]byte
	PoolTarget                PoolTarget
	// (optional) Iff ProofOfSpace has a pool pk
	PoolSignature          G2Element
	FarmerRewardPuzzleHash [32]byte
	// Used for future updates. Can be any 32 byte value initially
	ExtensionData [32]byte
}

func FoliageBlockDataFromBytes(buf *ParseBuf) (obj FoliageBlockData) {
	obj.UnfinishedRewardBlockHash = Bytes32FromBytes(buf)
	obj.PoolTarget = PoolTargetFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.PoolSignature = G2ElementFromBytes(buf)
	}
	obj.FarmerRewardPuzzleHash = Bytes32FromBytes(buf)
	obj.ExtensionData = Bytes32FromBytes(buf)
	return
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

func TransactionsInfoFromBytes(buf *ParseBuf) (obj TransactionsInfo) {
	obj.GeneratorRoot = Bytes32FromBytes(buf)
	obj.GeneratorRefsRoot = Bytes32FromBytes(buf)
	obj.AggregatedSignature = G2ElementFromBytes(buf)
	obj.Fees = Uint64FromBytes(buf)
	obj.Cost = Uint64FromBytes(buf)
	len_obj_RewardClaimsIncorporated := Uint32FromBytes(buf)
	obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
	for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
		obj.RewardClaimsIncorporated[i] = CoinFromBytes(buf)
		if buf.err != nil {
			return
		}
	}
	return
}

type RewardChainBlock struct {
	Weight               *big.Int
	Height               uint32
	TotalIters           *big.Int
	SignagePointIndex    uint8
	PosSsCcChallengeHash [32]byte
	ProofOfSpace         ProofOfSpace
	// (optional) Not present for first sp in slot
	ChallengeChainSpVdf       VDFInfo
	ChallengeChainSpSignature G2Element
	ChallengeChainIpVdf       VDFInfo
	// (optional) Not present for first sp in slot
	RewardChainSpVdf       VDFInfo
	RewardChainSpSignature G2Element
	RewardChainIpVdf       VDFInfo
	// (optional) Iff deficit < 16
	InfusedChallengeChainIpVdf VDFInfo
	IsTransactionBlock         bool
}

func RewardChainBlockFromBytes(buf *ParseBuf) (obj RewardChainBlock) {
	obj.Weight = Uint128FromBytes(buf)
	obj.Height = Uint32FromBytes(buf)
	obj.TotalIters = Uint128FromBytes(buf)
	obj.SignagePointIndex = Uint8FromBytes(buf)
	obj.PosSsCcChallengeHash = Bytes32FromBytes(buf)
	obj.ProofOfSpace = ProofOfSpaceFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.ChallengeChainSpVdf = VDFInfoFromBytes(buf)
	}
	obj.ChallengeChainSpSignature = G2ElementFromBytes(buf)
	obj.ChallengeChainIpVdf = VDFInfoFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.RewardChainSpVdf = VDFInfoFromBytes(buf)
	}
	obj.RewardChainSpSignature = G2ElementFromBytes(buf)
	obj.RewardChainIpVdf = VDFInfoFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChainIpVdf = VDFInfoFromBytes(buf)
	}
	obj.IsTransactionBlock = BoolFromBytes(buf)
	return
}

type ChallengeChainSubSlot struct {
	ChallengeChainEndOfSlotVdf VDFInfo
	// (optional) Only at the end of a slot
	InfusedChallengeChainSubSlotHash [32]byte
	// (optional) Only once per sub-epoch, and one sub-epoch delayed
	SubepochSummaryHash [32]byte
	// (optional) Only at the end of epoch, sub-epoch, and slot
	NewSubSlotIters uint64
	// (optional) Only at the end of epoch, sub-epoch, and slot
	NewDifficulty uint64
}

func ChallengeChainSubSlotFromBytes(buf *ParseBuf) (obj ChallengeChainSubSlot) {
	obj.ChallengeChainEndOfSlotVdf = VDFInfoFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChainSubSlotHash = Bytes32FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.SubepochSummaryHash = Bytes32FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.NewSubSlotIters = Uint64FromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.NewDifficulty = Uint64FromBytes(buf)
	}
	return
}

type InfusedChallengeChainSubSlot struct {
	InfusedChallengeChainEndOfSlotVdf VDFInfo
}

func InfusedChallengeChainSubSlotFromBytes(buf *ParseBuf) (obj InfusedChallengeChainSubSlot) {
	obj.InfusedChallengeChainEndOfSlotVdf = VDFInfoFromBytes(buf)
	return
}

type RewardChainSubSlot struct {
	EndOfSlotVdf              VDFInfo
	ChallengeChainSubSlotHash [32]byte
	// (optional)
	InfusedChallengeChainSubSlotHash [32]byte
	// 16 or less. usually zero
	Deficit uint8
}

func RewardChainSubSlotFromBytes(buf *ParseBuf) (obj RewardChainSubSlot) {
	obj.EndOfSlotVdf = VDFInfoFromBytes(buf)
	obj.ChallengeChainSubSlotHash = Bytes32FromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChainSubSlotHash = Bytes32FromBytes(buf)
	}
	obj.Deficit = Uint8FromBytes(buf)
	return
}

type SubSlotProofs struct {
	ChallengeChainSlotProof VDFProof
	// (optional)
	InfusedChallengeChainSlotProof VDFProof
	RewardChainSlotProof           VDFProof
}

func SubSlotProofsFromBytes(buf *ParseBuf) (obj SubSlotProofs) {
	obj.ChallengeChainSlotProof = VDFProofFromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.InfusedChallengeChainSlotProof = VDFProofFromBytes(buf)
	}
	obj.RewardChainSlotProof = VDFProofFromBytes(buf)
	return
}

type PoolTarget struct {
	PuzzleHash [32]byte
	// A max height of 0 means it is valid forever
	MaxHeight uint32
}

func PoolTargetFromBytes(buf *ParseBuf) (obj PoolTarget) {
	obj.PuzzleHash = Bytes32FromBytes(buf)
	obj.MaxHeight = Uint32FromBytes(buf)
	return
}

type ProofOfSpace struct {
	Challenge [32]byte
	// (optional) Only one of these two should be present
	PoolPublicKey G1Element
	// (optional)
	PoolContractPuzzleHash [32]byte
	PlotPublicKey          G1Element
	Size                   uint8
	Proof                  []byte
}

func ProofOfSpaceFromBytes(buf *ParseBuf) (obj ProofOfSpace) {
	obj.Challenge = Bytes32FromBytes(buf)
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.PoolPublicKey = G1ElementFromBytes(buf)
	}
	if flag := BoolFromBytes(buf); buf.err == nil && flag {
		obj.PoolContractPuzzleHash = Bytes32FromBytes(buf)
	}
	obj.PlotPublicKey = G1ElementFromBytes(buf)
	obj.Size = Uint8FromBytes(buf)
	obj.Proof = BytesFromBytes(buf)
	return
}

type G1Element struct{ Bytes []byte }

func G1ElementFromBytes(buf *ParseBuf) (obj G1Element) {
	obj.Bytes = BytesNFromBytes(buf, 48)
	return
}

type G2Element struct{ Bytes []byte }

func G2ElementFromBytes(buf *ParseBuf) (obj G2Element) {
	obj.Bytes = BytesNFromBytes(buf, 96)
	return
}
