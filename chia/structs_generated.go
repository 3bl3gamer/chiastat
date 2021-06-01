package chia

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
	SubEpochSummaryIncluded *SubEpochSummary
}

func BlockRecordFromBytes(buf *utils.ParseBuf) (obj BlockRecord) {
	obj.HeaderHash = buf.Bytes32()
	obj.PrevHash = buf.Bytes32()
	obj.Height = buf.Uint32()
	obj.Weight = buf.Uint128()
	obj.TotalIters = buf.Uint128()
	obj.SignagePointIndex = buf.Uint8()
	obj.ChallengeVdfOutput = ClassgroupElementFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = ClassgroupElementFromBytes(buf)
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
		obj.PrevTransactionBlockHash = buf.Bytes32()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.Fees = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		len_obj_RewardClaimsIncorporated := buf.Uint32()
		obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
		for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
			obj.RewardClaimsIncorporated[i] = CoinFromBytes(buf)
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
		var t = SubEpochSummaryFromBytes(buf)
		obj.SubEpochSummaryIncluded = &t
	}
	return
}

// This structure is used in the body for the reward and fees genesis coins.
type Coin struct {
	ParentCoinInfo [32]byte
	PuzzleHash     [32]byte
	Amount         uint64
}

func CoinFromBytes(buf *utils.ParseBuf) (obj Coin) {
	obj.ParentCoinInfo = buf.Bytes32()
	obj.PuzzleHash = buf.Bytes32()
	obj.Amount = buf.Uint64()
	return
}

// Represents a classgroup element (a,b,c) where a, b, and c are 512 bit signed integers. However this is using
// a compressed representation. VDF outputs are a single classgroup element. VDF proofs can also be one classgroup
// element (or multiple).
type ClassgroupElement struct {
	Data [100]byte
}

func ClassgroupElementFromBytes(buf *utils.ParseBuf) (obj ClassgroupElement) {
	obj.Data = buf.Bytes100()
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

func SubEpochSummaryFromBytes(buf *utils.ParseBuf) (obj SubEpochSummary) {
	obj.PrevSubepochSummaryHash = buf.Bytes32()
	obj.RewardChainHash = buf.Bytes32()
	obj.NumBlocksOverflow = buf.Uint8()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewDifficulty = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewSubSlotIters = buf.Uint64()
	}
	return
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

func FullBlockFromBytes(buf *utils.ParseBuf) (obj FullBlock) {
	len_obj_FinishedSubSlots := buf.Uint32()
	obj.FinishedSubSlots = make([]EndOfSubSlotBundle, len_obj_FinishedSubSlots)
	for i := uint32(0); i < len_obj_FinishedSubSlots; i++ {
		obj.FinishedSubSlots[i] = EndOfSubSlotBundleFromBytes(buf)
		if buf.Err() != nil {
			return
		}
	}
	obj.RewardChainBlock = RewardChainBlockFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFProofFromBytes(buf)
		obj.ChallengeChainSpProof = &t
	}
	obj.ChallengeChainIpProof = VDFProofFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFProofFromBytes(buf)
		obj.RewardChainSpProof = &t
	}
	obj.RewardChainIpProof = VDFProofFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFProofFromBytes(buf)
		obj.InfusedChallengeChainIpProof = &t
	}
	obj.Foliage = FoliageFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = FoliageTransactionBlockFromBytes(buf)
		obj.FoliageTransactionBlock = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = TransactionsInfoFromBytes(buf)
		obj.TransactionsInfo = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = SerializedProgramFromBytes(buf)
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
	return
}

type EndOfSubSlotBundle struct {
	ChallengeChain ChallengeChainSubSlot
	// (optional)
	InfusedChallengeChain *InfusedChallengeChainSubSlot
	RewardChain           RewardChainSubSlot
	Proofs                SubSlotProofs
}

