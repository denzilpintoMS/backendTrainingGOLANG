package main


func EvenOrOdd(nums []int) (int, int) {
	evenCounts := 0
	oddCounts := 0
	for _, num := range nums {
		if num % 2 == 0 {
			evenCounts++
		}else {
			oddCounts++
		}
	}
	return evenCounts, oddCounts
}
