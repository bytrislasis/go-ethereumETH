package satoshiturk

import (
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"net/http"
)

func startServer() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/api/tx", apiHandler)
	http.HandleFunc("/api/block", apiHandler)

	http.ListenAndServe(":1983", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	//disable cors
	/*	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1983")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")*/
	/*fmt.Println("method:", r.Method+" "+r.URL.Path)*/

}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		switch r.URL.Path {
		case "/api/tx":
			txDetay(w, r)
		case "/api/block": // Yeni rota
			blockInfo(w, r)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func init() {
	stack, err := node.New(&node.Config{})
	if err != nil {
		panic(fmt.Errorf("Failed to create Ethereum node: %v", err))
	}

	ethConf := ethconfig.Defaults
	ethereum, err := eth.New(stack, &ethConf)
	if err != nil {
		panic(fmt.Errorf("Failed to create Ethereum instance: %v", err))
	}

	if ethereum == nil {
		panic("Ethereum instance is nil")
	}

	go startServer()
}
