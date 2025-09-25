package beacon

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Marketen/POC-beaconchainAPI/backend/internal/db"
	"github.com/Marketen/POC-beaconchainAPI/backend/internal/model"
)

var (
	BaseURL        string
	APIKey         string
	requestTimeout = 5 * time.Second
	apiRateLimit   = time.Second / 4 // 4 req/sec
	rateLimiter    = make(chan struct{}, 1)
)

func init() {
	go func() {
		ticker := time.NewTicker(apiRateLimit)
		defer ticker.Stop()
		for {
			<-ticker.C
			select {
			case rateLimiter <- struct{}{}:
			default:
				// channel full, do nothing
			}
		}
	}()
}

func acquireRateLimit() {
	select {
	case <-rateLimiter:
		// allowed
	default:
		log.Println("[RateLimit] Waiting to avoid external API rate limit...")
		<-rateLimiter
	}
}

func batchIndices(indices []string, batchSize int) [][]string {
	var batches [][]string
	for batchSize < len(indices) {
		batches = append(batches, indices[:batchSize])
		indices = indices[batchSize:]
	}
	if len(indices) > 0 {
		batches = append(batches, indices)
	}
	return batches
}

func httpGetWithTimeout(url string) (*http.Response, error) {
	acquireRateLimit()
	log.Printf("[HTTP] Preparing request: GET %s", url)
	// Remove context timeout, use default client
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("[HTTP] Request details: Method=%s URL=%s Headers=%v", req.Method, req.URL.String(), req.Header)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[HTTP] Error for %s: %v", url, err)
	}
	return resp, err
}

