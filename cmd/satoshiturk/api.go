package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
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
	Index    uint32 `json:"index"`
}

type sendRandomEthRequest struct {
	PrivateKey string `json:"privatekey"`
	StartIndex uint32 `json:"startindex"`
	EndIndex   uint32 `json:"endindex"`
	PublicKey  string `json:"publickey"`
}

type sendRandomEthResponse struct {
	TxHashes []string `json:"tx_hashes"`
}

type blockScannerRequest struct {
	StartIndex uint32 `json:"start"`
	EndIndex   uint32 `json:"end"`
	Xpub       string `json:"publickey"`
	HdStart    uint32 `json:"hdstart"`
	HdEnd      uint32 `json:"hdend"`
}

type FoundAddressInfo struct {
	Address     string `json:"address"`
	BlockNumber string `json:"blockNumber"`
	Balance     string `json:"balance"`
}

func txDetay(w http.ResponseWriter, r *http.Request) {
	fmt.Print("--------------------------------------------------------------------")
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	client, err := ethclient.Dial(getIpcPath)
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

	client, err := ethclient.Dial(getIpcPath)
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	if r.Method == "POST" {
		var req getBalanceRequest

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request data: %v", err), http.StatusBadRequest)
			return
		}

		client, err := ethclient.Dial(getIpcPath)
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}

		address := common.HexToAddress(req.Address)

		numQueries := 1000
		resultChan := make(chan *big.Int, numQueries)
		var wg sync.WaitGroup
		wg.Add(numQueries)

		for i := 0; i < numQueries; i++ {
			go func() {
				defer wg.Done()
				balance, err := client.BalanceAt(context.Background(), address, nil)
				if err != nil {
					log.Printf("Error getting balance: %v", err)
					resultChan <- big.NewInt(0)
				} else {
					resultChan <- balance
				}
			}()
		}

		wg.Wait()
		close(resultChan)

		var totalBalance big.Int
		for balance := range resultChan {
			totalBalance.Add(&totalBalance, balance)
		}

		response := getBalanceResponse{
			Status:  "success",
			Message: "Address balance",
			Balance: totalBalance.String(),
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

		client, err := ethclient.Dial(getIpcPath)
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
				// Adresin bulunma süresini hesapla
				foundTime := time.Now()
				elapsedTime := foundTime.Sub(startTime)
				minutes := int(elapsedTime.Minutes())
				seconds := int(elapsedTime.Seconds()) % 60
				milliseconds := elapsedTime.Milliseconds() % 1000
				duration := fmt.Sprintf("%d dakika, %d saniye, %d milisaniye", minutes, seconds, milliseconds)

				// hdBalanceResult yapısına bulunma süresini ekle
				result := hdBalanceResult{
					Address:  ethAddress.String(),
					Balance:  balance.String(),
					Index:    i,
					Duration: duration,
				}
				results = append(results, result)
			}
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

