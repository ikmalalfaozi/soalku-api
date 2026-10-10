package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)

	addr := ":8080"
	log.Println("server berjalan di", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
