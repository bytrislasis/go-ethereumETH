package satoshiturk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

func startServer() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":1983", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	//disable cors
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1983")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	fmt.Println("method:", r.Method+" "+r.URL.Path)

	if r.Method == "GET" {
		if r.URL.Path == "/" {
			t, err := template.ParseFiles(filepath.Join("/home/metatime/goethereum/satoshiturk/public", "index.html"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t.Execute(w, nil)
		}
	}

	if r.Method == "POST" {
		{

			if r.URL.Path == "api/tx" {

				r.ParseForm()
				client, err := ethclient.Dial("/home/metatime/Masaüstü/node1/geth.ipc")
				if err != nil {
					log.Fatalf("Failed to connect to the Ethereum client: %v", err)
				}

				txHash := common.HexToHash(r.FormValue("sorgula"))

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
		}
	}

}

func init() {
	go startServer()
}
