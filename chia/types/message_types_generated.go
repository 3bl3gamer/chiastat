// Generated with `go genetare`. Do not edit.
package types

import "chiastat/chia/utils"

const (
	// Shared protocol (all services)
	MSG_HANDSHAKE = 1

	// Harvester protocol (harvester <-> farmer)
	MSG_NEW_SIGNAGE_POINT_HARVESTER = 4
	MSG_NEW_PROOF_OF_SPACE          = 5
	MSG_REQUEST_SIGNATURES          = 6
	MSG_RESPOND_SIGNATURES          = 7
	MSG_HARVESTER_HANDSHAKE         = 3

	// Farmer protocol (farmer <-> full_node)
	MSG_NEW_SIGNAGE_POINT      = 8
	MSG_DECLARE_PROOF_OF_SPACE = 9
	MSG_REQUEST_SIGNED_VALUES  = 10
	MSG_SIGNED_VALUES          = 11
	MSG_FARMING_INFO           = 12

	// Timelord protocol (timelord <-> full_node)
	MSG_NEW_UNFINISHED_BLOCK_TIMELORD = 14
	MSG_NEW_INFUSION_POINT_VDF        = 15
	MSG_NEW_SIGNAGE_POINT_VDF         = 16
	MSG_NEW_END_OF_SUB_SLOT_VDF       = 17
	MSG_REQUEST_COMPACT_PROOF_OF_TIME = 18
	MSG_RESPOND_COMPACT_PROOF_OF_TIME = 19
	MSG_NEW_PEAK_TIMELORD             = 13

	// Full node protocol (full_node <-> full_node)
	MSG_RESPOND_BLOCKS                           = 30
	MSG_NEW_SIGNAGE_POINT_OR_END_OF_SUB_SLOT     = 35
	MSG_REQUEST_SIGNAGE_POINT_OR_END_OF_SUB_SLOT = 36
	MSG_REQUEST_COMPACT_VDF                      = 40
	MSG_REQUEST_PEERS                            = 43
	MSG_NEW_TRANSACTION                          = 21
	MSG_RESPOND_PROOF_OF_WEIGHT                  = 25
	MSG_RESPOND_UNFINISHED_BLOCK                 = 34
	MSG_RESPOND_SIGNAGE_POINT                    = 37
	MSG_RESPOND_PEERS                            = 44
	MSG_RESPOND_TRANSACTION                      = 23
	MSG_REQUEST_TRANSACTION                      = 22
	MSG_RESPOND_BLOCK                            = 27
	MSG_REJECT_BLOCKS                            = 31
	MSG_REQUEST_UNFINISHED_BLOCK                 = 33
	MSG_RESPOND_END_OF_SUB_SLOT                  = 38
	MSG_NEW_PEAK                                 = 20
	MSG_REQUEST_BLOCK                            = 26
	MSG_REJECT_BLOCK                             = 28
	MSG_REQUEST_BLOCKS                           = 29
	MSG_NEW_UNFINISHED_BLOCK                     = 32
	MSG_REQUEST_MEMPOOL_TRANSACTIONS             = 39
	MSG_RESPOND_COMPACT_VDF                      = 41
	MSG_NEW_COMPACT_VDF                          = 42
	MSG_REQUEST_PROOF_OF_WEIGHT                  = 24

	// Wallet protocol (wallet <-> full_node)
	MSG_NEW_PEAK_WALLET          = 50
	MSG_REJECT_HEADER_REQUEST    = 53
	MSG_REQUEST_REMOVALS         = 54
	MSG_RESPOND_REMOVALS         = 55
	MSG_RESPOND_ADDITIONS        = 58
	MSG_REJECT_HEADER_BLOCKS     = 61
	MSG_RESPOND_PUZZLE_SOLUTION  = 46
	MSG_REJECT_PUZZLE_SOLUTION   = 47
	MSG_REQUEST_ADDITIONS        = 57
	MSG_REQUEST_HEADER_BLOCKS    = 60
	MSG_REQUEST_PUZZLE_SOLUTION  = 45
	MSG_RESPOND_BLOCK_HEADER     = 52
	MSG_TRANSACTION_ACK          = 49
	MSG_REJECT_ADDITIONS_REQUEST = 59
	MSG_REJECT_REMOVALS_REQUEST  = 56
	MSG_RESPOND_HEADER_BLOCKS    = 62
	MSG_SEND_TRANSACTION         = 48
	MSG_REQUEST_BLOCK_HEADER     = 51

	// Introducer protocol (introducer <-> full_node)
	MSG_REQUEST_PEERS_INTRODUCER = 63
	MSG_RESPOND_PEERS_INTRODUCER = 64

	// Simulator protocol
	MSG_FARM_NEW_BLOCK = 65
)

func MessageTypeStruct(type_ uint8) (utils.FromToBytes, bool) {
	switch type_ {
	case MSG_HANDSHAKE:
		return &Handshake{}, true
	case MSG_REQUEST_UNFINISHED_BLOCK:
		return &RequestUnfinishedBlock{}, true
	case MSG_RESPOND_END_OF_SUB_SLOT:
		return &RespondEndOfSubSlot{}, true
	case MSG_NEW_PEAK:
		return &NewPeak{}, true
	case MSG_REQUEST_TRANSACTION:
		return &RequestTransaction{}, true
	case MSG_RESPOND_BLOCK:
		return &RespondBlock{}, true
	case MSG_REJECT_BLOCKS:
		return &RejectBlocks{}, true
	case MSG_NEW_UNFINISHED_BLOCK:
		return &NewUnfinishedBlock{}, true
	case MSG_REQUEST_MEMPOOL_TRANSACTIONS:
		return &RequestMempoolTransactions{}, true
	case MSG_RESPOND_COMPACT_VDF:
		return &RespondCompactVDF{}, true
	case MSG_NEW_COMPACT_VDF:
		return &NewCompactVDF{}, true
	case MSG_REQUEST_PROOF_OF_WEIGHT:
		return &RequestProofOfWeight{}, true
	case MSG_REQUEST_BLOCK:
		return &RequestBlock{}, true
	case MSG_REJECT_BLOCK:
		return &RejectBlock{}, true
	case MSG_REQUEST_BLOCKS:
		return &RequestBlocks{}, true
	case MSG_REQUEST_COMPACT_VDF:
		return &RequestCompactVDF{}, true
	case MSG_REQUEST_PEERS:
		return &RequestPeers{}, true
	case MSG_NEW_TRANSACTION:
		return &NewTransaction{}, true
	case MSG_RESPOND_BLOCKS:
		return &RespondBlocks{}, true
	case MSG_NEW_SIGNAGE_POINT_OR_END_OF_SUB_SLOT:
		return &NewSignagePointOrEndOfSubSlot{}, true
	case MSG_REQUEST_SIGNAGE_POINT_OR_END_OF_SUB_SLOT:
		return &RequestSignagePointOrEndOfSubSlot{}, true
	case MSG_RESPOND_PEERS:
		return &RespondPeers{}, true
	case MSG_RESPOND_TRANSACTION:
		return &RespondTransaction{}, true
	case MSG_RESPOND_PROOF_OF_WEIGHT:
		return &RespondProofOfWeight{}, true
	case MSG_RESPOND_UNFINISHED_BLOCK:
		return &RespondUnfinishedBlock{}, true
	case MSG_RESPOND_SIGNAGE_POINT:
		return &RespondSignagePoint{}, true
	default:
		return nil, false
	}
}
