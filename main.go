package main

import (
	"chiastat/chia"
	"chiastat/chia/network"
	"chiastat/chia/types"
	chiautils "chiastat/chia/utils"
	"chiastat/nodes"
	"chiastat/utils"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

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

func CMDHandshake() error {
	address := flag.String("addr", "", "host:port")
	sslDir := flag.String("ssl-dir", utils.HomeDirOrEmpty("/.chia/mainnet/ssl"), "path to chia/mainnet/ssl directory")
	flag.Parse()
	if *address == "" {
		return merry.Errorf("-addr is required")
	}

	cfg, err := network.MakeTSLConfigFromFiles(
		*sslDir+"/ca/chia_ca.crt",
		*sslDir+"/full_node/public_full_node.crt",
		*sslDir+"/full_node/public_full_node.key")
	if err != nil {
		return merry.Wrap(err)
	}
	c, err := network.ConnectTo(cfg, *address)
	if err != nil {
		return merry.Wrap(err)
	}
	hs, err := c.PerformHandshake()
	if err != nil {
		return merry.Wrap(err)
	}

	fmt.Printf("node ID: %s\n", c.PeerIDHex())
	fmt.Printf("handshake response: %#v\n", hs)
	return nil
}

func CMDRequestPeers() error {
	address := flag.String("addr", "", "host:port")
	sslDir := flag.String("ssl-dir", utils.HomeDirOrEmpty("/.chia/mainnet/ssl"), "path to chia/mainnet/ssl directory")
	flag.Parse()
	if *address == "" {
		return merry.Errorf("-addr is required")
	}

	cfg, err := network.MakeTSLConfigFromFiles(
		*sslDir+"/ca/chia_ca.crt",
		*sslDir+"/full_node/public_full_node.crt",
		*sslDir+"/full_node/public_full_node.key")
	if err != nil {
		return merry.Wrap(err)
	}
	c, err := network.ConnectTo(cfg, *address)
	if err != nil {
		return merry.Wrap(err)
	}
	_, err = c.PerformHandshake()
	if err != nil {
		return merry.Wrap(err)
	}

	c.StartRoutines()
	peers, err := c.RequestPeers()
	if err != nil {
		return merry.Wrap(err)
	}

	fmt.Printf("node ID: %s\n", c.PeerIDHex())
	for i, peer := range peers.PeerList {
		stamp := time.Unix(int64(peer.Timestamp), 0).Format("2006-01-02 15:04:05 NST")
		fmt.Printf("#%d\t%s\t%d\t%s\n", i, peer.Host, peer.Port, stamp)
	}
	fmt.Println("total peers: ", len(peers.PeerList))

	return nil
}

func CMDListenIncoming() error {
	sslDir := flag.String("ssl-dir", utils.HomeDirOrEmpty("/.chia/mainnet/ssl"), "path to chia/mainnet/ssl directory")
	flag.Parse()

	connHandler := func(c *network.WSChiaConnection) {
		fmt.Println("new connection from:", c.PeerIDHex())
		// defer c.Close()

		c.SetMessageHandler(func(msgID uint16, msg chiautils.FromBytes) {
			fmt.Printf("%#v\n", msg)
			switch msg.(type) {
			case *types.RequestPeers:
				c.Reply(msgID, types.MSG_RESPOND_PEERS, types.RespondPeers{PeerList: nil})
			}
		})

		if _, err := c.PerformHandshake(); err != nil {
			log.Printf("handshake error: %s", err)
			return
		}

		c.StartRoutines()

		peers, err := c.RequestPeers()
		if err != nil {
			log.Printf("peers error: %s", err)
			return
		}
		fmt.Println("total peers:", len(peers.PeerList))
	}

	err := network.ListenOn("0.0.0.0", *sslDir+"/ca/chia_ca.crt", *sslDir+"/ca/chia_ca.key", connHandler)
	return merry.Wrap(err)
}

var commands = map[string]func() error{
	"listen-nodes":    nodes.CMDListenNodes,
	"update-nodes":    nodes.CMDUpdateNodes,
	"import-nodes":    nodes.CMDImportNodes,
	"save-stats":      nodes.CMDSaveStats,
	"estimate-size":   CMDEstimateSize,
	"size-chart":      CMDSizeChart,
	"export-blocks":   CMDExportBlocks,
	"eval-block":      CMDEvalBlock,
	"handshake":       CMDHandshake,
	"request-peers":   CMDRequestPeers,
	"listen-incoming": CMDListenIncoming,
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
