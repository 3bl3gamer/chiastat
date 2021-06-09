// Generated with `go genetare`. Do not edit.
package types

const (
	NODE_FULL       = 1
	NODE_HARVESTER  = 2
	NODE_FARMER     = 3
	NODE_TIMELORD   = 4
	NODE_INTRODUCER = 5
	NODE_WALLET     = 6
)

func NodeTypeName(type_ uint8) (string, bool) {
	switch type_ {
	case NODE_FULL:
		return "FULL_NODE", true
	case NODE_HARVESTER:
		return "HARVESTER", true
	case NODE_FARMER:
		return "FARMER", true
	case NODE_TIMELORD:
		return "TIMELORD", true
	case NODE_INTRODUCER:
		return "INTRODUCER", true
	case NODE_WALLET:
		return "WALLET", true
	default:
		return "UNKNOWN", false
	}
}
