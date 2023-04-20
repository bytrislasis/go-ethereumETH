package satoshiturk

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

func startServer() {
	http.HandleFunc("/", handler)
	http.HandleFunc("api/tx", apiHandler) // API yolu için rota

	http.ListenAndServe(":1983", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	//disable cors
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:1983")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	fmt.Println("method:", r.Method+" "+r.URL.Path)

	t, err := template.ParseFiles(filepath.Join("satoshiturk/public", "index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		switch r.URL.Path {
		case "api/tx":
			txSorgula(w, r)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	} else {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func init() {
	go startServer()
}
