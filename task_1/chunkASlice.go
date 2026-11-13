package main

func ChunkSlice(slice []int, n int) ([][]int){
	var ans [][]int
	if n <= 0{
		return [][]int{}
	}
	for i:=0;i<len(slice);i+=n{
		end := i + n
		if end > len(slice){
			end = len(slice)
		}
		ans = append(ans, slice[i:end])
	}
	return ans
}
