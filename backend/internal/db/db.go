package db

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/model"
)

var DBFile = "/app/db.json"

func LoadDB(db *model.DB) {
	b, err := ioutil.ReadFile(DBFile)
	if err != nil {
		log.Printf("db.json not found, creating new DB file: %v", err)
		// Create a new DB file with default values
		db.LastEpoch = 0
		db.Validators = make(map[string]model.ValidatorData)
		db.EthstoreAPR = 0
		SaveDB(db)
		return
	}
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	json.Unmarshal(b, db)
}

func SaveDB(db *model.DB) {
	db.Mutex.Lock()
	defer db.Mutex.Unlock()
	b, _ := json.MarshalIndent(db, "", "  ")
	_ = ioutil.WriteFile(DBFile, b, 0644)
}
