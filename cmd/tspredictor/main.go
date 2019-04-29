package main

import (
	"fmt"
	"os"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
	"github.com/cha87de/tsprofiler/utils"
	flags "github.com/jessevdk/go-flags"
)

var options struct {
	Steps     int                      `long:"steps" default:"40"`
	Mode      predictor.PredictionMode `long:"mode" default:"0"`
	Inputfile string
}

func main() {
	initializeFlags()

	profile := utils.ReadProfileFromFile(options.Inputfile)
	predictor := predictor.NewPredictor(profile)
	predictor.SetMode(options.Mode)
	/*predictor.SetState(map[string]string{
		"metric_0": "0",
	})*/
	simulation := predictor.Simulate(options.Steps)

	printSimulation(simulation)
}

func initializeFlags() {
	// initialize parser for flags
	parser := flags.NewParser(&options, flags.Default)
	parser.ShortDescription = "tspredictor"
	parser.LongDescription = "Simulates the next steps for a given tsprofile json file"
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

func printSimulation(simulation [][]models.TSState) {
	if len(simulation) <= 0 {
		return
	}

	// print header
	for i, tsstate := range simulation[0] {
		if i > 0 {
			fmt.Printf(",")
		}
		fmt.Printf("%s", tsstate.Metric)
	}
	fmt.Printf("\n")

	// print rows
	for _, simstep := range simulation {
		for i, tsstate := range simstep {
			if i > 0 {
				fmt.Printf(",")
			}
			fmt.Printf("%d", tsstate.State.Value)
		}
		fmt.Printf("\n")
	}
}
