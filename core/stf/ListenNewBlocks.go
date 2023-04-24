package stf

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redis/redis/v8"
)

func ListenNewBlocks(block *types.Block) {

	/*
		bot := NewTelegramBot("6191705778:AAH2aExyb-bJelRT_B8f-tMBoIYSKkEGBuU", "-1001927709952")
		bot.SendMessage("New Block Detected", nil)
	*/

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis sunucusu adresi
		Password: "",               // Redis şifresi, eğer varsa
		DB:       0,                // Seçilecek Redis veritabanı
	})

	/*blockNumber := block.Number().Uint64()
	blockHash := block.Hash().Hex()
	parentHash := block.ParentHash().Hex()
	nonce := block.Nonce()

	miner := block.Coinbase().Hex()
	difficulty := block.Difficulty().Uint64()
	gasLimit := block.GasLimit()
	gasUsed := block.GasUsed()
	timestamp := block.Time()*/
	txs := block.Transactions()
	txsCount := len(txs)
	txsTO := make([]string, txsCount)

	for i, tx := range txs {
		txsTO[i] = tx.To().Hex()
		CheckKeyExists(rdb, common.HexToAddress(txsTO[i]))
		bot := NewTelegramBot("6191705778:AAH2aExyb-bJelRT_B8f-tMBoIYSKkEGBuU", "-1001927709952")
		bot.SendMessage(tx.To().Hex(), nil)
	}

	fmt.Println(txsTO)

	/*fmt.Println("blockNumber: ", blockNumber)
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
