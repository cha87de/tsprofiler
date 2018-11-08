import numpy as np
import math
from random import randint
import matplotlib.pyplot as plt

def getNextState(txmatrix, currentState):
    stateProbs = txmatrix[currentState]
    validStates = [i for i in range(len(stateProbs))]
    nextState = weighted_choice(validStates, stateProbs)
    return nextState

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
    stateSize = (max-min) / states
    value = min + currentState * stateSize
    value += randint(0, stateSize) * (stddev/max) # add noise    
    return value

def getSimAvgValue(avg, min, max, stddev):
    value = avg
    value += randint(max*-1, max) * (stddev/max) # add noise
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
    plt.savefig("tsplot-" + name + ".png",  dpi=199)
    plt.clf()
    plt.close()
    plt.cla()

def printTXPlot(values):
    plt.imshow(values, cmap='Greys')
    plt.colorbar()
    plt.savefig("txplot.png")
    plt.clf()
    plt.close()
    plt.cla()

