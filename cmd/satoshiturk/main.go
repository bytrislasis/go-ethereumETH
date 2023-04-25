package satoshiturk

import (
	"encoding/base64"
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

var IPCPATH = "/home/metatime/Masaüstü/node1/geth.ipc"

var ethInstance *eth.Ethereum

const (
	username = "sabri"
	password = "MetaTime"
)

func startServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", handler)
	router.HandleFunc("/api/tx", basicAuthMiddleware(txDetay)).Methods("POST")
	router.HandleFunc("/api/block", basicAuthMiddleware(blockInfo)).Methods("POST")
	router.HandleFunc("/api/hdwallet", basicAuthMiddleware(hdwalletGenerateHandler)).Methods("POST")
	router.HandleFunc("/api/getbalance", basicAuthMiddleware(getBalanceHandler)).Methods("POST")
	router.HandleFunc("/api/hdgetbalance", basicAuthMiddleware(hdgetBalanceHandler)).Methods("POST")
	router.HandleFunc("/api/randomethsender", basicAuthMiddleware(sendRandomEthHandler)).Methods("POST")
	router.HandleFunc("/api/blockscanner", basicAuthMiddleware(blockScannerHandler)).Methods("POST")
	router.HandleFunc("/api/alltransactions", basicAuthMiddleware(getAllTransactionAddressesHandler)).Methods("POST")

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
		case "/api/getbalance": // hd wallet oluşturup redise yazar
			getBalanceHandler(w, r)
		/*{
		    "address": "0x9849943a82AFA29EcFEC61e80AfdfE7EA4357a33"
		}*/
		case "/api/hdgetbalance": // hd wallet oluşturup redise yazar
			getBalanceHandler(w, r)
		/*{
		    "address": "0x9849943a82AFA29EcFEC61e80AfdfE7EA4357a33"
		}*/
		case "/api/randomethsender": // hd wallet oluşturup redise yazar
			sendRandomEthHandler(w, r)
		/*{
		    "address": "0x9849943a82AFA29EcFEC61e80AfdfE7EA4357a33"
		}*/
		case "/api/blockscanner": // hd wallet oluşturup redise yazar
			blockScannerHandler(w, r)

		case "/api/alltransactions":
			getAllTransactionAddressesHandler(w, r)

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

func basicAuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Kimlik Doğrulanamadı", http.StatusUnauthorized)
			return
		}

		authParts := strings.Split(auth, " ")
		if len(authParts) != 2 || authParts[0] != "Basic" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(authParts[1])
		if err != nil {
			http.Error(w, "Invalid base64 encoding", http.StatusUnauthorized)
			return
		}

		userAndPassword := strings.Split(string(decoded), ":")
		if len(userAndPassword) != 2 || userAndPassword[0] != username || userAndPassword[1] != password {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	}
}
