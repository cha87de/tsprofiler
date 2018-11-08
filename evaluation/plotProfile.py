#!/usr/bin/python
from util import getNextState, getSimTXValue, getSimAvgValue, aggregate
import json
import matplotlib.pyplot as plt
import sys
import argparse
from util import printTSPlot, printTXPlot

def simulateTX(metric, length):
    output = []
    currentState = 0
    for x in range(length):
        txmatrix = metric["txmatrix"]
        max = metric["stats"]["max"]
        min = metric["stats"]["min"]
        stddev = metric["stats"]["stddev"]

        currentState = getNextState(txmatrix, currentState)

        simValue = getSimTXValue(txmatrix, currentState, min, max, stddev)
        output.append(simValue)

        sys.stdout.write("{}/{}\r".format(x, length))
    sys.stdout.write("\n")
    return output

def simulateAvg(metric, length):
    output = []
    currentState = 0
    for x in range(length):
        avg = metric["stats"]["avg"]
        max = metric["stats"]["max"]
        min = metric["stats"]["min"]
        stddev = metric["stats"]["stddev"]
        simValue = getSimAvgValue(avg, min, max, stddev)
        output.append(simValue)
        sys.stdout.write("{}/{}\r".format(x, length))
    sys.stdout.write("\n")        
    return output

def main(file, metricName, simlength, graphlength):
    # read profile from json file
    with open(file) as f:
        profile = json.load(f)

        # take metric with metricName
        metric = {}
        for m in profile["metrics"]:
            if m["name"] == metricName:
                metric = m
                break

        # have we found a metric? if not, return
        if not 'name' in metric:
            print("could not find metric " + metricName)
            return

        # first: print txmatrix
        printTXPlot(metric["txmatrix"])

        # second: run simulation
        print("start tx simulation")
        simulationTX = simulateTX(metric, simlength)
        print("start avg simulation")
        simulationAvg = simulateAvg(metric, simlength)

        # third: print simulated ts
        printTSPlot("tx", aggregate(simulationTX, graphlength))
        printTSPlot("avg", aggregate(simulationAvg, graphlength))


# bootstrap application: handle arguments, then call main
parser = argparse.ArgumentParser(description='Print a TSProfiler profile.')
# file, metricName, simlength, graphlength
parser.add_argument('--simlength', dest='simlength', action='store',
                    default=4000, type=int,
                    help='amount of time points to simulate')
parser.add_argument('--graphlength', dest='graphlength', action='store',
                    default=400, type=int,
                    help='amount of time points in the plotted graph. if less than simlength, values are aggregated')
parser.add_argument('profile', metavar='profile',
                    help='the file which contains the profile as json')
parser.add_argument('metric', metavar='metric',
                    help='specifies the metric name to use, which must be set in the profile')

args = parser.parse_args()
print(args)
main(args.profile, args.metric, args.simlength, args.graphlength)
