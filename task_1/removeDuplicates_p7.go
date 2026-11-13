package main

// remove dups 

func removeDuplicates(nums []int) []int  {
	seen := make(map[int]int)
	ans := []int{}
	for _, n := range nums {
		_, ok := seen[n] 
		if ok {
			seen[n]++
		}else{
			seen[n] = 1
		}
	}
	for k , _:= range seen {
	ans = append(ans, k)
	}
	return ans
}
