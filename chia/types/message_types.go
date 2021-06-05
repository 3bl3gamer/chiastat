package types

// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/protocols/protocol_message_types.py
const (
	// Shared protocol (all services)
	MSG_HANDSHAKE uint8 = 1

	// Harvester protocol (harvester <-> farmer)
	MSG_HARVESTER_HANDSHAKE         uint8 = 3
	MSG_NEW_SIGNAGE_POINT_HARVESTER uint8 = 4
	MSG_NEW_PROOF_OF_SPACE          uint8 = 5
	MSG_REQUEST_SIGNATURES          uint8 = 6
	MSG_RESPOND_SIGNATURES          uint8 = 7

	// Farmer protocol (farmer <-> full_node)
	MSG_NEW_SIGNAGE_POINT      uint8 = 8
	MSG_DECLARE_PROOF_OF_SPACE uint8 = 9
	MSG_REQUEST_SIGNED_VALUES  uint8 = 10
	MSG_SIGNED_VALUES          uint8 = 11
	MSG_FARMING_INFO           uint8 = 12

	// Timelord protocol (timelord <-> full_node)
	MSG_NEW_PEAK_TIMELORD             uint8 = 13
	MSG_NEW_UNFINISHED_BLOCK_TIMELORD uint8 = 14
	MSG_NEW_INFUSION_POINT_VDF        uint8 = 15
	MSG_NEW_SIGNAGE_POINT_VDF         uint8 = 16
	MSG_NEW_END_OF_SUB_SLOT_VDF       uint8 = 17
	MSG_REQUEST_COMPACT_PROOF_OF_TIME uint8 = 18
	MSG_RESPOND_COMPACT_PROOF_OF_TIME uint8 = 19

	// Full node protocol (full_node <-> full_node)
	MSG_NEW_PEAK                                 uint8 = 20
	MSG_NEW_TRANSACTION                          uint8 = 21
	MSG_REQUEST_TRANSACTION                      uint8 = 22
	MSG_RESPOND_TRANSACTION                      uint8 = 23
	MSG_REQUEST_PROOF_OF_WEIGHT                  uint8 = 24
	MSG_RESPOND_PROOF_OF_WEIGHT                  uint8 = 25
	MSG_REQUEST_BLOCK                            uint8 = 26
	MSG_RESPOND_BLOCK                            uint8 = 27
	MSG_REJECT_BLOCK                             uint8 = 28
	MSG_REQUEST_BLOCKS                           uint8 = 29
	MSG_RESPOND_BLOCKS                           uint8 = 30
	MSG_REJECT_BLOCKS                            uint8 = 31
	MSG_NEW_UNFINISHED_BLOCK                     uint8 = 32
	MSG_REQUEST_UNFINISHED_BLOCK                 uint8 = 33
	MSG_RESPOND_UNFINISHED_BLOCK                 uint8 = 34
	MSG_NEW_SIGNAGE_POINT_OR_END_OF_SUB_SLOT     uint8 = 35
	MSG_REQUEST_SIGNAGE_POINT_OR_END_OF_SUB_SLOT uint8 = 36
	MSG_RESPOND_SIGNAGE_POINT                    uint8 = 37
	MSG_RESPOND_END_OF_SUB_SLOT                  uint8 = 38
	MSG_REQUEST_MEMPOOL_TRANSACTIONS             uint8 = 39
	MSG_REQUEST_COMPACT_VDF                      uint8 = 40
	MSG_RESPOND_COMPACT_VDF                      uint8 = 41
	MSG_NEW_COMPACT_VDF                          uint8 = 42
	MSG_REQUEST_PEERS                            uint8 = 43
	MSG_RESPOND_PEERS                            uint8 = 44

	// Wallet protocol (wallet <-> full_node)
	MSG_REQUEST_PUZZLE_SOLUTION  uint8 = 45
	MSG_RESPOND_PUZZLE_SOLUTION  uint8 = 46
	MSG_REJECT_PUZZLE_SOLUTION   uint8 = 47
	MSG_SEND_TRANSACTION         uint8 = 48
	MSG_TRANSACTION_ACK          uint8 = 49
	MSG_NEW_PEAK_WALLET          uint8 = 50
	MSG_REQUEST_BLOCK_HEADER     uint8 = 51
	MSG_RESPOND_BLOCK_HEADER     uint8 = 52
	MSG_REJECT_HEADER_REQUEST    uint8 = 53
	MSG_REQUEST_REMOVALS         uint8 = 54
	MSG_RESPOND_REMOVALS         uint8 = 55
	MSG_REJECT_REMOVALS_REQUEST  uint8 = 56
	MSG_REQUEST_ADDITIONS        uint8 = 57
	MSG_RESPOND_ADDITIONS        uint8 = 58
	MSG_REJECT_ADDITIONS_REQUEST uint8 = 59
	MSG_REQUEST_HEADER_BLOCKS    uint8 = 60
	MSG_REJECT_HEADER_BLOCKS     uint8 = 61
	MSG_RESPOND_HEADER_BLOCKS    uint8 = 62

	// Introducer protocol (introducer <-> full_node)
	MSG_REQUEST_PEERS_INTRODUCER uint8 = 63
	MSG_RESPOND_PEERS_INTRODUCER uint8 = 64

	// Simulator protocol
	MSG_FARM_NEW_BLOCK uint8 = 65
)
