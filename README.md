# TSProfiler

*a profiler for time series data*

[![GoDoc](https://godoc.org/github.com/cha87de/tsprofiler/impl?status.svg)](https://godoc.org/github.com/cha87de/tsprofiler/impl)
[![GoDoc](https://godoc.org/github.com/cha87de/tsprofiler/spec?status.svg)](https://godoc.org/github.com/cha87de/tsprofiler/spec)

## Purpose

*TSProfiler* provides a go implementation to convert time series stream data like monitoring data online into statistical representative profiles.

TSProfiler is integrated into the KVM monitoring tool [kvmtop](https://github.com/cha87de/kvmtop/tree/profiler)

## Usage Guide

Create a new TSProfiler:

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

```
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
