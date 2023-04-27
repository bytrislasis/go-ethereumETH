package stf

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ws"
	"github.com/go-redis/redis/v8"
	"math"
	"math/big"
	"time"
)

type transactionInfo struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Nonce uint64 `json:"nonce"`
	Hash  string `json:"hash"`
}

func formatAddress(address string, maxLength int) string {
	if len(address) > maxLength {
		return address[:6] + "....." + address[len(address)-6:]
	}
	return address
}

func weiToEther(weiValue *big.Int) string {
	etherValue := new(big.Float).SetInt(weiValue)
	etherValue = etherValue.Quo(etherValue, big.NewFloat(math.Pow10(18)))
	etherStr := fmt.Sprintf("%.18f", etherValue) // Ether değerini string olarak saklayın
	return etherStr
}

func ListenNewBlocks(client *ethclient.Client, block *types.Block) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis sunucusu adresi
		Password: "",               // Redis şifresi, eğer varsa
		DB:       0,                // Seçilecek Redis veritabanı
	})

	txs := block.Transactions()

	bot := NewTelegramBot("6191705778:AAH2aExyb-bJelRT_B8f-tMBoIYSKkEGBuU", "-1001927709952")

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return
	}

	signer := types.NewEIP155Signer(chainID)

	for _, tx := range txs {
		from, _ := types.Sender(signer, tx)
		toAddress := tx.To().Hex()
		fmt.Println(toAddress)
		check, _ := CheckKeyExists(rdb, common.HexToAddress(toAddress))
		time.Sleep(1 * time.Second)
		fmt.Println(check)
		if check {
			etherValue := weiToEther(tx.Value())
			shortFrom := formatAddress(from.Hex(), 10)
			shortTo := formatAddress(toAddress, 10)
			shortHash := formatAddress(tx.Hash().Hex(), 10)

			txInfo := transactionInfo{
				From: from.Hex(),
				To:   toAddress,
				//Value: tx.Value().String(),
				Value: etherValue,
				Nonce: tx.Nonce(),
				Hash:  tx.Hash().Hex(),
			}

			txInfoJSON, err := json.Marshal(txInfo)
			if err != nil {
				fmt.Println("JSON marshaling error:", err)
				continue
			}

			message := string(txInfoJSON)
			ws.BroadcastNewBlock(message)

			// Veriyi tablo şeklinde düzenle
			tableMessage := fmt.Sprintf("From: %s\nTo: %s\nValue: %s\nNonce: %d\nHash: %s", shortFrom, shortTo, weiToEther(tx.Value()), tx.Nonce(), shortHash)
			bot.SendMessage(tableMessage, nil)
		}
	}
}
