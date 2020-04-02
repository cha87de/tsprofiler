package main

import (
	"encoding/csv"
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
	States        int `long:"states" default:"4"`
	BufferSize    int `long:"buffersize" default:"10"`
	History       int `long:"history" default:"1"`
	FilterStdDevs int `long:"filterstddevs" default:"2"`

	FixedBound bool    `long:"fixedbound"`
	FixedMin   float64 `long:"fixedmin" default:"0" description:"if fixedbound is set, set the min value"`
	FixedMax   float64 `long:"fixedmax" default:"100" description:"if fixedbound is set, set the max value"`

	PeriodSize string `long:"periodsize" default:"" description:"comma separated list of ints, specifies descrete states per period"`

	PhaseChangeLikeliness     float32 `long:"phasechangelikeliness" default:""`
	PhaseChangeHistory        int64   `long:"phasechangehistory" default:"1"`
	PhaseChangeHistoryFadeout bool    `long:"phasechangehistoryfadeout"`

	Outputfile  string `long:"output" default:"-" description:"path to write profile to, stdout if '-'"`
	Historyfile string `long:"out.history" default:"" description:"path to write last historic values to, stdout if '-', empty to disable"`
	PhasesFile  string `long:"out.phases" default:""`
	PeriodsFile string `long:"out.periods" default:""`
	StatesFile  string `long:"out.states" default:""`

	Inputfile string
}

var tsprofiler api.TSProfiler
var phasesfile *os.File
var periodsfile *os.File
var statesfile *os.File

func main() {
	initializeFlags()

	// create new ts profiler
	initProfiler()

	// create & open output file
	if options.PhasesFile != "" && options.PhasesFile != "-" {
		var err error
		phasesfile, err = os.OpenFile(options.PhasesFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer phasesfile.Close()
	}
	if options.PeriodsFile != "" && options.PeriodsFile != "-" {
		var err error
		periodsfile, err = os.OpenFile(options.PeriodsFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer periodsfile.Close()
	}
	if options.StatesFile != "" && options.StatesFile != "-" {
		var err error
		statesfile, err = os.OpenFile(options.StatesFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer statesfile.Close()
	}

	// read file line by line
	readFile(options.Inputfile)

	// get and print profile
	outputProfile()

	// print last states and positions
	if options.Historyfile != "" {
		outputHistory()
	}

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
	periodSize := make([]int, 0)
	for _, s := range periodSizeStr {
		if s == "" {
			continue
		}
		si, _ := strconv.Atoi(s)
		periodSize = append(periodSize, si)
	}

	// create new profiler
	tsprofiler = profiler.NewProfiler(models.Settings{
		Name:                      "csv2tsprofile",
		BufferSize:                options.BufferSize,
		States:                    options.States,
		FilterStdDevs:             options.FilterStdDevs,
		History:                   options.History,
		FixBound:                  options.FixedBound,
		PeriodSize:                periodSize,
		PhaseChangeLikeliness:     options.PhaseChangeLikeliness,
		PhaseChangeHistory:        options.PhaseChangeHistory,
		PhaseChangeHistoryFadeout: options.PhaseChangeHistoryFadeout,
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

	// print phases
	phaseid := tsprofiler.GetCurrentPhase()
	if options.PhasesFile == "-" {
		// use stdout
		row := fmt.Sprintf("%d\n", phaseid)
		fmt.Print(row)
	} else if options.PhasesFile == "" {
		// ignore likeliness
	} else {
		// print to file
		row := fmt.Sprintf("%d\n", phaseid)
		if _, err := phasesfile.Write([]byte(row)); err != nil {
			log.Fatal(err)
		}
	}

	// print periods
	periodPath := strings.Trim(strings.Replace(fmt.Sprint(tsprofiler.GetCurrentPeriodPath()), " ", ",", -1), "[]")
	if options.PeriodsFile == "-" {
		// use stdout
		row := fmt.Sprintf("%s\n", periodPath)
		fmt.Print(row)
	} else if options.PeriodsFile == "" {
		// ignore
	} else {
		// print to file
		row := fmt.Sprintf("%s\n", periodPath)
		if _, err := periodsfile.Write([]byte(row)); err != nil {
			log.Fatal(err)
		}
	}

	// print states
	state := tsprofiler.GetCurrentState()
	stateRow := ""
	for _, s := range state {
		if stateRow != "" {
			stateRow = stateRow + " "
		}
		stateRow = fmt.Sprintf("%s%d", stateRow, s.State.Value)
	}
	stateRow = stateRow + "\n"
	if options.StatesFile == "-" {
		// use stdout
		fmt.Print(stateRow)
	} else if options.StatesFile == "" {
		// ignore
	} else {
		// print to file
		if _, err := statesfile.Write([]byte(stateRow)); err != nil {
			log.Fatal(err)
		}
	}
}
