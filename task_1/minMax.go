package main

func findMinMax(nums []int) (int, int){
	if len(nums) == 0 {
		return 0, 0
	}
	minn := nums[0]
	maxx := nums[0]
	for _, n := range nums {
		if n > maxx{
			maxx = n
		}
		if n < minn {
			minn = n
		}
}
		return minn, maxx
}
