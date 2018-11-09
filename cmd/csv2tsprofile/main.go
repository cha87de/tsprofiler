package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/cha87de/tsprofiler/impl"
	"github.com/cha87de/tsprofiler/spec"
	flags "github.com/jessevdk/go-flags"
)

var options struct {
	States     int `long:"states" default:"4"`
	BufferSize int `long:"buffersize" default:"10"`
	Inputfile  string
}

var profiler spec.TSProfiler

func main() {
	initializeFlags()

	// create new ts profiler
	initProfiler()

	// read file line by line
	readFile(options.Inputfile)

	// get and print profile
	profileOutput()
}

func initializeFlags() {
	// initialize parser for flags
	parser := flags.NewParser(&options, flags.Default)
	parser.ShortDescription = "csv2tsprofile"
	parser.LongDescription = "Reads time series values from a CSV file and generates a tsprofile"
	parser.ArgsRequired = true

	// Parse parameters
	args, err := parser.Parse()
	if err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		if code != 0 {
			fmt.Printf("Error parsing flags: %s", err)
		}
		os.Exit(code)
	}

	if len(args) < 1 {
		fmt.Printf("No input file specified.\n")
		os.Exit(1)
	}
	options.Inputfile = args[0]
}

func initProfiler() {
	profiler = impl.NewProfiler(spec.Settings{
		Name:       "csv2tsprofile",
		BufferSize: options.BufferSize,
		States:     options.States,
	})
}

func readFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		var utilValues []float64
		for _, rawValue := range record {
			utilValue, err := strconv.ParseFloat(rawValue, 64)
			if err != nil {
				continue
			}
			utilValues = append(utilValues, utilValue)
		}
		putMeasurement(utilValues)
	}

}

func putMeasurement(utilValue []float64) {
	metrics := make([]spec.TSDataMetric, 0)
	for i, value := range utilValue {
		metrics = append(metrics, spec.TSDataMetric{
			Name:  fmt.Sprintf("metric_%d", i),
			Value: value,
			Max:   float64(100), // TODO get from data dynamically
		})
	}
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
