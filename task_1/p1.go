// 1. Sum and average: Write a function that takes a slice of integers and returns both the sum and
// the average.

package main

func SumAndAverage(numbers []int) (int, int) {
	if len(numbers) == 0 {
		return 0, 0
	}
	sum := 0
	for _, num := range numbers{
		sum += num 
	}
	average := sum / len(numbers)
	return sum, average
}
