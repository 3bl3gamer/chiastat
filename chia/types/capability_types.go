package types

// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/protocols/shared_protocol.py#L18
// Capabilities can be added here when new features are added to the protocol
// These are passed in as uint16 into the Handshake
const (
	CAP_BASE uint16 = 1 // Base capability just means it supports the chia protocol at mainnet
)
