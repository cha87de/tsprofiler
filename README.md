# TSProfiler

*a profiler for time series data* - **breaking changes may occur, since currently under development.**

[![Build Status](https://travis-ci.org/cha87de/tsprofiler.svg?branch=master)](https://travis-ci.org/cha87de/tsprofiler)
[![GoDoc](https://godoc.org/github.com/cha87de/tsprofiler/impl?status.svg)](https://godoc.org/github.com/cha87de/tsprofiler/impl)

```
+------------+       +------------+      +---------------+
| TimeSeries | +---> | TSProfiler | +--> |  Statistical  |
|    Data    |       |            |      |    Profile    |
+------------+       +------------+      +---------------+
```

## Purpose

*TSProfiler* provides a go implementation to convert time series stream data
like monitoring data online into statistical representative profiles. TSProfiler
is integrated into the KVM monitoring tool
[kvmtop](https://github.com/cha87de/kvmtop/tree/profiler) directly, or for
distributed setups into the [DisResc Monitoring
Library](https://github.com/disresc/profiler).

The core concept bases on *Markov Chain*s to represent the probability of a
discretized utilisation states, and a *Decision Tree* to handle periodic
recurrent Markov transition matrices (the period tree). Automatic phase
detection further identifies pattern changes.

![TSProfiler Architecture](./docs/extendend-tsprofiler.svg "TSProfiler Architecture")


## Usage Guide

### Command line tool **csv2tsprofile**

The TSProfiler comes with a command line tool to read a CSV file and generate a
TSProfile. [Get the most recent stable build from
Releases.](https://github.com/cha87de/tsprofiler/releases)

```
Usage:
  csv2tsprofile [OPTIONS]

Reads time series values from a CSV file and generates a tsprofile

Application Options:
      --states=
      --buffersize=
      --history=
      --filterstddevs=
      --fixedbound
      --fixedmin=              if fixedbound is set, set the min value (default: 0)
      --fixedmax=              if fixedbound is set, set the max value (default: 100)
      --periodsize=            comma separated list of ints, specifies descrete states per period
      --phasechangelikeliness=
      --phasechangehistory=
      --output=                path to write profile to, stdout if '-' (default: -)
      --out.history=           path to write last historic values to, stdout if '-', empty to disable
      --out.phases=
      --out.periods=
      --out.states=

Help Options:
  -h, --help                   Show this help message
```

Example: `csv2tsprofile --states 4 --history 1 --filterstddevs 4 --buffersize 6 --periodsize 2,24,48 path/to/tsinput.csv`

### Command line tool **tspredictor**

The TSPredictor reads a TSProfile and the current position to provide simulation
or likeliness calculations for future next states. The mode can be either 0
(root tx), 1 (detected phases), or 2 (periods). Simulation or likeliness has to
be specified as the requested task.

```
Usage:
  tspredictor [OPTIONS]

Reads a TSProfile from file and runs tasks on in (Simulate or Likeliness)

Application Options:
      --steps=
      --mode=
      --periodDepth=
  -p, --profile=
  -h, --history=

Help Options:
  -h, --help         Show this help message
```

Example (with csv2tsprofile):

```
csv2tsprofile \
	--fixedbound \
	--fixedmin 0 \
	--fixedmax 100 \
	--states 10 \
	--buffersize 1 \
	--history 1 \
	--periodsize 16,4 \
	--phasechangelikeliness 0.50 \
	--phasechangehistory 10 \
	--out.history /tmp/history.json \
	--output /tmp/profile.json \
	--out.phases /tmp/out.phases.log \
	--out.periods /tmp/out.periods.log \
	--out.states /tmp/out.states.log \
	tsinput.csv

tspredictor \
	--profile /tmp/profile.json \
	--history /tmp/history.json \
	--steps 4 \
	--mode 0 \
	simulate		
```

### Integrate into Go Code via TSProfiler API

Create a new TSProfiler:

```go
tsprofiler := profiler.NewProfiler(models.Settings{
		Name:          "profiler-hostX",
		BufferSize:    10,
		States:        4,
		FilterStdDevs: 4,
		History:       1,
		FixBound:      false,
		PeriodSize:    []int{60,720,1440},
		// ... many more settings
		OutputFreq:     time.Duration(20) * time.Second,
		OutputCallback: profileOutput,
	})

func profileOutput(data models.TSProfile) {
  // handle profiler output via OutputFreq
}

// get profile independently of OutputFreq
profile := profiler.Get()
```

Provide metric value to profiler:

```go
metrics := make([]models.TSInputMetric, 0)
metrics = append(metrics, models.TSInputMetric{
		Name:  "CPU-Util",
		Value: float64(utilValue),
		FixedMin: options.FixedMin, // optional, required for FixBound = true
		FixedMax: options.FixedMax, // optional, required for FixBound = true
	})
tsinput := models.TSInput{
		Metrics: metrics,
	}
profiler.Put(tsinput)
```
