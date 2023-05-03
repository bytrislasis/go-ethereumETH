package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

//apiden gelenleri cevaplar

type AddressRangeRequest struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type AddressBalance struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
}

type AddressListResponse struct {
	Addresses []AddressBalance `json:"addresses"`
}

func getAllTransactionAddressesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddressRangeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(IPCPATH)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	addresses, err := getAllTransactionAddresses(client, req.Start, req.End)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching addresses: %v", err), http.StatusInternalServerError)
		return
	}

	response := AddressListResponse{
		Addresses: addresses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getAllTransactionAddresses(client *ethclient.Client, startBlock string, endBlock string) ([]AddressBalance, error) {
	addressMap := make(map[string]struct{})
	balances := make([]AddressBalance, 0)
	start, ok := new(big.Int).SetString(startBlock, 10)
	if !ok {
		return nil, fmt.Errorf("Invalid start block number")
	}
	end, ok := new(big.Int).SetString(endBlock, 10)
	if !ok {
		return nil, fmt.Errorf("Invalid end block number")
	}

	for blockNumber := new(big.Int).Set(end); blockNumber.Cmp(start) >= 0; blockNumber.Sub(blockNumber, big.NewInt(1)) {
		block, err := client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions() {
			to := tx.To()
			if to != nil {
				addressMap[to.Hex()] = struct{}{}
			}
		}
	}

	oneEthInWei := big.NewInt(1e18)
	for addr := range addressMap {
		address := common.HexToAddress(addr)
		balance, err := client.BalanceAt(context.Background(), address, nil)
		if err != nil {
			return nil, err
		}

		if balance.Cmp(oneEthInWei) > 0 {
			balanceInEth := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(oneEthInWei))
			balances = append(balances, AddressBalance{
				Address: addr,
				Balance: balanceInEth.String(),
			})
		}
	}

	return balances, nil
}
