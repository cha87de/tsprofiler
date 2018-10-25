import numpy as np
import matplotlib.pyplot as plt
import math
from random import randint

# UUID: af505bce-6eb0-417e-abae-a0f181477a16
#         0       1       2       3
# 0       82%     18%     0%      0%
# 1       77%     23%     0%      0%
# 2       0%      0%      0%      0%
# 3       0%      0%      0%      0%
#
profile = [
    [82/100, 18/100, 0/100, 0/100],
    [77/100, 23/100, 0/100, 0/100],
    [0/100, 0/100, 0/100, 0/100],
    [0/100, 0/100, 0/100, 0/100]
]

# UUID: c572553e-1514-4c72-8936-4f211ccaaef3
#         0       1       2       3
# 0       25%     0%      0%      75%
# 1       0%      0%      0%      100%
# 2       0%      0%      33%     67%
# 3       0%      1%      0%      98%
profile = [
    [25/100, 0/100, 0/100, 75/100],
    [0/100, 0/100, 0/100, 100/100],
    [0/100, 0/100, 33/100, 67/100],
    [0/100, 1/100, 0/100, 98/100]
]

# UUID: 463d4924-c799-43b0-a769-e160b6e58c6c
#         0       1       2       3
# 0       93%     3%      2%      2%
# 1       67%     8%      6%      19%
# 2       62%     3%      9%      26%
# 3       36%     24%     14%     26%
profile = [
    [93/100, 3/100, 2/100, 2/100],
    [67/100, 8/100, 6/100, 19/100],
    [62/100, 3/100, 9/100, 26/100],
    [36/100, 24/100, 14/100, 26/100]
]

def getNextState(currentState):
    global profile

    probabilities = profile[currentState]

    states = [0, 1, 2, 3]
    nextState = weighted_choice(states, probabilities)

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
        if x < partition[i]:
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

def printPlot(values):
    plt.plot(values)
    plt.show()

def aggregate(values, max):
    """ takes input array and aggregates usind mean
    the output array to have a maximum of MAX elements
    """
    inarr = np.array(values)
    window_sz = math.floor(len(inarr) / max)
    outarr = inarr.reshape(-1,window_sz).mean(1) 
    return outarr

def main():
    global profile

    output = []
    currentState = 0
    for x in range(2000):
        currentState = getNextState(currentState)
        value = currentState * 25
        value += randint(0, 24)
        output.append(value)

    maxPlotTs = len(output) # no aggregation
    # maxPlotTs = 400 # with aggregation if < len(output)
    printPlot(aggregate(output, maxPlotTs))

main()
