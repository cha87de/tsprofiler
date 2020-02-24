package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	States                int     `long:"states" default:"4"`
	BufferSize            int     `long:"buffersize" default:"10"`
	History               int     `long:"history" default:"1"`
	FilterStdDevs         int     `long:"filterstddevs" default:"2"`
	FixedBound            bool    `long:"fixedbound"`
	FixedMin              float64 `long:"fixedmin" default:"0" description:"if fixedbound is set, set the min value"`
	FixedMax              float64 `long:"fixedmax" default:"100" description:"if fixedbound is set, set the max value"`
	PeriodSize            string  `long:"periodsize" default:"" description:"comma separated list of ints, specifies descrete states per period"`
	PeriodChangeRatio     float64 `long:"periodchangeratio" default:"0.2" description:"accepted ratio [0,1] for changes, alert if above"`
	PhaseChangeLikeliness float32 `long:"phasechangelikeliness" default:"0.6"`
	PhaseChangeMincount   int64   `long:"phasechangemincount" default:"60"`
	Outputfile            string  `long:"output" default:"-" description:"path to write profile to, stdout if '-'"`
	PhasesFile            string  `long:"out.phases" default:""`
	StatesFile            string  `long:"out.states" default:""`
	Inputfile             string
}

var tsprofiler api.TSProfiler
var phasesfile *os.File
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
		Name:                  "csv2tsprofile",
		BufferSize:            options.BufferSize,
		States:                options.States,
		FilterStdDevs:         options.FilterStdDevs,
		History:               options.History,
		FixBound:              options.FixedBound,
		PeriodSize:            periodSize,
		PhaseChangeLikeliness: options.PhaseChangeLikeliness,
		PhaseChangeMincount:   options.PhaseChangeMincount,
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

func profileOutput() {
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
