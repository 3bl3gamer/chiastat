package main

import (
	"chiastat/nodes"
	"chiastat/utils"
	"database/sql"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ansel1/merry"
	_ "github.com/mattn/go-sqlite3"
)

func CMDEstimateSize() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"blockchain_v1_mainnet.sqlite", "path to blockchain_v1_mainnet.sqlite")
	flag.Parse()

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		return merry.Errorf("not found: %s", *dbPath)
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	spaceEstimate, err := EstimateNetworkSpace(db, 280000, 4608)
	if err != nil {
		return merry.Wrap(err)
	}
	fmt.Printf("%dB %dPiB\n", spaceEstimate, (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)))
	return nil
}

var commands = map[string]func() error{
	"listen-nodes":  nodes.CMDListenNodes,
	"update-nodes":  nodes.CMDUpdateNodes,
	"import-nodes":  nodes.CMDImportNodes,
	"save-stats":    nodes.CMDSaveStats,
	"estimate-size": CMDEstimateSize,
}

func printUsage() {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	fmt.Printf("usage: %s [%s]\n", os.Args[0], strings.Join(names, "|"))
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printUsage()
		os.Exit(2)
	}
	cmd, ok := commands[os.Args[1]]
	if !ok {
		printUsage()
		os.Exit(1)
	}

	os.Args = os.Args[1:]
	if err := cmd(); err != nil {
		println(merry.Details(err))
		os.Exit(1)
	}
}
