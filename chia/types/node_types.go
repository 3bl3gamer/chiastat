package types

// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/server/outbound_message.py#L10
const (
	NODE_FULL       uint8 = 1
	NODE_HARVESTER  uint8 = 2
	NODE_FARMER     uint8 = 3
	NODE_TIMELORD   uint8 = 4
	NODE_INTRODUCER uint8 = 5
	NODE_WALLET     uint8 = 6
)
