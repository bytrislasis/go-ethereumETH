package satoshiturk

import (
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/gorilla/mux"
	"net/http"
)

var ethInstance *eth.Ethereum

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.HandleFunc("/api/tx", txDetay).Methods("POST")
	router.HandleFunc("/api/block", blockInfo).Methods("POST")
	router.HandleFunc("/api/hdwallet", hdwalletGenerateHandler).Methods("POST")
	router.HandleFunc("/api/getbalance", getBalanceHandler).Methods("POST")

	http.ListenAndServe(":1983", router)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		switch r.URL.Path {
		case "/api/tx": //txid sorgulama
			txDetay(w, r)
			/*{
			    "txid":"0x6c902242161e63ccebe3a6d3ad2ebaf1b8a06c86ffc8ddc6cce7d0b6dff0cc37"
			}*/
		case "/api/block": // block sorgulama
			blockInfo(w, r)
			/*{
			    "blocknumer":"5"
			}*/
		case "/api/hdwallet": // hd wallet oluşturup redise yazar
			hdwalletGenerateHandler(w, r)
			/*{
			  "start": 0,
			  "num": 100000,
			  "publickey": "xpub6D4EL9ZAG8Vf9dYXsEeXh3B4K9FYG5BL7j31drLYzYssVfASuXSAvdSHNKxmGVoPDGhJdCKZ8JU4Q8KaF52zknrCcFrfmXoUfrW8ZYGTPw4",
			  "maxcore":100
			}*/

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
	ethInstance, err = eth.New(stack, &ethConf)
	if err != nil {
		panic(fmt.Errorf("Failed to create Ethereum instance: %v", err))
	}

	if ethInstance == nil {
		panic("Ethereum instance is nil")
	}

	go startServer()
}
