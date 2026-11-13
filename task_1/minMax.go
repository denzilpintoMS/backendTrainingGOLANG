package main

func findMinMax(nums []int) (int, int){
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
