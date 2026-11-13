package main

import "fmt"


func TotalPerUser(userIds []int, amounts []int) map[int]int{
	totals := make(map[int]int)
	if len(userIds) != len(amounts) {
		fmt.Errorf("Invalid input: size don't match")
		return nil
	}
	for i, v := range userIds{
		totals[v] += amounts[i]
	}
	return totals 
}
