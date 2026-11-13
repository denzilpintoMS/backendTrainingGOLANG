package main

func FilterTransactions(transactions []int) []int {
	ans := []int{}
	for _, t := range transactions {
		if t > 0 {
			ans = append(ans, t)
		}
	}
	return ans
}
