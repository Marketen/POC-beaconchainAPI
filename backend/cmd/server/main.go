package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Marketen/POC-beaconchainAPI/backend/internal/api"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/util"
)

var (
	rateLimit = time.Second / 4 // 4 calls per second (rate limit is 5 but we are cautious)
	apiTokens = make(chan struct{}, 1) // strict global rate limiter, no burst
)

func init() {
	go func() {
		ticker := time.NewTicker(rateLimit)
		defer ticker.Stop()
		for {
			<-ticker.C
			apiTokens <- struct{}{}
		}
	}()
}

func acquireAPIToken() {
	<-apiTokens
}

func main() {
	db.LoadDB()
	http.HandleFunc("/validators", api.WithCORS(api.HandleGetValidators))
	http.HandleFunc("/validators/add", api.WithCORS(api.HandleAddIndices))
	http.HandleFunc("/validators/delete", api.WithCORS(api.HandleDeleteIndices))
	http.HandleFunc("/ethstore_apr", api.WithCORS(api.HandleGetEthstoreAPR))
	http.HandleFunc("/force_update", api.WithCORS(api.HandleForceUpdate))
	http.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("."))))

	mux := http.NewServeMux()
	mux.HandleFunc("/validators", api.WithCORS(api.HandleGetValidators))
	mux.HandleFunc("/validators/add", api.WithCORS(api.HandleAddIndices))
	mux.HandleFunc("/validators/delete", api.WithCORS(api.HandleDeleteIndices))
	mux.HandleFunc("/ethstore_apr", api.WithCORS(api.HandleGetEthstoreAPR))
	mux.HandleFunc("/force_update", api.WithCORS(api.HandleForceUpdate))
	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("."))))

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	}

	go func() {
		for {
			log.Println("[CRON] tryFetchAndUpdate started")
			api.TryFetchAndUpdate()
			log.Println("[CRON] tryFetchAndUpdate finished")
			time.Sleep(60 * time.Minute)
		}
	}()

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handler)))
}