func EndOfSubSlotBundleFromBytes(buf *utils.ParseBuf) (obj EndOfSubSlotBundle) {
	obj.ChallengeChain = ChallengeChainSubSlotFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = InfusedChallengeChainSubSlotFromBytes(buf)
		obj.InfusedChallengeChain = &t
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

func VDFProofFromBytes(buf *utils.ParseBuf) (obj VDFProof) {
	obj.WitnessType = buf.Uint8()
	obj.Witness = buf.Bytes()
	obj.NormalizedToIdentity = buf.Bool()
	return
}

type VDFInfo struct {
	// Used to generate the discriminant (VDF group)
	Challenge          [32]byte
	NumberOfIterations uint64
	Output             ClassgroupElement
}

func VDFInfoFromBytes(buf *utils.ParseBuf) (obj VDFInfo) {
	obj.Challenge = buf.Bytes32()
	obj.NumberOfIterations = buf.Uint64()
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
	FoliageTransactionBlockSignature *G2Element
}

func FoliageFromBytes(buf *utils.ParseBuf) (obj Foliage) {
	obj.PrevBlockHash = buf.Bytes32()
	obj.RewardBlockHash = buf.Bytes32()
	obj.FoliageBlockData = FoliageBlockDataFromBytes(buf)
	obj.FoliageBlockDataSignature = G2ElementFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.FoliageTransactionBlockHash = buf.Bytes32()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = G2ElementFromBytes(buf)
		obj.FoliageTransactionBlockSignature = &t
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

func FoliageTransactionBlockFromBytes(buf *utils.ParseBuf) (obj FoliageTransactionBlock) {
	obj.PrevTransactionBlockHash = buf.Bytes32()
	obj.Timestamp = buf.Uint64()
	obj.FilterHash = buf.Bytes32()
	obj.AdditionsRoot = buf.Bytes32()
	obj.RemovalsRoot = buf.Bytes32()
	obj.TransactionsInfoHash = buf.Bytes32()
	return
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

func FoliageBlockDataFromBytes(buf *utils.ParseBuf) (obj FoliageBlockData) {
	obj.UnfinishedRewardBlockHash = buf.Bytes32()
	obj.PoolTarget = PoolTargetFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = G2ElementFromBytes(buf)
		obj.PoolSignature = &t
	}
	obj.FarmerRewardPuzzleHash = buf.Bytes32()
	obj.ExtensionData = buf.Bytes32()
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

func TransactionsInfoFromBytes(buf *utils.ParseBuf) (obj TransactionsInfo) {
	obj.GeneratorRoot = buf.Bytes32()
	obj.GeneratorRefsRoot = buf.Bytes32()
	obj.AggregatedSignature = G2ElementFromBytes(buf)
	obj.Fees = buf.Uint64()
	obj.Cost = buf.Uint64()
	len_obj_RewardClaimsIncorporated := buf.Uint32()
	obj.RewardClaimsIncorporated = make([]Coin, len_obj_RewardClaimsIncorporated)
	for i := uint32(0); i < len_obj_RewardClaimsIncorporated; i++ {
		obj.RewardClaimsIncorporated[i] = CoinFromBytes(buf)
		if buf.Err() != nil {
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

func RewardChainBlockFromBytes(buf *utils.ParseBuf) (obj RewardChainBlock) {
	obj.Weight = buf.Uint128()
	obj.Height = buf.Uint32()
	obj.TotalIters = buf.Uint128()
	obj.SignagePointIndex = buf.Uint8()
	obj.PosSsCcChallengeHash = buf.Bytes32()
	obj.ProofOfSpace = ProofOfSpaceFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFInfoFromBytes(buf)
		obj.ChallengeChainSpVdf = &t
	}
	obj.ChallengeChainSpSignature = G2ElementFromBytes(buf)
	obj.ChallengeChainIpVdf = VDFInfoFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFInfoFromBytes(buf)
		obj.RewardChainSpVdf = &t
	}
	obj.RewardChainSpSignature = G2ElementFromBytes(buf)
	obj.RewardChainIpVdf = VDFInfoFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFInfoFromBytes(buf)
		obj.InfusedChallengeChainIpVdf = &t
	}
	obj.IsTransactionBlock = buf.Bool()
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

func ChallengeChainSubSlotFromBytes(buf *utils.ParseBuf) (obj ChallengeChainSubSlot) {
	obj.ChallengeChainEndOfSlotVdf = VDFInfoFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.InfusedChallengeChainSubSlotHash = buf.Bytes32()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.SubepochSummaryHash = buf.Bytes32()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewSubSlotIters = buf.Uint64()
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.NewDifficulty = buf.Uint64()
	}
	return
}

type InfusedChallengeChainSubSlot struct {
	InfusedChallengeChainEndOfSlotVdf VDFInfo
}

func InfusedChallengeChainSubSlotFromBytes(buf *utils.ParseBuf) (obj InfusedChallengeChainSubSlot) {
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

func RewardChainSubSlotFromBytes(buf *utils.ParseBuf) (obj RewardChainSubSlot) {
	obj.EndOfSlotVdf = VDFInfoFromBytes(buf)
	obj.ChallengeChainSubSlotHash = buf.Bytes32()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.InfusedChallengeChainSubSlotHash = buf.Bytes32()
	}
	obj.Deficit = buf.Uint8()
	return
}

type SubSlotProofs struct {
	ChallengeChainSlotProof VDFProof
	// (optional)
	InfusedChallengeChainSlotProof *VDFProof
	RewardChainSlotProof           VDFProof
}

func SubSlotProofsFromBytes(buf *utils.ParseBuf) (obj SubSlotProofs) {
	obj.ChallengeChainSlotProof = VDFProofFromBytes(buf)
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = VDFProofFromBytes(buf)
		obj.InfusedChallengeChainSlotProof = &t
	}
	obj.RewardChainSlotProof = VDFProofFromBytes(buf)
	return
}

type PoolTarget struct {
	PuzzleHash [32]byte
	// A max height of 0 means it is valid forever
	MaxHeight uint32
}

func PoolTargetFromBytes(buf *utils.ParseBuf) (obj PoolTarget) {
	obj.PuzzleHash = buf.Bytes32()
	obj.MaxHeight = buf.Uint32()
	return
}

type ProofOfSpace struct {
	Challenge [32]byte
	// (optional) Only one of these two should be present
	PoolPublicKey *G1Element
	// (optional)
	PoolContractPuzzleHash [32]byte
	PlotPublicKey          G1Element
	Size                   uint8
	Proof                  []byte
}

func ProofOfSpaceFromBytes(buf *utils.ParseBuf) (obj ProofOfSpace) {
	obj.Challenge = buf.Bytes32()
	if flag := buf.Bool(); buf.Err() == nil && flag {
		var t = G1ElementFromBytes(buf)
		obj.PoolPublicKey = &t
	}
	if flag := buf.Bool(); buf.Err() == nil && flag {
		obj.PoolContractPuzzleHash = buf.Bytes32()
	}
	obj.PlotPublicKey = G1ElementFromBytes(buf)
	obj.Size = buf.Uint8()
	obj.Proof = buf.Bytes()
	return
}

type G1Element struct{ Bytes []byte }

func G1ElementFromBytes(buf *utils.ParseBuf) (obj G1Element) {
	obj.Bytes = buf.BytesN(48)
	return
}

type G2Element struct{ Bytes []byte }

func G2ElementFromBytes(buf *utils.ParseBuf) (obj G2Element) {
	obj.Bytes = buf.BytesN(96)
	return
}
