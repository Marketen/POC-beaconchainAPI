package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Marketen/POC-beaconchainAPI/backend/internal/api"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/db"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/beacon"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/model"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/util"
)

var (
	rateLimit = time.Second / 4 // 4 calls per second (rate limit is 5 but we are cautious)
	apiTokens = make(chan struct{}, 1) // strict global rate limiter, no burst
)

var dbInst model.DB

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
	beacon.BaseURL = util.GetEnv("BEACON_API_BASE", "https://beaconcha.in")
	beacon.APIKey = util.GetEnv("BEACON_API_KEY", "")
	
	db.LoadDB(&dbInst)

	mux := http.NewServeMux()
	mux.HandleFunc("/validators", api.WithCORS(api.HandleGetValidators(&dbInst, util.GetEnv("INDICES_ENV", ""))))
	mux.HandleFunc("/validators/add", api.WithCORS(api.HandleAddIndices(&dbInst)))
	mux.HandleFunc("/validators/delete", api.WithCORS(api.HandleDeleteIndices(&dbInst)))
	mux.HandleFunc("/ethstore_apr", api.WithCORS(api.HandleGetEthstoreAPR(&dbInst)))
	mux.HandleFunc("/force_update", api.WithCORS(api.HandleForceUpdate(func() {
		dbInst.Mutex.Lock()
		indices := make([]string, 0, len(dbInst.Validators))
		for idx := range dbInst.Validators {
			indices = append(indices, idx)
		}
		dbInst.Mutex.Unlock()
		beacon.TryFetchAndUpdate(indices, &dbInst)
	})))
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
			dbInst.Mutex.Lock()
			indices := make([]string, 0, len(dbInst.Validators))
			for idx := range dbInst.Validators {
				indices = append(indices, idx)
			}
			dbInst.Mutex.Unlock()
			log.Println("[CRON] tryFetchAndUpdate started")
			beacon.TryFetchAndUpdate(indices, &dbInst)
			log.Println("[CRON] tryFetchAndUpdate finished")
			time.Sleep(60 * time.Minute)
		}
	}()

	log.Println("API listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(handler)))
}
