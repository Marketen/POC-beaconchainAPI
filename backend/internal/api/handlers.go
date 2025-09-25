package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/model"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/db"
)

func HandleGetValidators(dbInst *model.DB, indicesEnv string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /validators called from %s", r.RemoteAddr)
		indicesParam := r.URL.Query().Get("indices")
		if indicesParam == "" {
			indicesParam = indicesEnv
		}
		indicesList := strings.Split(indicesParam, ",")
		resp := make(map[string]model.ValidatorData)
		dbInst.Mutex.Lock()
		for _, idx := range indicesList {
			if v, ok := dbInst.Validators[idx]; ok {
				resp[idx] = v
			}
		}
		dbInst.Mutex.Unlock()
		log.Printf("[API] /validators response: %d validators returned", len(resp))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func HandleAddIndices(dbInst *model.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /validators/add called from %s", r.RemoteAddr)
		var req struct {
			Indices []interface{} `json:"indices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[API] /validators/add error: invalid json: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf(`{"error":"invalid json: %v"}`, err)))
			return
		}
		if len(req.Indices) == 0 {
			log.Printf("[API] /validators/add error: no indices provided")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"no indices provided"}`))
			return
		}
		indicesStr := make([]string, 0, len(req.Indices))
		for _, idx := range req.Indices {
			switch v := idx.(type) {
			case string:
				indicesStr = append(indicesStr, v)
			case float64:
				indicesStr = append(indicesStr, fmt.Sprintf("%.0f", v))
			case int:
				indicesStr = append(indicesStr, fmt.Sprintf("%d", v))
			case json.Number:
				indicesStr = append(indicesStr, v.String())
			default:
				log.Printf("[API] /validators/add error: unknown index type: %T", v)
			}
		}
		if len(indicesStr) == 0 {
			log.Printf("[API] /validators/add error: no valid indices after parsing")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"no valid indices after parsing"}`))
			return
		}
		dbInst.Mutex.Lock()
		if dbInst.Validators == nil {
			log.Printf("[API] /validators/add error: dbInst.Validators map is nil, initializing.")
			dbInst.Validators = make(map[string]model.ValidatorData)
		}
		added := 0
		for _, idx := range indicesStr {
			if _, exists := dbInst.Validators[idx]; !exists {
				dbInst.Validators[idx] = model.ValidatorData{Index: idx}
				added++
			}
		}
		dbInst.Mutex.Unlock()
		db.SaveDB(dbInst)
		log.Printf("[API] /validators/add: %d new validators added", added)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"indices added", "added": %d}`, added)))
	}
}

func HandleDeleteIndices(dbInst *model.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /validators/delete called from %s", r.RemoteAddr)
		var req struct {
			Indices []string `json:"indices"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid json"}`))
			log.Printf("[API] /validators/delete error: invalid json")
			return
		}
		if len(req.Indices) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"no indices provided"}`))
			log.Printf("[API] /validators/delete error: no indices provided")
			return
		}
		dbInst.Mutex.Lock()
		deleted := 0
		for _, idx := range req.Indices {
			if _, exists := dbInst.Validators[idx]; exists {
				delete(dbInst.Validators, idx)
				deleted++
			}
		}
		dbInst.Mutex.Unlock()
		db.SaveDB(dbInst)
		log.Printf("[API] /validators/delete: %d validators deleted", deleted)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"indices deleted"}`))
	}
}

func HandleGetEthstoreAPR(dbInst *model.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbInst.Mutex.Lock()
		apr := dbInst.EthstoreAPR
		dbInst.Mutex.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ethstore_apr": apr})
	}
}

func HandleForceUpdate(tryFetchAndUpdate func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[API] /force_update called from %s", r.RemoteAddr)
		go func() {
			log.Println("[API] /force_update: tryFetchAndUpdate started")
			tryFetchAndUpdate()
			log.Println("[API] /force_update: tryFetchAndUpdate finished")
		}()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"update started"}`))
	}
}

func WithCORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}
