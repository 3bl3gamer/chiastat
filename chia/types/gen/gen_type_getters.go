package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type valueItem struct {
	name  string
	value int
}

type constGroup struct {
	comment     string
	values      []valueItem
	useInGetter bool
}

type constGroups struct {
	fnamePrefix      string
	getterPrefix     string
	constPrefix      string
	stringNameGetter bool
	stringNameMap    map[string]string
	imports          []string
	groups           []constGroup
}

var groups = []constGroups{
	{
		// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/server/outbound_message.py#L10
		fnamePrefix:      "node_types",
		getterPrefix:     "NodeType",
		constPrefix:      "NODE",
		stringNameGetter: true,
		stringNameMap: map[string]string{
			"FULL": "FULL_NODE",
		},
		groups: []constGroup{
			{
				values: []valueItem{
					{"FULL", 1},
					{"HARVESTER", 2},
					{"FARMER", 3},
					{"TIMELORD", 4},
					{"INTRODUCER", 5},
					{"WALLET", 6},
				},
			},
		},
	},
	{
		// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/protocols/protocol_message_types.py
		fnamePrefix:  "message_types",
		getterPrefix: "MessageType",
		constPrefix:  "MSG",
		imports:      []string{"chiastat/chia/utils"},
		groups: []constGroup{
			{
				comment: "Shared protocol (all services)",
				values: []valueItem{
					{"HANDSHAKE", 1},
				},
				useInGetter: true,
			},
			{
				comment: "Harvester protocol (harvester <-> farmer)",
				values: []valueItem{
					{"HARVESTER_HANDSHAKE", 3},
					{"NEW_SIGNAGE_POINT_HARVESTER", 4},
					{"NEW_PROOF_OF_SPACE", 5},
					{"REQUEST_SIGNATURES", 6},
					{"RESPOND_SIGNATURES", 7},
				},
			},
			{
				comment: "Farmer protocol (farmer <-> full_node)",
				values: []valueItem{
					{"NEW_SIGNAGE_POINT", 8},
					{"DECLARE_PROOF_OF_SPACE", 9},
					{"REQUEST_SIGNED_VALUES", 10},
					{"SIGNED_VALUES", 11},
					{"FARMING_INFO", 12},
				},
			},
			{
				comment: "Timelord protocol (timelord <-> full_node)",
				values: []valueItem{
					{"NEW_PEAK_TIMELORD", 13},
					{"NEW_UNFINISHED_BLOCK_TIMELORD", 14},
					{"NEW_INFUSION_POINT_VDF", 15},
					{"NEW_SIGNAGE_POINT_VDF", 16},
					{"NEW_END_OF_SUB_SLOT_VDF", 17},
					{"REQUEST_COMPACT_PROOF_OF_TIME", 18},
					{"RESPOND_COMPACT_PROOF_OF_TIME", 19},
				},
			},
			{
				comment: "Full node protocol (full_node <-> full_node)",
				values: []valueItem{
					{"NEW_PEAK", 20},
					{"NEW_TRANSACTION", 21},
					{"REQUEST_TRANSACTION", 22},
					{"RESPOND_TRANSACTION", 23},
					{"REQUEST_PROOF_OF_WEIGHT", 24},
					{"RESPOND_PROOF_OF_WEIGHT", 25},
					{"REQUEST_BLOCK", 26},
					{"RESPOND_BLOCK", 27},
					{"REJECT_BLOCK", 28},
					{"REQUEST_BLOCKS", 29},
					{"RESPOND_BLOCKS", 30},
					{"REJECT_BLOCKS", 31},
					{"NEW_UNFINISHED_BLOCK", 32},
					{"REQUEST_UNFINISHED_BLOCK", 33},
					{"RESPOND_UNFINISHED_BLOCK", 34},
					{"NEW_SIGNAGE_POINT_OR_END_OF_SUB_SLOT", 35},
					{"REQUEST_SIGNAGE_POINT_OR_END_OF_SUB_SLOT", 36},
					{"RESPOND_SIGNAGE_POINT", 37},
					{"RESPOND_END_OF_SUB_SLOT", 38},
					{"REQUEST_MEMPOOL_TRANSACTIONS", 39},
					{"REQUEST_COMPACT_VDF", 40},
					{"RESPOND_COMPACT_VDF", 41},
					{"NEW_COMPACT_VDF", 42},
					{"REQUEST_PEERS", 43},
					{"RESPOND_PEERS", 44},
				},
				useInGetter: true,
			},
			{
				comment: "Wallet protocol (wallet <-> full_node)",
				values: []valueItem{
					{"REQUEST_PUZZLE_SOLUTION", 45},
					{"RESPOND_PUZZLE_SOLUTION", 46},
					{"REJECT_PUZZLE_SOLUTION", 47},
					{"SEND_TRANSACTION", 48},
					{"TRANSACTION_ACK", 49},
					{"NEW_PEAK_WALLET", 50},
					{"REQUEST_BLOCK_HEADER", 51},
					{"RESPOND_BLOCK_HEADER", 52},
					{"REJECT_HEADER_REQUEST", 53},
					{"REQUEST_REMOVALS", 54},
					{"RESPOND_REMOVALS", 55},
					{"REJECT_REMOVALS_REQUEST", 56},
					{"REQUEST_ADDITIONS", 57},
					{"RESPOND_ADDITIONS", 58},
					{"REJECT_ADDITIONS_REQUEST", 59},
					{"REQUEST_HEADER_BLOCKS", 60},
					{"REJECT_HEADER_BLOCKS", 61},
					{"RESPOND_HEADER_BLOCKS", 62},
				},
			},
			{
				comment: "Introducer protocol (introducer <-> full_node)",
				values: []valueItem{
					{"REQUEST_PEERS_INTRODUCER", 63},
					{"RESPOND_PEERS_INTRODUCER", 64},
				},
			},
			{
				comment: "Simulator protocol",
				values: []valueItem{
					{"FARM_NEW_BLOCK", 65},
				},
			},
		},
	},
}

