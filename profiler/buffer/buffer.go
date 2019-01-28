package buffer

import (
	"sync"

	"github.com/cha87de/tsprofiler/api"
	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/utils"
)

// NewBuffer initializes and returns a new Buffer
func NewBuffer(filterStdDevs int, profiler api.TSProfiler) Buffer {
	return Buffer{
		profiler:      profiler,
		items:         make([]models.TSBuffer, 0),
		metricIndex:   make(map[string]int),
		access:        &sync.Mutex{},
		filterStdDevs: filterStdDevs,
	}
}

// Buffer manages access and holds TSBuffer items to buffer TSInput data
type Buffer struct {
	// upper level profiler
	profiler api.TSProfiler
	// cache
	items       []models.TSBuffer
	metricIndex map[string]int
	access      *sync.Mutex
	// configs
	filterStdDevs int
}

// Add adds the given tsdata item to its metric buffer
func (buffer *Buffer) Add(data models.TSInput) {
	currentStates := buffer.profiler.GetCurrentState()

	// for each metric ...
	for _, input := range data.Metrics {
		buffer.access.Lock()
		metric := input.Name
		index, exists := buffer.metricIndex[metric]
		if !exists {
			// create new buffer for metric
			index = len(buffer.items)
			buffer.items = append(buffer.items, models.NewTSBuffer(metric))
			buffer.metricIndex[metric] = index
		}

		// OUTLIER CHECK?
		var currentState models.TSState
		currentStateFound := false
		for _, n := range currentStates {
			if n.Metric == metric {
				currentState = n
				currentStateFound = true
			}
		}
		if currentStateFound {
			utils.IsOutlier(input.Value, currentState.Statistics.Avg, currentState.Statistics.Stddev, buffer.filterStdDevs)
		}

		// add to the metric buffer
		buffer.items[index].Append(input.Value)
		buffer.items[index].FixedMin = input.FixedMin
		buffer.items[index].FixedMax = input.FixedMax
		buffer.access.Unlock()
	}
}

// Reset clears the buffer after it returned a copy
func (buffer *Buffer) Reset() []models.TSBuffer {
	var buffers []models.TSBuffer
	buffer.access.Lock()

	// make a copy
	copy(buffers, buffer.items)

	// clear the buffer
	buffer.items = make([]models.TSBuffer, 0)
	buffer.metricIndex = make(map[string]int)

	buffer.access.Unlock()
	return buffers
}
