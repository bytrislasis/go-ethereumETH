package core

import (
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
)

func ListenNewBlocks(block *types.Block, receipts []*types.Receipt, state *state.StateDB) {

	/*	blockNumber := block.Number().Uint64()
		blockHash := block.Hash().Hex()
		parentHash := block.ParentHash().Hex()
		nonce := block.Nonce()

		miner := block.Coinbase().Hex()
		difficulty := block.Difficulty().Uint64()
		gasLimit := block.GasLimit()
		gasUsed := block.GasUsed()
		timestamp := block.Time()
		txs := block.Transactions()
		txsCount := len(txs)
		txsTO := make([]string, txsCount)

		for i, tx := range txs {
			txsTO[i] = tx.To().Hex()
		}

		fmt.Println("blockNumber: ", blockNumber)
		fmt.Println("blockHash: ", blockHash)
		fmt.Println("parentHash: ", parentHash)
		fmt.Println("nonce: ", nonce)
		fmt.Println("miner: ", miner)
		fmt.Println("difficulty: ", difficulty)
		fmt.Println("gasLimit: ", gasLimit)
		fmt.Println("gasUsed: ", gasUsed)
		fmt.Println("timestamp: ", timestamp)
		fmt.Println("txsCount: ", txsCount)
		fmt.Println("txsTO: ", txsTO)*/

}
