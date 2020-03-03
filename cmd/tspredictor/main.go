package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/cha87de/tsprofiler/cmd/tspredictor/task"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/predictor"
	"github.com/cha87de/tsprofiler/utils"
	flags "github.com/jessevdk/go-flags"
)

var options struct {
	Steps       int                      `long:"steps" default:"40"`
	Mode        predictor.PredictionMode `long:"mode" default:"0"`
	PeriodDepth int                      `long:"periodDepth" default:"0"`
	Profilefile string                   `long:"profile" short:"p"`
	Historyfile string                   `long:"history" short:"h"`
	Task        string
}

func main() {
	initializeFlags()

	profile := utils.ReadProfileFromFile(options.Profilefile)
	history := models.ReadHistoryFromFile(options.Historyfile)

	var err error

	switch options.Task {
	case "simulate":
		simulate := task.NewSimulate(profile, options.Mode, history)
		err = simulate.Run(options.Steps, options.PeriodDepth)
		simulate.Print()
	case "likeliness":
		fmt.Printf("likeliness not implemented yet")
	default:
		fmt.Printf("task %s unknown. Select \"simulate\" or \"likeliness\" as task.", options.Task)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func initializeFlags() {
	// initialize parser for flags
	parser := flags.NewParser(&options, flags.Default)
	parser.ShortDescription = "tspredictor"
	parser.LongDescription = "Reads a TSProfile from file and runs tasks on in (Simulate or Likeliness)"
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
		fmt.Printf("No task specified. Select \"simulate\" or \"likeliness\" as task.\n")
		os.Exit(1)
	}
	options.Task = strings.ToLower(args[0])
}
