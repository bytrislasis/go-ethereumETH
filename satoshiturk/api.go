package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"net/http"
)

type blockRequest struct {
	BlockNumber string `json:"blocknumber"`
}

type txResponse struct {
	Hash     string `json:"hash"`
	Nonce    uint64 `json:"nonce"`
	Gas      uint64 `json:"gas"`
	GasPrice string `json:"gasPrice"`
	To       string `json:"to"`
	Value    string `json:"value"`
	Data     []byte `json:"data"`
	Pending  bool   `json:"pending"`
}

type blockResponse struct {
	Number       string      `json:"number"`
	Hash         string      `json:"hash"`
	ParentHash   string      `json:"parentHash"`
	Nonce        uint64      `json:"nonce"`
	Miner        string      `json:"miner"`
	Difficulty   string      `json:"difficulty"`
	GasLimit     uint64      `json:"gasLimit"`
	GasUsed      uint64      `json:"gasUsed"`
	Timestamp    uint64      `json:"timestamp"`
	Transactions interface{} `json:"transactions"`
}

func txDetay(w http.ResponseWriter, r *http.Request) {
	fmt.Print("--------------------------------------------------------------------")
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial("/home/sbr/Masaüstü/node1/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	txHash := common.HexToHash(data["txid"])

	tx, pending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	response := txResponse{
		Hash:     tx.Hash().Hex(),
		Nonce:    tx.Nonce(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice().String(),
		To:       tx.To().Hex(),
		Value:    tx.Value().String(),
		Data:     tx.Data(),
		Pending:  pending,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func blockInfo(w http.ResponseWriter, r *http.Request) {
	var req blockRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial("/home/sbr/Masaüstü/node1/geth.ipc")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	blockNumber := new(big.Int)
	blockNumber.SetString(req.BlockNumber, 10)

	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}

	response := blockResponse{
		Number:       block.Number().String(),
		Hash:         block.Hash().Hex(),
		ParentHash:   block.ParentHash().Hex(),
		Nonce:        block.Nonce(),
		Miner:        block.Coinbase().Hex(),
		Difficulty:   block.Difficulty().String(),
		GasLimit:     block.GasLimit(),
		GasUsed:      block.GasUsed(),
		Timestamp:    block.Time(),
		Transactions: block.Transactions(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
