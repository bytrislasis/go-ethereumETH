package stf

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ws"
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
		fmt.Println(toAddress)
		check, _ := CheckKeyExists(rdb, common.HexToAddress(toAddress))
		time.Sleep(1 * time.Second)
		fmt.Println(check)
		if check {

			message := fmt.Sprintf(toAddress)
			ws.BroadcastNewBlock(message)
			bot.SendMessage(message, nil)

		}
	}
}