func structName(s string) string {
	s = strings.ToLower(s)
	s = strings.Replace(s, "_", " ", -1)
	s = strings.Title(s)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "Vdf", "VDF", -1)
	return s
}

func main() {
	for _, groupFile := range groups {
		fname := groupFile.fnamePrefix + "_generated.go"
		outFile, err := os.Create(fname)
		if err != nil {
			log.Fatal(err)
		}

		write := func(format string, a ...interface{}) {
			if _, err := fmt.Fprintf(outFile, format, a...); err != nil {
				log.Fatal(err)
			}
		}

		write("// Generated with `go genetare`. Do not edit.\n")
		write("package types\n\n")

		if len(groupFile.imports) > 0 {
			write("import (\n")
			for _, path := range groupFile.imports {
				write(`"` + path + `"` + "\n")
			}
			write(")\n\n")
		}

		write("const (\n")
		for i, group := range groupFile.groups {
			if i > 0 {
				write("\n")
			}
			if group.comment != "" {
				write("// " + group.comment + "\n")
			}
			for _, v := range group.values {
				write(groupFile.constPrefix + "_" + v.name + " = " + strconv.Itoa(v.value) + "\n")
			}
		}
		write(")\n\n")

		needStructGetters := false
		for _, group := range groupFile.groups {
			if group.useInGetter {
				needStructGetters = true
				break
			}
		}
		if needStructGetters {
			write("func " + groupFile.getterPrefix + "Struct(type_ uint8) (utils.FromToBytes, bool) {\n")
			write("switch type_ {\n")
			for _, group := range groupFile.groups {
				if group.useInGetter {
					for _, v := range group.values {
						write("case " + groupFile.constPrefix + "_" + v.name + ":\n")
						write("return &" + structName(v.name) + "{}, true\n")
					}
				}
			}
			write("default:\n")
			write("return nil, false\n")
			write("}\n")
			write("}\n\n")

			write("func " + groupFile.getterPrefix + "FromStruct(obj interface{}) (uint8, bool) {\n")
			write("switch obj.(type) {\n")
			for _, group := range groupFile.groups {
				if group.useInGetter {
					for _, v := range group.values {
						write("case " + structName(v.name) + ", *" + structName(v.name) + ":\n")
						write("return " + groupFile.constPrefix + "_" + v.name + ", true\n")
					}
				}
			}
			write("default:\n")
			write("return 0, false\n")
			write("}\n")
			write("}\n\n")
		}

		if groupFile.stringNameGetter {
			write("func " + groupFile.getterPrefix + "Name(type_ uint8) (string, bool) {\n")
			write("switch type_ {\n")
			for _, group := range groupFile.groups {
				for _, v := range group.values {
					name, ok := groupFile.stringNameMap[v.name]
					if !ok {
						name = v.name
					}
					write("case " + groupFile.constPrefix + "_" + v.name + ":\n")
					write(`return "` + name + `", true` + "\n")
				}
			}
			write("default:\n")
			write(`return "UNKNOWN", false` + "\n")
			write("}\n")
			write("}\n\n")
		}

		if err := outFile.Close(); err != nil {
			log.Fatal(err)
		}

		if err := exec.Command("go", "fmt", fname).Run(); err != nil {
			log.Fatal(err)
		}
	}
}
