package main

import (
	"database/sql"
	"fmt"
	"math/big"
	"os"

	"github.com/ansel1/merry"
	_ "github.com/mattn/go-sqlite3"
)

type Scanner interface {
	Scan(dest ...interface{}) error
}

const BlockRecordSelectCols = "header_hash, prev_hash, height, block, sub_epoch_summary, is_peak, is_block"

func BlockRecordFromRow(rows Scanner) (*BlockRecord, error) {
	var headerHash string
	var prevHash string
	var height int64
	var block []byte
	var subEpochSummary []byte
	var isPeak int8
	var isBlock int8
	err := rows.Scan(&headerHash, &prevHash, &height, &block, &subEpochSummary, &isPeak, &isBlock)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	// fmt.Println(headerHash, prevHash, height, isPeak, isBlock)
	// fmt.Println(block)
	buf := NewParseBuf(block)
	br := BlockRecordFromBytes(buf)
	if buf.err != nil {
		return nil, buf.err
	}
	// fmt.Printf("%#v\n", br)
	// fmt.Println(br.height, br.weight, br.totalIters)
	return &br, nil
}

func BlockRecordByHeight(db *sql.DB, height int64) (*BlockRecord, error) {
	row := db.QueryRow("SELECT "+BlockRecordSelectCols+" FROM block_records WHERE height = ?", height)
	return BlockRecordFromRow(row)
}

func EstimateNetworkSpace(db *sql.DB, lastHeight, pastOffset int64) (*big.Int, error) {
	firstHeight := lastHeight - pastOffset
	if firstHeight < 0 {
		firstHeight = 0
	}

	br0, err := BlockRecordByHeight(db, firstHeight)
	if err != nil {
		return nil, merry.Wrap(err)
	}
	br1, err := BlockRecordByHeight(db, lastHeight)
	if err != nil {
		return nil, merry.Wrap(err)
	}

	// https://github.com/Chia-Network/chia-blockchain/blob/latest/chia/rpc/full_node_rpc_api.py#L276
	deltaWeight := (&big.Int{}).Sub(br1.Weight, br0.Weight)
	deltaIters := (&big.Int{}).Sub(br1.TotalIters, br0.TotalIters)
	additionalDifficultyConstant := (&big.Int{}).Exp(big.NewInt(2), big.NewInt(67), nil)
	eligiblePlotsFilterMultiplier := (&big.Int{}).Exp(big.NewInt(2), big.NewInt(9), nil)
	UI_ACTUAL_SPACE_CONSTANT_FACTOR := 0.762

	spaceEstimate := &big.Int{}
	spaceEstimate.Mul(additionalDifficultyConstant, eligiblePlotsFilterMultiplier)
	spaceEstimate.Mul(spaceEstimate, deltaWeight)
	spaceEstimate.Div(spaceEstimate, deltaIters)
	spaceEstimate.Mul(spaceEstimate, big.NewInt(int64(UI_ACTUAL_SPACE_CONSTANT_FACTOR*1000000)))
	spaceEstimate.Div(spaceEstimate, big.NewInt(1000000))
	return spaceEstimate, nil
}

func mainErr() error {
	db, err := sql.Open("sqlite3", "blockchain_v1_mainnet.sqlite")
	if err != nil {
		return merry.Wrap(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT header_hash, prev_hash, height, block, sub_epoch_summary, is_peak, is_block FROM block_records LIMIT 10")
	if err != nil {
		return merry.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		br, err := BlockRecordFromRow(rows)
		if err != nil {
			return merry.Wrap(err)
		}
		fmt.Println(br.HeaderHash, br.Height, br.Weight, br.TotalIters)
	}
	err = rows.Err()
	if err != nil {
		return merry.Wrap(err)
	}

	spaceEstimate, err := EstimateNetworkSpace(db, 280000, 4608)
	fmt.Println(spaceEstimate, (&big.Int{}).Div(spaceEstimate, big.NewInt(1024*1024*1024*1024*1024)))

	return nil
}

func main() {
	if err := mainErr(); err != nil {
		println(merry.Details(err))
		os.Exit(1)
	}
}
