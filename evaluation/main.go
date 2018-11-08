package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cha87de/tsprofiler/impl"
	"github.com/cha87de/tsprofiler/spec"
)

var profiler spec.TSProfiler

func main() {
	// create new ts profiler
	initProfiler()

	// read file line by line
	readFile("tsinput.csv")

	// get and print profile
	profileOutput()
}

func initProfiler() {
	profiler = impl.NewProfiler(spec.Settings{
		Name:       "tsinput",
		BufferSize: 10,
		States:     4,
	})
}

func readFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		utilValue, err := strconv.ParseFloat(line, 64)
		if err != nil {
			continue
		}
		putMeasurement(utilValue)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func putMeasurement(utilValue float64) {
	metrics := make([]spec.TSDataMetric, 0)
	metrics = append(metrics, spec.TSDataMetric{
		Name:  "example",
		Value: utilValue,
		Max:   float64(100), // TODO get from data dynamically
	})
	tsdata := spec.TSData{
		Metrics: metrics,
	}
	profiler.Put(tsdata)
}

func profileOutput() {
	profile := profiler.Get()
	json, err := json.Marshal(profile)
	if err != nil {
		fmt.Printf("cannot create json: %s (original: %v)\n", err, profile)
		return
	}
	fmt.Printf("%s\n", json)
}
