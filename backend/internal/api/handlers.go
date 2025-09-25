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

// ...other handlers (HandleAddIndices, HandleDeleteIndices, HandleGetEthstoreAPR, etc.)
