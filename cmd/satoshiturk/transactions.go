// cmd/satoshiturk/transactions.go

package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/ethclient"
)

type AddressListResponse struct {
	Addresses []string `json:"addresses"`
}

func getAllTransactionAddressesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	client, err := ethclient.Dial(IPCPATH)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	addresses, err := getAllTransactionAddresses(client)
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

func getAllTransactionAddresses(client *ethclient.Client) ([]string, error) {
	addressMap := make(map[string]struct{})
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	for blockNumber := header.Number; blockNumber.Sign() > 0; blockNumber = new(big.Int).Sub(blockNumber, big.NewInt(1)) {
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

	addresses := make([]string, 0, len(addressMap))
	for address := range addressMap {
		addresses = append(addresses, address)
	}

	return addresses, nil

}
