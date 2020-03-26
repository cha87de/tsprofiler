package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cha87de/tsprofiler/models"
)

func outputHistory() {
	if options.Historyfile == "" {
		// output disabled, ignore method call
		return
	}
	state := tsprofiler.GetCurrentState()
	historicStates := make([]map[string]string, 1)
	historicStates[0] = make(map[string]string)
	for _, s := range state {
		historicStates[0][s.Metric] = fmt.Sprintf("%d", s.State.Value)
	}

	history := models.History{
		CurrentPhase:   tsprofiler.GetCurrentPhase(),
		PeriodPath:     tsprofiler.GetCurrentPeriodPath(),
		HistoricStates: historicStates,
	}

	json, err := json.Marshal(history)
	if err != nil {
		fmt.Printf("cannot create json: %s (original: %+v)\n", err, history)
		return
	}

	// print to stdout
	if options.Historyfile == "-" {
		// print to stdout
		fmt.Printf("%s\n", json)
	} else {
		// write to file
		err := ioutil.WriteFile(options.Historyfile, json, 0644)
		if err != nil {
			fmt.Printf("cannot write json to file %s: %s\n", options.Historyfile, err)
			return
		}
	}
}

func outputProfile() {
	profile := tsprofiler.Get()
	json, err := json.Marshal(profile)
	if err != nil {
		fmt.Printf("cannot create json: %s (original: %+v)\n", err, profile)
		return
	}

	if options.Outputfile == "-" {
		// print to stdout
		fmt.Printf("%s\n", json)
	} else {
		// write to file
		err := ioutil.WriteFile(options.Outputfile, json, 0644)
		if err != nil {
			fmt.Printf("cannot write json to file %s: %s\n", options.Outputfile, err)
			return
		}
	}
}
