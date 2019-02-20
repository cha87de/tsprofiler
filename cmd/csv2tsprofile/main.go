package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/profiler"
	flags "github.com/jessevdk/go-flags"
)

var options struct {
	States            int     `long:"states" default:"4"`
	BufferSize        int     `long:"buffersize" default:"10"`
	History           int     `long:"history" default:"1"`
	FilterStdDevs     int     `long:"filterstddevs" default:"2"`
	FixedBound        bool    `long:"fixedbound"`
	FixedMin          float64 `long:"fixedmin" default:"0" description:"if fixedbound is set, set the min value"`
	FixedMax          float64 `long:"fixedmax" default:"100" description:"if fixedbound is set, set the max value"`
	PeriodSize        string  `long:"periodsize" default:"60,720,1440" description:"comma separated list of ints, specifies descrete states per period"`
	PeriodChangeRatio float64 `long:"periodchangeratio" default:"0.2" description:"accepted ratio [0,1] for changes, alert if above"`
	Inputfile         string
}

var tsprofiler api.TSProfiler

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
	// convert periodSize string array to int array
	periodSizeStr := strings.Split(options.PeriodSize, ",")
	periodSize := make([]int, len(periodSizeStr))
	for i, s := range periodSizeStr {
		periodSize[i], _ = strconv.Atoi(s)
	}

	// create new profiler
	tsprofiler = profiler.NewProfiler(models.Settings{
		Name:          "csv2tsprofile",
		BufferSize:    options.BufferSize,
		States:        options.States,
		FilterStdDevs: options.FilterStdDevs,
		History:       options.History,
		FixBound:      options.FixedBound,
		PeriodSize:    periodSize,
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
	metrics := make([]models.TSInputMetric, 0)
	for i, value := range utilValue {
		metrics = append(metrics, models.TSInputMetric{
			Name:     fmt.Sprintf("metric_%d", i),
			Value:    value,
			FixedMin: options.FixedMin,
			FixedMax: options.FixedMax,
		})
	}
	tsinput := models.TSInput{
		Metrics: metrics,
	}
	tsprofiler.Put(tsinput)
}

func profileOutput() {
	profile := tsprofiler.Get()
	json, err := json.Marshal(profile)
	if err != nil {
		fmt.Printf("cannot create json: %s (original: %+v)\n", err, profile)
		return
	}
	fmt.Printf("%s\n", json)
}
