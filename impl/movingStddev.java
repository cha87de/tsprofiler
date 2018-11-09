/**
 * standardDeviation() - designed to calculate the standard deviation of a data set incrementally by taking the last entered value and the previous sum of differences to the mean recorded.
 * (i.e; upon adding a value to the data set this function should immediately be called)
 * 
 * NOTE: do not call this function if the data set size it less than 2 since standard deviation cannot be calculated on a single value
 * NOTE: sum_avg, sum_sd and avg are all static variables
 * NOTE: on attempting to use this on another set following previous use, the static values will have to be reset**
 * 
 * @param vector - List<Double> - data with only one additional value from previous method call
 * @return updated value for the Standard deviation
 */
public static double standardDeviation(List<Double> vector)
{   
    double N = (double) vector.size();                  //size of the data set
    double oldavg = avg;                                //save the old average
    avg = updateAverage(vector);                        //update the new average

    if(N==2.0)                                          //if there are only two, we calculate the standard deviation using the standard formula 
    {                                                               
        for(double d:vector)                            //cycle through the 2 elements of the data set - there is no need to use a loop here, the set is quite small to just do manually
        {
            sum_sd += Math.pow((Math.abs(d)-avg), 2);   //sum the following according to the formula
        }
    }
    else if(N>2)                                        //once we have calculated the base sum_std  
    {   
        double newel = (vector.get(vector.size()-1));   //get the latest addition to the data set

        sum_sd = sum_sd + (newel - oldavg)*(newel-avg); //https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Online_algorithm
    }
    return Math.sqrt((sum_sd)/(N));                     //N or N-1 depends on your choice of sample of population standard deviation

}

/**
 * simplistic method for incrementally calculating the mean of a data set
 * 
 * @param vector - List<Double> - data with only one additional value from previous method call
 * @return updated value for the mean of the given data set
 */
public static double updateAverage(List<Double> vector)
{
    if(vector.size()==2){
        sum_avg = vector.get(vector.size()-1) + vector.get(vector.size()-2);
    }
    else{
        sum_avg += vector.get(vector.size()-1);
    }


    return sum_avg/(double)vector.size();

}