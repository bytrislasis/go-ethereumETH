package stf

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"time"
)

const requiredConfirmations = 3

func Dinle() {
	fmt.Println("Dinleme başladı.")
	time.Sleep(5 * time.Second)
	client, err := rpc.Dial("/home/metatime/Masaüstü/node1/geth.ipc") // Geth IPC dosyasının yolunu değiştirin.
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	ethClient := ethclient.NewClient(client)

	headers := make(chan *types.Header)
	sub, err := client.EthSubscribe(context.Background(), headers, "newHeads")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := ethClient.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			latestBlock, err := ethClient.BlockByNumber(context.Background(), nil)
			if err != nil {
				log.Fatal(err)
			}

			confirmations := new(big.Int).Sub(latestBlock.Number(), block.Number())
			if confirmations.Cmp(big.NewInt(requiredConfirmations)) >= 0 {
				fmt.Println("Block", block.Number().String(), "has enough confirmations.")

				ListenNewBlocks(ethClient, block)
			}
		}
	}
}
