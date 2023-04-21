package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"net/http"
	"runtime"
	"time"
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

type hdwalletRequest struct {
	Start     uint32 `json:"start"`
	Num       uint32 `json:"num"`
	PublicKey string `json:"publickey"`
	maxCore   uint32 `json:"maxcore"`
}

type hdwalletResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type getBalanceRequest struct {
	Address string `json:"address"`
}

type getBalanceResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Balance string `json:"balance"`
}

type hdgetBalanceRequest struct {
	Publickey string `json:"publickey"`
	Start     uint32 `json:"start"`
	Num       uint32 `json:"num"`
}

type hdBalanceResult struct {
	Address  string `json:"address"`
	Balance  string `json:"balance"`
	Duration string `json:"duration"`
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

func hdwalletGenerateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req hdwalletRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request data: %v", err), http.StatusBadRequest)
			return
		}

		startTime := time.Now() // İşlemin başlangıç zamanını kaydedin
		Generate(req.Start, req.Num, req.PublicKey)
		elapsedTime := time.Since(startTime) // Geçen süreyi hesaplayın

		minutes := int(elapsedTime.Minutes())
		seconds := int(elapsedTime.Seconds()) % 60
		milliseconds := elapsedTime.Milliseconds() % 1000

		response := hdwalletResponse{
			Status:  "success",
			Message: fmt.Sprintf("Adres ekleme süresi: %d dakika, %d saniye, %d milisaniye, toplam %d kanal ile", minutes, seconds, milliseconds, runtime.NumCPU()),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var req getBalanceRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request data: %v", err), http.StatusBadRequest)
			return
		}

		client, err := ethclient.Dial("/home/sbr/Masaüstü/node1/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

		address := common.HexToAddress(req.Address)
		balance, err := client.BalanceAt(context.Background(), address, nil)

		//çoklu sorgulama
		/*for i := 0; i < 1000000; i++ {
			balance, err := client.BalanceAt(context.Background(), address, nil)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error getting balance: %v", err), http.StatusInternalServerError)
				return
			}
			fmt.Println(balance.String())
		}*/

		response := getBalanceResponse{
			Status:  "success",
			Message: "Address balance",
			Balance: balance.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func hdgetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if r.Method == "POST" {
		var req hdgetBalanceRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request data: %v", err), http.StatusBadRequest)
			return
		}

		client, err := ethclient.Dial("/home/sbr/Masaüstü/node1/geth.ipc")
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

		extPubKeyStr := req.Publickey
		extKey, err := hdkeychain.NewKeyFromString(extPubKeyStr)
		if err != nil {
			panic(err)
		}
		startTime := time.Now()

		var results []hdBalanceResult

		basla := req.Start
		bitis := req.Num

		for i := basla; i < bitis; i++ {
			path := fmt.Sprintf("0/%d", i)

			childKey, err := DerivePath(extKey, path)
			if err != nil {
				panic(err)
			}

			rawPubKey, err := childKey.ECPubKey()
			if err != nil {
				panic(err)
			}

			ethAddress := crypto.PubkeyToAddress(*rawPubKey.ToECDSA())

			balance, err := client.BalanceAt(context.Background(), ethAddress, nil)
			if err != nil {
				http.Error(w, fmt.Sprintf("Error getting balance: %v", err), http.StatusInternalServerError)
				return
			}

			if balance.String() != "0" {
				result := hdBalanceResult{
					Address: ethAddress.String(),
					Balance: balance.String(),
				}
				results = append(results, result)
			}
		}

		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)

		minutes := int(elapsedTime.Minutes())
		seconds := int(elapsedTime.Seconds()) % 60
		milliseconds := elapsedTime.Milliseconds() % 1000

		duration := fmt.Sprintf("%d dakika, %d saniye, %d milisaniye", minutes, seconds, milliseconds)

		for i := range results {
			results[i].Duration = duration
		}

		// JSON response
		response, err := json.Marshal(results)
		if err != nil {
			http.Error(w, fmt.Sprintf("JSON conversion error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}
