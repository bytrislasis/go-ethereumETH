package stf

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/go-redis/redis/v8"
	"time"
)

func ListenNewBlocks(block *types.Block) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis sunucusu adresi
		Password: "",               // Redis şifresi, eğer varsa
		DB:       0,                // Seçilecek Redis veritabanı
	})

	txs := block.Transactions()

	bot := NewTelegramBot("6191705778:AAH2aExyb-bJelRT_B8f-tMBoIYSKkEGBuU", "-1001927709952")

	for _, tx := range txs {
		toAddress := tx.To().Hex()
		check, _ := CheckKeyExists(rdb, common.HexToAddress(toAddress))
		if check != false {
			txHash := tx.Hash().Hex()

			value := tx.Value()
			gas := tx.Gas()
			gasPrice := tx.GasPrice()
			nonce := tx.Nonce()

			message := fmt.Sprintf("Transaction Detected:\n %s\nTo: %s\nTx Hash: %s\nValue: %s ETH\nGas: %d\nGas Price: %s GWei\nNonce: %d",
				toAddress, txHash, value.String(), gas, gasPrice.String(), nonce)

			bot.SendMessage(message, nil)

			//sleep 3 seconds
			time.Sleep(1 * time.Second)
		}
	}
}
