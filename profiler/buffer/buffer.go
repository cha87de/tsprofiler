package buffer

import (
	"sync"

	"github.com/cha87de/tsprofiler/models"
	"github.com/cha87de/tsprofiler/utils"
)

// NewBuffer initializes and returns a new Buffer
func NewBuffer(filterStdDevs int) Buffer {
	return Buffer{
		items:         make([]models.TSBuffer, 0),
		metricIndex:   make(map[string]int),
		access:        &sync.Mutex{},
		filterStdDevs: filterStdDevs,
	}
}

// Buffer manages access and holds TSBuffer items to buffer TSInput data
type Buffer struct {
	// cache
	items       []models.TSBuffer
	metricIndex map[string]int
	access      *sync.Mutex
	// configs
	filterStdDevs int
}

// Add adds the given tsdata item to its metric buffer
func (buffer *Buffer) Add(data models.TSInput) {
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
		utils.IsOutlier(input.Value, currentAvg, currentStddev, buffer.filterStdDevs)

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
