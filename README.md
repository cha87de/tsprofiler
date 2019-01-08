# TSProfiler

*a profiler for time series data*

[![Build Status](https://travis-ci.org/cha87de/tsprofiler.svg?branch=master)](https://travis-ci.org/cha87de/tsprofiler)
[![GoDoc](https://godoc.org/github.com/cha87de/tsprofiler/impl?status.svg)](https://godoc.org/github.com/cha87de/tsprofiler/impl)


## Purpose

*TSProfiler* provides a go implementation to convert time series stream data like monitoring data online into statistical representative profiles.

TSProfiler is integrated into the KVM monitoring tool [kvmtop](https://github.com/cha87de/kvmtop/tree/profiler).

*The output format is currently under development.* The core concept is a
transition matrix (Markov Chain) to represent the probability of a discretized
utilisation state. Example: the probability if the current cpu utilisation is in
state 2 (50% < util <= 75%) that it will change to state 3 (75% < util <= 100%)
is 26%. The matrix and a visual representation of the actual utilisation is shown below.

```
#         0       1       2       3
# 0       93%     3%      2%      2%
# 1       67%     8%      6%      19%
# 2       62%     3%      9%      26%
# 3       36%     24%     14%     26%
profile = [
    [93, 3, 2, 2],
    [67, 8, 6, 19],
    [62, 3, 9, 26],
    [36, 24, 14, 26]
]
```

## Usage Guide

*The API is still under development.* Create a new TSProfiler:

```go
profiler = impl.NewProfiler(spec.Settings{
		Name:           "profiler-hostX",
		BufferSize:     10,
		States:         4,
		OutputFreq:     time.Duration(20) * time.Second,
		OutputCallback: profileOutput,
	})

func profileOutput(data spec.TSProfile) {
  // handle profiler output
}
```

Provide metric value to profiler:

```go
metrics := make([]spec.TSDataMetric, 0)
metrics = append(metrics, spec.TSDataMetric{
		Name:  "CPU-Util",
		Value: float64(utilValue),
	})
tsdata := spec.TSData{
		Metrics: metrics,
	}
profiler.Put(tsdata)
```
