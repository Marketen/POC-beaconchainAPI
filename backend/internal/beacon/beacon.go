package beacon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/model"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/util"
)

var (
	BaseURL   string
	APIKey    string
	RateLimit = time.Second / 4 // 4 calls per second
	APITokens = make(chan struct{}, 1)
)

func InitRateLimiter() {
	go func() {
		ticker := time.NewTicker(RateLimit)
		defer ticker.Stop()
		for {
			<-ticker.C
			APITokens <- struct{}{}
		}
	}()
}

func AcquireAPIToken() {
	<-APITokens
}

func FetchValidatorInfo(indices string) (map[string]model.ValidatorData, error) {
	AcquireAPIToken()
	url := fmt.Sprintf("%s/api/v1/validator/%s?apikey=%s", BaseURL, indices, APIKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Data []struct {
			Index            json.Number `json:"validatorindex"`
			Balance          uint64      `json:"balance"`
			EffectiveBalance uint64      `json:"effectivebalance"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	out := make(map[string]model.ValidatorData)
	for _, v := range result.Data {
		idxStr := v.Index.String()
		out[idxStr] = model.ValidatorData{
			Index:            idxStr,
			Balance:          v.Balance,
			EffectiveBalance: v.EffectiveBalance,
		}
	}
	return out, nil
}

// ...other fetch functions (FetchValidatorPerformance, FetchValidatorExecPerformance, FetchEthstoreAPR, etc.)
