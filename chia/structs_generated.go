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
