import numpy as np
import math
from random import randint
import matplotlib.pyplot as plt
import csv
import sys

def getNextState(txmatrix, currentState):
    stateProbs = txmatrix[currentState]["nextProbs"]
    validStates = [i for i in range(len(stateProbs))]
    nextState = weighted_choice(validStates, stateProbs)
    parts = currentState.split("-")
    del parts[0]
    parts.append(str(nextState))
    return "-".join(parts)

def find_interval(x, partition):
    """ find_interval -> i
        partition is a sequence of numerical values
        x is a numerical value
        The return value "i" will be the index for which applies
        partition[i] < x < partition[i+1], if such an index exists.
        -1 otherwise
    """
    for i in range(0, len(partition)):
        if x < (partition[i] / 100):
            return i-1
    return -1

def weighted_choice(sequence, weights):
    """ 
    weighted_choice selects a random element of 
    the sequence according to the list of weights
    """
    x = np.random.random()
    cum_weights = [0] + list(np.cumsum(weights))
    index = find_interval(x, cum_weights)
    return sequence[index]    

def getSimTXValue(txmatrix, currentState, min, max, stddev):
    states = len(txmatrix[currentState])
    stateSize = round((max-min) / states)
    parts = currentState.split("-")
    stateValue = int(parts[len(parts)-1])
    value = min + stateValue * stateSize
    value += randint(0, stateSize) * (stddev/max) # add noise    
    return value

def getSimAvgValue(avg, min, max, stddev):
    value = avg
    value += randint(round(max*-1), round(max)) * (stddev/max) # add noise
    return value

def aggregate(values, max):
    """ takes input array and aggregates usind mean
    the output array to have a maximum of MAX elements
    """
    inarr = np.array(values)
    window_sz = math.floor(len(inarr) / max)
    outarr = inarr.reshape(-1,window_sz).mean(1) 
    return outarr

def printTSPlot(name, values):
    plt.figure(figsize=(16,9))
    plt.plot(values, linewidth=0.8)
    plt.savefig("results/tsplot-" + name + ".png",  dpi=199)
    plt.clf()
    plt.close()
    plt.cla()

def printTXPlot(values):
    plt.imshow(values, cmap='Greys')
    plt.colorbar()
    plt.savefig("results/txplot.png")
    plt.clf()
    plt.close()
    plt.cla()


def readTsValues(filename):
    values = []
    reader = csv.reader(open(filename), delimiter=' ')
    for row in reader:
        try:
            if len(row) >= 1:
                values.append(float(row[0]))
        except ValueError:
            pass
    return values

def simulateTX(metric, length):
    output = []
    currentState = "0" # TODO get history count
    txmatrix = metric["txmatrix"]
    max = metric["stats"]["max"]
    min = metric["stats"]["min"]
    stddev = metric["stats"]["stddev"]    
    for x in range(length):
        currentState = getNextState(txmatrix, currentState)
        if not currentState in txmatrix:
            print("no exact state found")
            currentStateMin = currentState
            while len(currentStateMin) > 0:
                parts = currentStateMin.split("-")
                if len(parts) <= 0:
                    break
                del parts[0]
                currentStateMin = "-".join(parts)
                if currentStateMin in txmatrix:
                    # found a match
                    currentState = currentStateMin
                    break

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
