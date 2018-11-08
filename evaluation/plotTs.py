#!/usr/bin/python
import numpy
import csv
from util import printTSPlot

def main():
    values = []
    reader = csv.reader(open("tsinput.csv"), delimiter=' ')
    for row in reader:
        try:
            if len(row) >= 1:
                values.append(float(row[0]))
        except ValueError:
            pass
    numpyarr = numpy.array(values)

    # print line graph
    printTSPlot("original", numpyarr)  

main()