func sendRandomEthHandler(w http.ResponseWriter, r *http.Request) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if r.Method == "POST" {
		var req sendRandomEthRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request data: %v", err), http.StatusBadRequest)
			return
		}

		// IPC bağlantısı
		client, err := ethclient.Dial(getIpcPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("IPC bağlantısı başarısız: %v", err), http.StatusInternalServerError)
			return
		}

		// Private key'i ecdsa.PrivateKey'e dönüştür
		privateKey, err := crypto.HexToECDSA(req.PrivateKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("private key dönüşümü başarısız: %v", err), http.StatusInternalServerError)
			return
		}

		// Gönderici adresini elde et
		fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

		// Adresleri oluştur
		xpub := req.PublicKey
		start := req.StartIndex
		end := req.EndIndex
		addresses, err := addrgenerate(xpub, start, end)
		if err != nil {
			http.Error(w, fmt.Sprintf("Address generation error: %v", err), http.StatusInternalServerError)
			return
		}

		// Nonce al
		nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			http.Error(w, fmt.Sprintf("nonce alma başarısız: %v", err), http.StatusInternalServerError)
			return
		}

		// Chain ID al
		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			http.Error(w, fmt.Sprintf("chain id alma başarısız: %v", err), http.StatusInternalServerError)
		}

		var txHashes []string
		for _, address := range addresses {
			// Rastgele bir değerde ETH miktarı (0.1 ETH'den az) oluştur
			rand.Seed(time.Now().UnixNano())
			ethValue := big.NewInt(rand.Int63n(100000000000000000)) // 0.1 ETH'den az rastgele değer

			// Transfer işlemi için gas limit ve gas fiyatı belirle
			gasLimit := uint64(21000)
			gasPrice, err := client.SuggestGasPrice(context.Background())
			if err != nil {
				http.Error(w, fmt.Sprintf("gas fiyatı önerisi başarısız: %v", err), http.StatusInternalServerError)
				return
			}

			// İşlem yapısını oluştur
			toAddress := common.HexToAddress(address)
			tx := types.NewTransaction(nonce, toAddress, ethValue, gasLimit, gasPrice, nil)

			// İşlemi imzala
			signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("işlem imzalama başarısız: %v", err), http.StatusInternalServerError)
				return
			}

			// İşlemi gönder
			err = client.SendTransaction(context.Background(), signedTx)
			if err != nil {
				http.Error(w, fmt.Sprintf("işlem gönderimi başarısız: %v", err), http.StatusInternalServerError)
				return
			}

			txHash := signedTx.Hash().Hex()
			fmt.Printf("İşlem gönderildi: %s\n", txHash)
			txHashes = append(txHashes, txHash)

			// Nonce değerini güncelle
			nonce++
		}

		response := sendRandomEthResponse{
			TxHashes: txHashes,
		}

		jsonResponse, err := json.Marshal(response)
		if err != nil {
			http.Error(w, fmt.Sprintf("JSON conversion error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func deriveAddress(extKey *hdkeychain.ExtendedKey, i uint32, addresses chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	path := fmt.Sprintf("0/%d", i)
	childKey, err := DerivePath(extKey, path)
	if err != nil {
		return
	}

	rawPubKey, err := childKey.ECPubKey()
	if err != nil {
		return
	}

	ethAddress := crypto.PubkeyToAddress(*rawPubKey.ToECDSA()).Hex()
	addresses <- ethAddress
}

func addrgenerate(xpub string, startIndex, endIndex uint32) ([]string, error) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	extPubKeyStr := xpub
	extKey, err := hdkeychain.NewKeyFromString(extPubKeyStr)
	if err != nil {
		return nil, err
	}

	basla := startIndex
	bitis := endIndex

	addresses := make(chan string, bitis-basla)
	var wg sync.WaitGroup

	for i := basla; i < bitis; i++ {
		wg.Add(1)
		go deriveAddress(extKey, i, addresses, &wg)
	}

	wg.Wait()
	close(addresses)

	addressList := []string{}
	for addr := range addresses {
		addressList = append(addressList, addr)
	}

	return addressList, nil
}

func blockScannerHandler(w http.ResponseWriter, r *http.Request) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var req blockScannerRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	client, err := ethclient.Dial(getIpcPath)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	basla := req.StartIndex
	bitis := req.EndIndex
	xpub := req.Xpub

	addresses, err := addrgenerate(xpub, req.HdStart, req.HdEnd)
	if err != nil {
		http.Error(w, "Error generating addresses", http.StatusInternalServerError)
		return
	}

	foundAddresses := []FoundAddressInfo{}

	for i := basla; i <= bitis; i++ {
		blockNumber := big.NewInt(int64(i))
		block, err := client.BlockByNumber(context.Background(), blockNumber)
		if err != nil {
			http.Error(w, "Block not found", http.StatusNotFound)
			return
		}

		for _, tx := range block.Transactions() {
			from, err := client.TransactionSender(context.Background(), tx, block.Hash(), 0)
			if err != nil {
				http.Error(w, "Error getting transaction sender", http.StatusInternalServerError)
				return
			}

			to := tx.To()

			for _, addr := range addresses {
				targetAddress := common.HexToAddress(addr)
				if from == targetAddress || (to != nil && *to == targetAddress) {
					balance, err := client.BalanceAt(context.Background(), targetAddress, nil)
					if err != nil {
						http.Error(w, "Error getting address balance", http.StatusInternalServerError)
						return
					}

					foundAddresses = append(foundAddresses, FoundAddressInfo{
						Address:     addr,
						BlockNumber: blockNumber.String(),
						Balance:     balance.String(),
					})
					break
				}
			}
		}
	}

	elapsed := time.Since(startTime)
	elapsedStr := fmt.Sprintf("%d minutes %d seconds %d milliseconds", int(elapsed.Minutes()), int(elapsed.Seconds())%60, int(elapsed.Milliseconds())%1000)

	response := struct {
		Status      string
		Message     string
		ElapsedTime string
		Data        []FoundAddressInfo
	}{
		Status:      "success",
		Message:     "Arama tamamlandı.",
		ElapsedTime: elapsedStr,
		Data:        foundAddresses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//telegram mesaj atma
/*bot := NewTelegramBot("6191705778:AAH2aExyb-bJelRT_B8f-tMBoIYSKkEGBuU", "-1001927709952")

response, err = bot.SendMessage("text", nil)
if err != nil {
log.Printf("Error sending message: %v", err)
} else {
log.Printf("Message sent: %v", response)

}*/