func FetchValidatorInfo(indices string) (map[string]model.ValidatorData, error) {
	idxList := strings.Split(indices, ",")
	const batchSize = 100
	batches := batchIndices(idxList, batchSize)
	resultMap := make(map[string]model.ValidatorData)
	for i, batch := range batches {
		joined := strings.Join(batch, ",")
		log.Printf("[FetchValidatorInfo] Batch %d/%d size: %d", i+1, len(batches), len(batch))
		url := fmt.Sprintf("%s/api/v1/validator/%s?apikey=%s", BaseURL, joined, APIKey)
		resp, err := httpGetWithTimeout(url)
		if err != nil {
			log.Printf("[FetchValidatorInfo] HTTP error: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		var result struct {
			Data []struct {
				Index            json.Number `json:"validatorindex"`
				Balance          uint64      `json:"balance"`
				EffectiveBalance uint64      `json:"effective_balance"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[FetchValidatorInfo] HTTP status: %d, body: %s", resp.StatusCode, string(body))
			return nil, err
		}
		for _, v := range result.Data {
			idxStr := v.Index.String()
			resultMap[idxStr] = model.ValidatorData{
				Index:            idxStr,
				Balance:          v.Balance,
				EffectiveBalance: v.EffectiveBalance,
			}
		}
	}
	return resultMap, nil
}

func FetchValidatorPerformance(indices string) (map[string]model.ValidatorPerformance, error) {
	idxList := strings.Split(indices, ",")
	const batchSize = 100
	batches := batchIndices(idxList, batchSize)
	resultMap := make(map[string]model.ValidatorPerformance)
	for i, batch := range batches {
		joined := strings.Join(batch, ",")
		log.Printf("[FetchValidatorPerformance] Batch %d/%d size: %d", i+1, len(batches), len(batch))
		url := fmt.Sprintf("%s/api/v1/validator/%s/performance?apikey=%s", BaseURL, joined, APIKey)
		resp, err := httpGetWithTimeout(url)
		if err != nil {
			log.Printf("[FetchValidatorPerformance] HTTP error: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		var result struct {
			Data []struct {
				Index           json.Number `json:"validatorindex"`
				Performance1d   float64     `json:"performance1d"`
				Performance7d   float64     `json:"performance7d"`
				Performance31d  float64     `json:"performance31d"`
				Performance365d float64     `json:"performance365d"`
				Rank7d          int         `json:"rank7d"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[FetchValidatorPerformance] HTTP status: %d, body: %s", resp.StatusCode, string(body))
			return nil, err
		}
		for _, v := range result.Data {
			idxStr := v.Index.String()
			resultMap[idxStr] = model.ValidatorPerformance{
				Index:           idxStr,
				Performance1d:   v.Performance1d,
				Performance7d:   v.Performance7d,
				Performance31d:  v.Performance31d,
				Performance365d: v.Performance365d,
				Rank7d:          v.Rank7d,
			}
		}
	}
	return resultMap, nil
}

func FetchValidatorExecPerformance(indices string) (map[string]model.ValidatorExecPerformance, error) {
	idxList := strings.Split(indices, ",")
	const batchSize = 100
	batches := batchIndices(idxList, batchSize)
	resultMap := make(map[string]model.ValidatorExecPerformance)
	for i, batch := range batches {
		joined := strings.Join(batch, ",")
		log.Printf("[FetchValidatorExecPerformance] Batch %d/%d size: %d", i+1, len(batches), len(batch))
		url := fmt.Sprintf("%s/api/v1/validator/%s/execution/performance?apikey=%s", BaseURL, joined, APIKey)
		resp, err := httpGetWithTimeout(url)
		if err != nil {
			log.Printf("[FetchValidatorExecPerformance] HTTP error: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		var result struct {
			Data []struct {
				Index          json.Number `json:"validatorindex"`
				Performance1d  float64     `json:"performance1d"`
				Performance7d  float64     `json:"performance7d"`
				Performance31d float64     `json:"performance31d"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			body, _ := io.ReadAll(resp.Body)
			log.Printf("[FetchValidatorExecPerformance] HTTP status: %d, body: %s", resp.StatusCode, string(body))
			return nil, err
		}
		for _, v := range result.Data {
			idxStr := v.Index.String()
			resultMap[idxStr] = model.ValidatorExecPerformance{
				Index:          idxStr,
				Performance1d:  v.Performance1d,
				Performance7d:  v.Performance7d,
				Performance31d: v.Performance31d,
			}
		}
	}
	return resultMap, nil
}

func FetchEthstoreAPR() (float64, error) {
	url := fmt.Sprintf("%s/api/v1/ethstore/latest?apikey=%s", BaseURL, APIKey)
	resp, err := httpGetWithTimeout(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			APR float64 `json:"apr"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Data.APR, nil
}

func FetchFinalizedEpoch() (uint64, error) {
	url := fmt.Sprintf("%s/api/v1/epoch/finalized?apikey=%s", BaseURL, APIKey)
	resp, err := httpGetWithTimeout(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var result struct {
		Data struct {
			Epoch uint64 `json:"epoch"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Data.Epoch, nil
}

func TryFetchAndUpdate(indices []string, dbInst *model.DB) error {
	if len(indices) == 0 {
		log.Println("[TryFetchAndUpdate] No indices provided, skipping fetch.")
		return nil
	}
	if dbInst.Validators == nil {
		dbInst.Validators = make(map[string]model.ValidatorData)
	}
	log.Printf("[TryFetchAndUpdate] Updating %d indices", len(indices))

	log.Println("[TryFetchAndUpdate] Calling FetchValidatorInfo")
	info, err := FetchValidatorInfo(strings.Join(indices, ","))
	if err != nil {
		log.Printf("[TryFetchAndUpdate] Error in FetchValidatorInfo: %v", err)
		return err
	}

	log.Println("[TryFetchAndUpdate] Calling FetchValidatorPerformance")
	perf, err := FetchValidatorPerformance(strings.Join(indices, ","))
	if err != nil {
		log.Printf("[TryFetchAndUpdate] Error in FetchValidatorPerformance: %v", err)
		return err
	}

	log.Println("[TryFetchAndUpdate] Calling FetchValidatorExecPerformance")
	execPerf, err := FetchValidatorExecPerformance(strings.Join(indices, ","))
	if err != nil {
		log.Printf("[TryFetchAndUpdate] Error in FetchValidatorExecPerformance: %v", err)
		return err
	}

	for _, idx := range indices {
		v, vOk := info[idx]
		p, pOk := perf[idx]
		e, eOk := execPerf[idx]

		if vOk {
			// Attach consensus performance if available
			if pOk {
				v.Consensus = &model.ConsensusPerformance{
					Performance1d:   p.Performance1d,
					Performance7d:   p.Performance7d,
					Performance31d:  p.Performance31d,
					Performance365d: p.Performance365d,
					Rank7d:          p.Rank7d,
				}
			}
			// Attach execution performance if available
			if eOk {
				v.Execution = &model.ExecutionPerformance{
					Performance1d:  e.Performance1d,
					Performance7d:  e.Performance7d,
					Performance31d: e.Performance31d,
				}
			}
			dbInst.Validators[idx] = v
		}
	}

	log.Println("[TryFetchAndUpdate] Calling FetchEthstoreAPR")
	apr, err := FetchEthstoreAPR()
	if err == nil {
		dbInst.EthstoreAPR = apr
		log.Printf("[TryFetchAndUpdate] EthstoreAPR updated: %f", apr)
	} else {
		log.Printf("[TryFetchAndUpdate] Error in FetchEthstoreAPR: %v", err)
	}

	db.SaveDB(dbInst)
	return nil
}
