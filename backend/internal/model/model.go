package model

import "sync"

type ConsensusPerformance struct {
	Performance1d   float64 `json:"performance1d"`
	Performance7d   float64 `json:"performance7d"`
	Performance31d  float64 `json:"performance31d"`
	Performance365d float64 `json:"performance365d,omitempty"`
	Rank7d          int     `json:"rank7d,omitempty"`
}

type ExecutionPerformance struct {
	Performance1d  float64 `json:"performance1d"`
	Performance7d  float64 `json:"performance7d"`
	Performance31d float64 `json:"performance31d"`
}

type ValidatorPerformance struct {
	Index          string  `json:"index"`
	Performance1d  float64 `json:"performance1d"`
	Performance7d  float64 `json:"performance7d"`
	Performance31d float64 `json:"performance31d"`
	Performance365d float64 `json:"performance365d"`
	Rank7d         int     `json:"rank7d"`
}

type ValidatorExecPerformance struct {
	Index          string  `json:"index"`
	Performance1d  float64 `json:"performance1d"`
	Performance7d  float64 `json:"performance7d"`
	Performance31d float64 `json:"performance31d"`
}

type ValidatorData struct {
	Index            string                `json:"index"`
	Balance          uint64                `json:"balance"`
	EffectiveBalance uint64                `json:"effective_balance"`
	Consensus        *ConsensusPerformance `json:"consensus,omitempty"`
	Execution        *ExecutionPerformance `json:"execution,omitempty"`
}

type DB struct {
	LastEpoch        uint64                            `json:"last_epoch"`
	Validators       map[string]ValidatorData          `json:"validators"`
	EthstoreAPR      float64                           `json:"ethstore_apr,omitempty"`
	Mutex            sync.Mutex                        `json:"-"`
}
