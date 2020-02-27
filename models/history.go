package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// History defines the historic path and next step for tspredictor
type History struct {
	CurrentPhase    int                 `json:"currentPhase"`
	HistoricStates  []map[string]string `json:"historicStates"`
	PeriodPath      []int               `json:"periodPath"`
	PeriodPathDepth int                 `json:"periodPathDepth"`
	NextState       map[string]string   `json:"nextState"`
}

// ReadHistoryFromFile reads History from json file and returns as object
func ReadHistoryFromFile(filepath string) History {
	filehandler, err := os.Open(filepath)
	if err != nil {
		fmt.Println(err)
	}
	defer filehandler.Close()
	byteValue, _ := ioutil.ReadAll(filehandler)

	var history History
	json.Unmarshal(byteValue, &history)

	return history
}
