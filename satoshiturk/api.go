package main

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"net/http"
)

func txSorgula(w http.ResponseWriter, r *http.Request) {
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

	txHash := common.HexToHash(data["sorgula"])

	tx, pending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"hash":     tx.Hash().Hex(),
		"nonce":    tx.Nonce(),
		"gas":      tx.Gas(),
		"gasPrice": tx.GasPrice().String(),
		"to":       tx.To().Hex(),
		"value":    tx.Value().String(),
		"data":     tx.Data(),
		"pending":  pending,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
