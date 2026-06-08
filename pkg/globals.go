package h0neytr4p

import (
	"os"
	"sync"
)

type Trap struct {
	Basicinfo BasicInfo   `json:"BasicInfo"`
	Behaviour []Behaviour `json:"Behaviour"`
}

type BasicInfo struct {
	Name            string `json:"Name"`
	Port            string `json:"Port"`
	Protocol        string `json:"Protocol"`
	Mitreattacktags string `json:"MitreAttackTags"`
	RiskRating      string `json:"RiskRating"`
	References      string `json:"References"`
	Description     string `json:"Description"`
}

type Behaviour struct {
	Request  Request  `json:"Request"`
	Response Response `json:"Response"`
	Trap     string   `json:"trap,omitempty"`
}

type Request struct {
	URL            string                 `json:"Url"`
	Method         string                 `json:"Method"`
	Headers        map[string]interface{} `json:"Headers"`
	Params         map[string]interface{} `json:"Params"`
	HeaderContains map[string][]string    `json:"HeaderContains"`
}

type Response struct {
	Statuscode int                    `json:"StatusCode"`
	Body       string                 `json:"Body"`
	Headers    map[string]interface{} `json:"Headers"`
	Type       string                 `json:"Type"`
}

const (
	MaxMultipartSize = 101 * 1024 // 101KB
	MaxJSONFormSize  = 11 * 1024  // 11KB

	// T-Pot expects group-readable/group-writable runtime artifacts.
	RuntimeDirMode  os.FileMode = 0775
	RuntimeFileMode os.FileMode = 0775
)

var (
	logFile       *os.File
	logFileMutex  sync.Mutex
	payloadFolder string
	Verbose       bool
)
