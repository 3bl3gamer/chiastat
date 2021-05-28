package main

import (
	"chiastat/chia"
	"chiastat/nodes"
	"chiastat/utils"
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

	db, err := utils.OpenExistingSqlite3(*dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	spaceEstimate, err := chia.EstimateNetworkSpaceFromDB(db, -1, 4608)
	if err != nil {
		return merry.Wrap(err)
	}
	fmt.Printf("%dB %dPiB\n", spaceEstimate, (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)))
	return nil
}

func CMDSizeChart() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"blockchain_v1_mainnet.sqlite", "path to blockchain_v1_mainnet.sqlite")
	flag.Parse()

	db, err := utils.OpenExistingSqlite3(*dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	if err := chia.PrintNetworkSpaceChartFromDB(db); err != nil {
		return merry.Wrap(err)
	}
	return nil
}

func CMDExportBlocks() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"blockchain_v1_mainnet.sqlite", "path to blockchain_v1_mainnet.sqlite")
	tableName := flag.String("table", "full_blocks", `table name, "full_blocks" or "block_records"`)
	fname := flag.String("fname", "", "out file name (<table>.raw by default)")
	flag.Parse()

	if *fname == "" {
		*fname = *tableName + ".raw"
	}

	db, err := utils.OpenExistingSqlite3(*dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	f, err := os.Create(*fname)
	if err != nil {
		return merry.Wrap(err)
	}
	defer f.Close()

	if err := chia.ExportBlocksData(db, *tableName, f); err != nil {
		return merry.Wrap(err)
	}

	return merry.Wrap(f.Close())
}

func CMDEvalBlock() error {
	dbPath := flag.String("db-path", utils.HomeDirOrEmpty("/.chia/mainnet/db/")+"blockchain_v1_mainnet.sqlite", "path to blockchain_v1_mainnet.sqlite")
	height := flag.Int("height", 225698, "block height (225698 is the first block with non-empty transaction generator, 225703 is the next one)")
	flag.Parse()

	db, err := utils.OpenExistingSqlite3(*dbPath)
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	return merry.Wrap(chia.EvalFullBlockFromDB(db, uint32(*height)))
}

var commands = map[string]func() error{
	"listen-nodes":  nodes.CMDListenNodes,
	"update-nodes":  nodes.CMDUpdateNodes,
	"import-nodes":  nodes.CMDImportNodes,
	"save-stats":    nodes.CMDSaveStats,
	"estimate-size": CMDEstimateSize,
	"size-chart":    CMDSizeChart,
	"export-blocks": CMDExportBlocks,
	"eval-block":    CMDEvalBlock,
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
