package impl

import (
	"fmt"
	"os"
	"time"

	"github.com/cha87de/tsprofiler/spec"
)

func (profiler *simpleProfiler) profileRunner() {
	for {
		start := time.Now()
		profiler.profile()
		nextRun := start.Add(time.Duration(profiler.settings.Frequency) * time.Second)
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

func (profiler *simpleProfiler) profilePrintRunner() {
	for {
		start := time.Now()
		output := profiler.printTransitMatrix()

		// write output to file
		filename := "profile_" + profiler.settings.Name + ".txt"
		//f, err := os.Create(filename)
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("error opening profile file %s: %s", filename, err)
		}
		f.WriteString(output)
		f.Close()

		nextRun := start.Add(time.Duration(1) * time.Minute)
		time.Sleep(nextRun.Sub(time.Now()))
	}
}

func (profiler *simpleProfiler) profile() {
	profiler.dataaccess.Lock()
	var data []float64
	for _, d := range profiler.data {
		data = append(data, d.Value)
	}
	profiler.data = make([]spec.TSData, 0)
	profiler.dataaccess.Unlock()

	newState := discretize(aggregate(data))
	profiler.transit(newState)
}

func (profiler *simpleProfiler) transit(state state) {
	profiler.statematrix[profiler.currentState.value][state.value]++
	// finally: update current state
	profiler.currentState = state
}

func (profiler *simpleProfiler) printTransitMatrix() string {
	output := fmt.Sprintf("\n\t")
	for i := 0; i < maxstates; i++ {
		output = output + fmt.Sprintf("%d\t", i)
	}
	output = output + fmt.Sprintf("\n")
	for i, row := range profiler.statematrix {
		output = output + fmt.Sprintf("%d\t", i)
		sum := sum(row)
		for _, v := range row {
			var frac float64
			if sum == 0 {
				frac = 0.0
			} else {
				frac = float64(v) / float64(sum) * 100
			}
			output = output + fmt.Sprintf("%.0f%%\t", frac)
		}
		output = output + fmt.Sprintf("\n")
	}
	return output
}
